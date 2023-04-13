package main

import (
	"fmt"
	"github.com/itksb/go-url-shortener/internal/app"
	"github.com/itksb/go-url-shortener/internal/config"
	"log"
)

// go run -ldflags "-X main.Version=v1.1.1 \
// -X 'main.buildVersion=1.1.1'" ./cmd/shortener/main.go
var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {

	fmt.Printf(
		"Build version: %s\nBuild date: %s\nBuild commit: %s\n",
		buildVersion,
		buildDate,
		buildCommit,
	)

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
