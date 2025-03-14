package main

type InstanceType = string

const (
	SONARR InstanceType = "sonarr"
	RADARR InstanceType = "radarr"
)

type Profile struct {
	RequiredLanguagesAudio []string `json:"required_languages_audio"`
	RequiredLanguagesSubs  []string `json:"required_languages_sub"`
}

type ArrInstance struct {
	InstType    InstanceType        `json:"inst_type"`
	BasePath    string              `json:"base_path"`
	ApiKey      string              `json:"api_key"`
	LanguageMap map[string]*Profile `json:"language_map"`
	arrClient   ArrClient
}

// InitClient sets up a ArrClient instance based on the type of inst
// no action is taken if client is already initialized
func (ar *ArrInstance) InitClient() {
	if ar.arrClient == nil {
		switch ar.InstType {
		case SONARR:
			ar.arrClient = NewSonarr(ar.BasePath, ar.ApiKey, func(s string) (*Profile, bool) {
				val, ok := ar.LanguageMap[s]
				return val, ok
			})
		case RADARR:
			ar.arrClient = NewRadarr()
		}
	}
}
