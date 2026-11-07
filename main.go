package main

import (
	"grafana-alert-to-telegram/src/config"
	"grafana-alert-to-telegram/src/handler"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("main: %v", err)
	}

	if err := config.LoadBots(); err != nil {
		log.Fatalf("main: %v", err)
	}

	log.Println("Initialization stage successful")

	http.HandleFunc("/", handler.HandleAlert)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
