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

	app, err := app.NewApp(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer app.Close()
	log.Fatal(app.Run())
}
