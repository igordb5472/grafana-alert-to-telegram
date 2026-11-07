package test

import (
	"grafana-alert-to-telegram/src/config"
	"log"
	"testing"

	"github.com/joho/godotenv"
)

func TestInit(t *testing.T) {
	if err := godotenv.Load("test.env"); err != nil {
		log.Fatal(err)
	}

	if err := config.LoadBots(); err != nil {
		log.Fatal(err)
	}
}
