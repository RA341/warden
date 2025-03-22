package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"path/filepath"
	"resty.dev/v3"
	"strconv"
)

// SonarWebhookPayload represents the structure of the incoming webhook JSON
type SonarWebhookPayload struct {
	Series struct {
		Path             string   `json:"path"`
		Tags             []string `json:"tags"`
		OriginalLanguage struct {
			Name string `json:"name"`
		} `json:"originalLanguage"`
	} `json:"series"`
	Episodes []struct {
		Id int `json:"id"`
	} `json:"episodes"`
	EpisodeFile struct {
		ID        int64 `json:"id"`
		MediaInfo struct {
			AudioLanguages []string `json:"audioLanguages"`
			Subtitles      []string `json:"subtitles"`
		} `json:"mediaInfo"`
	} `json:"episodeFile"`
}

type SonarMediaInfo struct {
	EpisodeID        int
	EpisodeFileID    string
	MediaPath        string
	Tags             []string
	OriginalLanguage string
	Subtitles        []string
	Audios           []string
}

type SonarrInst struct {
	client     *resty.Client
	getProfile GetProfileCallback
}

func NewSonarr(baseUrl, apiKey string, callback GetProfileCallback) *SonarrInst {
	return &SonarrInst{
		client: resty.New().
			SetBaseURL(baseUrl).
			SetHeader("X-Api-Key", apiKey).
			SetDebug(false),
		getProfile: callback,
	}
}

// NewSonarrWithEmptyCallback used for tests, callback always returns false
func NewSonarrWithEmptyCallback(baseUrl, apiKey string) *SonarrInst {
	return NewSonarr(
		baseUrl,
		apiKey,
		func(s string) (*Profile, bool) {
			return nil, false
		},
	)
}

func (s *SonarrInst) ProcessWebhook(jsonData []byte) error {
	info, err := s.ParseJson(jsonData)
	if err != nil {
		return err
	}
	s.RunCheck(info)
	return nil
}

func (s *SonarrInst) RunCheck(info *SonarMediaInfo) {
	prof := s.matchProfile(info)
	if prof == nil {
		log.Warn().
			Interface("tags", info.Tags).
			Str("MediaPath", info.MediaPath).
			Msgf("No profile found, checked tags and root folder")
		return
	}

	if !isSubset(info.Audios, prof.RequiredLanguagesAudio) {
		log.Info().Msgf("Found missing audio languages, \nneed: %v \ngot:%v", prof.RequiredLanguagesAudio, info.Audios)
		s.DeleteAndReMonitor(info)
		return
	}
	if !isSubset(info.Subtitles, prof.RequiredLanguagesSubs) {
		log.Info().Msgf("Found missing subtitles languages, \nneed: %v \ngot: %v", prof.RequiredLanguagesSubs, info.Subtitles)
		s.DeleteAndReMonitor(info)
		return
	}

	log.Debug().Msg("All required languages found")
}

func (s *SonarrInst) ParseJson(jsonData []byte) (*SonarMediaInfo, error) {
	var payload SonarWebhookPayload
	// Unmarshal directly into our structured type
	if err := json.Unmarshal(jsonData, &payload); err != nil {
		return nil, err
	}

	// this only checks for the first episodes, so maybe bulk imports will not work
	if len(payload.Episodes) == 0 || payload.Episodes[0].Id == 0 {
		return nil, errors.New("missing episode id")
	}

	// Validate required fields are present
	if payload.Series.Path == "" {
		return nil, errors.New("missing series path")
	}

	if payload.Series.OriginalLanguage.Name == "" {
		return nil, errors.New("missing original language name")
	}

	if payload.EpisodeFile.ID == 0 {
		return nil, errors.New("missing episodeFile id")
	}

	// shouldn't check for this probably
	//if len(payload.EpisodeFile.MediaInfo.AudioLanguages) == 0 {
	//	return nil, errors.New("missing audio languages")
	//}
	//if len(payload.EpisodeFile.MediaInfo.Subtitles) == 0 {
	//	return nil, errors.New("missing subtitles")
	//}
	//if len(payload.Series.Tags) == 0 {
	//	return nil, errors.New("missing tags")
	//}

	basePath := filepath.Dir(payload.Series.Path)
	basePath = filepath.ToSlash(basePath)

	return &SonarMediaInfo{
		EpisodeID:        payload.Episodes[0].Id,
		EpisodeFileID:    strconv.FormatInt(payload.EpisodeFile.ID, 10),
		MediaPath:        basePath,
		Tags:             payload.Series.Tags,
		OriginalLanguage: payload.Series.OriginalLanguage.Name,
		Subtitles:        payload.EpisodeFile.MediaInfo.Subtitles,
		Audios:           payload.EpisodeFile.MediaInfo.AudioLanguages,
	}, nil
}

func (s *SonarrInst) DeleteAndReMonitor(info *SonarMediaInfo) {
	log.Info().Msgf("Deleting file and remonitoring")

	err := s.deleteEpisode(info.EpisodeFileID)
	if err != nil {
		log.Error().Err(err).Msg(" failed to delete episode")
		return
	}

	err = s.monitorEpisode([]int{info.EpisodeID})
	if err != nil {
		log.Error().Err(err).Msg(" failed to re-monitor episode")
		return
	}

	err = s.SearchEpisodes(info.EpisodeID)
	if err != nil {
		log.Error().Err(err).Msg("failed to search episode")
		return
	}
}

// SearchEpisodes searches for episodes based on series ID and season number.
func (s *SonarrInst) SearchEpisodes(epID int) error {
	resp, err := s.client.R().
		SetBody(map[string]interface{}{
			"name":       "EpisodeSearch",
			"episodeIds": []int{epID},
		}).
		Post("/api/v3/command")
	if err != nil {
		return fmt.Errorf("error performing request: %w", err)
	}

	if resp.IsSuccess() {
		return fmt.Errorf("GET request failed with status code %d: %s", resp.StatusCode(), resp.String())
	}

	return nil
}

func (s *SonarrInst) matchProfile(info *SonarMediaInfo) *Profile {
	for _, tag := range info.Tags {
		prof, ok := s.getProfile(tag)
		if ok {
			return prof
		}
	}
	// if no tag was matched use the media path
	prof, ok := s.getProfile(info.MediaPath)
	if ok {
		return prof
	}
	return nil
}

func (s *SonarrInst) deleteEpisode(episodeID string) error {
	res, err := s.client.R().Delete("/api/v3/episodefile/" + episodeID)
	if err != nil {
		return err
	}
	if res.IsError() {
		return fmt.Errorf("failed to delete episode "+episodeID+"\nReason: %s", res.String())
	}

	return nil
}

func (s *SonarrInst) monitorEpisode(episodeIds []int) error {
	bodyMap := map[string]interface{}{
		"episodeIds": episodeIds,
		"monitored":  true,
	}
	res, err := s.client.R().
		SetBody(bodyMap).
		Put("/api/v3/episode/monitor")
	if err != nil {
		return err
	}
	if res.IsError() {
		return fmt.Errorf("failed to re-monitor episode %s", res.String())
	}

	return nil
}
