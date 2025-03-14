package main

import (
	"flag"
	"fmt"
	"github.com/rs/zerolog/log"
	"net"
	"os"
	"slices"
)

type GetProfileCallback = func(string) (*Profile, bool)

type ArrClient interface {
	ProcessWebhook(payload []byte) error
}

func isSubset[T comparable](a, b []T) bool {
	for _, item := range b {
		if !slices.Contains(a, item) {
			return false
		}
	}
	return true
}

func parseConfigType() string {
	confType := flag.String("prof", "yaml", "Set profile file type (available values: json, yaml, toml)")
	flag.Parse()
	validValues := map[string]bool{
		"json": true,
		"yaml": true,
		"toml": true,
	}

	if !validValues[*confType] {
		fmt.Printf("Error: Invalid value for --conf flag: %s\n", *confType)
		fmt.Println("Available values: json, yaml, toml")
		flag.Usage()
		os.Exit(1)
	}

	fmt.Printf("Profile file type is set to: %s\n", *confType)
	return *confType
}

func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to connect to outbound IP")
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Fatal().Err(err).Msg("error occurred while closing connection")
		}
	}(conn)

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}
