package main

import (
	"grafana-alert-to-telegram/src/handler"
	"grafana-alert-to-telegram/src/service"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("main: %v", err)
	}
	log.Println("environment variables loaded successfully")

	if err := service.InjectAuthCreds(); err != nil {
		log.Fatalf("main: %v", err)
	}
	log.Println("authorization credentials injected successfully")

	var err error
	service.Bots, err = service.GetBots()
	if err != nil {
		log.Fatalf("main: %v", err)
	}
	log.Println("settings parsed successfully")

	if err := service.ScheduleAllDutyNotifs(); err != nil {
		log.Fatalf("main: %v", err)
	}
	log.Println("all duty notifs scheduled successfully")

	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	http.HandleFunc("/webhooks/grafana", handler.HandleAlert)

	log.Println("initialization completed successfully")

	log.Println("starting HTTP-server")

	log.Fatal(server.ListenAndServe())
}
