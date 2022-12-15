package main

import (
	"github.com/itksb/go-url-shortener/internal/app"
	"github.com/itksb/go-url-shortener/internal/config"
	"log"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}
	cfg.UseOsEnv()
	cfg.UseFlags()

	application, err := app.NewApp(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer application.Close()
	log.Fatal(application.Run())
}
