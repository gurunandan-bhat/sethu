package main

import (
	"fmt"
	"log"
	"net/http"
	"sethupay/lib/config"
	"sethupay/lib/service"
)

func main() {

	cfg, err := config.Configuration()
	if err != nil {
		log.Fatalf("Error reading application configuration: %s", err)
	}

	service, err := service.NewService(cfg)
	if err != nil {
		log.Fatalf("Error creating new service: %s", err)
	}

	httpServer := &http.Server{
		Addr:    fmt.Sprintf("localhost:%d", cfg.AppPort),
		Handler: service.Muxer,
	}

	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatalf("Error running http server: %s", err)
	}
}
