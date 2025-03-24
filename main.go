package main

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
)

const (
	targetHeader = "warden-key"
)

func main() {
	printInfo()
	log.Logger = ConsoleLogger()
	fileType := parseConfigType()
	pm := NewProfileManager(fileType)

	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", handlePayload(pm))

	port := "8080"
	addr := fmt.Sprintf(":%s", port)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	val := GetOutboundIP()
	log.Info().Msgf("use http://%s:%s/webhook to send webhooks", val, port)

	if err := server.ListenAndServe(); err != nil {
		log.Fatal().Err(err).Msg("unable to start server")
	}
}

func handlePayload(pm *ProfileManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		headerValue := r.Header.Get(targetHeader)
		if headerValue == "" {
			log.Error().Msg("Missing header: " + targetHeader)
			http.Error(w, "Missing header: "+targetHeader, http.StatusBadRequest)
			return
		}

		inst, ok := pm.GetProfile(headerValue)
		if !ok {
			log.Error().Msgf("No profile found for %s", headerValue)
			http.Error(w, "no associated instance was found for key "+headerValue, http.StatusBadRequest)
			return
		}

		payload, err := readToBytes(r.Body)
		if err != nil {
			log.Error().Err(err).Msg("Error reading payload")
			http.Error(w, "unable to read payload: "+err.Error(), http.StatusInternalServerError)
			return
		}

		go func() {
			if inst.arrClient == nil {
				log.Error().Msg("client was not initialized")
				return
			}
			err := inst.arrClient.ProcessWebhook(payload)
			if err != nil {
				log.Error().Err(err).Msgf("Error processing webhook for %s", headerValue)
			}
		}()

		w.WriteHeader(http.StatusOK)
	}
}

func readToBytes(reader io.ReadCloser) ([]byte, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	// Make sure to close the reader when done
	defer func(reader io.ReadCloser) {
		err := reader.Close()
		if err != nil {
			log.Error().Err(err).Msg("an error occurred while to closing reader")
		}
	}(reader)

	return data, nil
}
