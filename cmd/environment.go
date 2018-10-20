package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
)

func mustGetEnv(key string) string {
	ret, ok := os.LookupEnv(key)
	if !ok {
		log.Fatalf("Missing environment variable %s", key)
	}
	return ret
}
