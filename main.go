package main

import (
	"discord-bot/app"
	"discord-bot/health"
	"log"
	"os"
)

func main() {
	// Railway 헬스체크를 위한 HTTP 서버 시작
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	health.StartHealthServer(port)

	application, err := app.New()
	if err != nil {
		log.Fatal(err)
	}

	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
