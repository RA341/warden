package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSonarr_ParseWebhook(t *testing.T) {
	testPayload := `{
  "series": {
    "path": "/media/anime/I'm Getting Married to a Girl I Hate in My Class",
    "tags": [],
    "originalLanguage": {
      "id": 8,
      "name": "Japanese"
    }
  },
  "episodes": [
    {
      "id": 13947
    }
  ],
  "episodeFile": {
    "id": 11729,
    "mediaInfo": {
      "audioLanguages": [
        "jpn"
      ],
      "subtitles": [
        "eng",
        "ara",
        "ger",
        "spa",
        "fre",
        "ita",
        "por",
        "rus"
      ]
    }
  }
}`

	cli := NewSonarrWithEmptyCallback("http://localhost:8080", "sdsd")
	webhook, err := cli.ParseJson([]byte(testPayload))
	if err != nil {
		t.Fatalf("ParseJson failed: %v", err)
		return
	}

	assert.Equal(t, webhook.EpisodeID, 13947)
	assert.Equal(t, webhook.Audios, []string{"jpn"})
	assert.Equal(t, webhook.Subtitles, []string{
		"eng",
		"ara",
		"ger",
		"spa",
		"fre",
		"ita",
		"por",
		"rus",
	})
	assert.Equal(t, webhook.OriginalLanguage, "Japanese")
	assert.Equal(t, webhook.MediaPath, "/media/anime")
	assert.Equal(t, webhook.EpisodeFileID, "11729")
	assert.Equal(t, webhook.Tags, []string{})
}

func TestDelete_Remonitor(t *testing.T) {
	cli := NewSonarrWithEmptyCallback("https://sonar.dumbapps.org", "0d79c87bb0fc4cdd9039d2266519cde3")
	err := cli.deleteEpisode("11593")
	if err != nil {
		t.Fatalf("deleteEpisode failed: %v", err)
		return
	}

	err = cli.monitorEpisode([]int{13947})
	if err != nil {
		t.Fatalf("Remonitor failed: %v", err)
		return
	}
}

//func(s string) (*Profile, bool) {
//	return &Profile{
//		RequiredLanguagesAudio:   []string{"en"},
//		DisallowedLanguages: []string{},
//	}, true
//}
