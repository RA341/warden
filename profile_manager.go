package main

import (
	"errors"
	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type ProfileManager struct {
	profileMap *Map[string, *ArrInstance]
	v          *viper.Viper
}

func NewProfileManager(profileFileType string) *ProfileManager {
	v := createViperInstance(profileFileType)
	profs := loadProfiles(v)
	profMan := &ProfileManager{
		profileMap: profs,
		v:          v,
	}

	v.OnConfigChange(func(e fsnotify.Event) {
		log.Info().Msgf("config file changed, refershing profiles: %s", e.Name)
		profMan.ReloadProfiles()
	})
	log.Info().Msgf("watching: %s for changes", v.ConfigFileUsed())
	v.WatchConfig()

	return profMan
}

func (pm *ProfileManager) UpsertProfile(key string, profile Profile) {
	pm.v.Set(key, profile)
	pm.WriteAndSave()
}

func (pm *ProfileManager) GetProfile(key string) (*ArrInstance, bool) {
	return pm.profileMap.Load(key)
}

func (pm *ProfileManager) WriteAndSave() {
	err := pm.v.WriteConfig()
	if err != nil {
		log.Error().Err(err).Msg("Unable to write to config")
		return
	}
	return
}

func (pm *ProfileManager) ReloadProfiles() {
	log.Debug().Msg("Reloading profiles")
	pm.profileMap.Clear()
	pm.profileMap = loadProfiles(pm.v)
}

func createViperInstance(fType string) *viper.Viper {
	v := viper.New()
	v.SetConfigName("profiles")
	v.SetConfigType(fType)
	v.AddConfigPath(".") // Path to look for the config file

	err := v.ReadInConfig()
	var configFileNotFoundError viper.ConfigFileNotFoundError
	if errors.As(err, &configFileNotFoundError) {
		log.Warn().Msg("no profile file not found, creating with default values")
		createExample(v)

		err = v.SafeWriteConfig()
		if err != nil {
			log.Error().Err(err).Msg("profile.toml could not be created")
		}
	}

	return v
}

func loadProfiles(v *viper.Viper) *Map[string, *ArrInstance] {
	instanceMap := Map[string, *ArrInstance]{}
	// Get all top-level keys (profile nicknames)
	profileNames := v.AllSettings()
	// Unmarshal each profile
	for nickname := range profileNames {
		var instance ArrInstance
		err := v.UnmarshalKey(nickname, &instance)
		if err != nil {
			log.Warn().Msgf("Error unmarshaling profile %s: %s", nickname, err)
			continue
		}
		instanceMap.Store(nickname, &instance)
		instance.InitClient()
		log.Info().Interface("inst", instance).Msgf("Loaded instance %s", nickname)
	}

	if instanceMap.Length() == uint(0) {
		log.Warn().Msg("Loaded 0 instances, please add a instance")
	}

	return &instanceMap
}

func createExample(v *viper.Viper) {
	nick := "some-meaningful-nickname"
	inst := ArrInstance{
		InstType: SONARR,
		BasePath: "https://sonarr.example.com",
		ApiKey:   "your_api_key_here",
		LanguageMap: map[string]*Profile{
			"/media/shows": {
				RequiredLanguagesAudio: []string{"en", "fr"},
				RequiredLanguagesSubs:  []string{"en", "kr"},
			},
			"/media/kdramas": {
				RequiredLanguagesAudio: []string{"en", "kr"},
				RequiredLanguagesSubs:  []string{"en", "kr"},
			},
		},
	}
	v.SetDefault(nick, inst)
}
