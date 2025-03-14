package main

import "github.com/rs/zerolog/log"

type RadarrInst struct {
}

func NewRadarr() *RadarrInst {
	return &RadarrInst{}
}

func (r *RadarrInst) ProcessWebhook(payload []byte) error {
	//TODO implement me
	log.Error().Msg("Unimplemented")
	return nil
}
