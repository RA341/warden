package main

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
)

type contextKey string

const (
	ProfileManagerKey contextKey = "profileManager"
	targetHeader                 = "warden-key"
)

func main() {
	log.Logger = ConsoleLogger()
	fileType := parseConfigType()
	pm := NewProfileManager(fileType)

	e := echo.New()
	e.HideBanner = true
	//e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(string(ProfileManagerKey), pm)
			return next(c)
		}
	})
	e.POST("/webhook", handlePayload)

	val := GetOutboundIP()
	port := "8080"
	log.Info().Msgf("use http://%s:%s/webhook to send webhooks", val, port)

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", port)))
}

// handlePayload processes incoming requests to the /payload endpoint
func handlePayload(c echo.Context) error {
	headers := c.Request().Header
	headerValue := headers.Get(targetHeader)
	if headerValue == "" {
		log.Error().Msg("Missing header: " + targetHeader)
		return c.String(http.StatusBadRequest, "Missing header: "+targetHeader)
	}

	tmp := c.Get(string(ProfileManagerKey))
	if tmp == nil {
		log.Error().Msg("No profile manager found in context")
		return echo.NewHTTPError(http.StatusInternalServerError, "No profile manager found in context")
	}

	pm := tmp.(*ProfileManager)
	inst, ok := pm.GetProfile(headerValue)
	if !ok {
		log.Error().Msgf("No profile found for %s", headerValue)
		return echo.NewHTTPError(http.StatusBadRequest, "no associated instance was found for key "+headerValue)
	}

	payload, err := readToBytes(c.Request().Body)
	if err != nil {
		log.Error().Err(err).Msg("Error reading payload")
		return echo.NewHTTPError(http.StatusInternalServerError, "unable to read payload"+err.Error())
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

	return c.NoContent(http.StatusOK)
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
