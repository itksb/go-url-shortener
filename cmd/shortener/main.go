package main

import (
	"fmt"
	"github.com/itksb/go-url-shortener/internal/app"
	"github.com/itksb/go-url-shortener/internal/config"
	"log"
	"os"
	"strings"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}
	useOsEnv(&cfg)

	app, err := app.NewApp(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer app.Close()
	log.Fatal(app.Run())
}

func useOsEnv(cfg *config.Config) {
	host, ok := os.LookupEnv("SERVER_ADDRESS")
	if ok {
		addr := strings.SplitN(host, ":", 2)
		if len(addr) == 2 {
			cfg.AppHost = addr[0]
			intValue := 8080
			_, err := fmt.Sscan(addr[1], &intValue)
			if err != nil {
				log.Panic("PORT value is invalid")
			}
			if ok {
				cfg.AppPort = intValue
			}
		} else {
			cfg.AppHost = host
		}
	}

	baseURL, ok := os.LookupEnv("BASE_URL")
	if ok {
		cfg.ShortBaseURL = baseURL
	}

	appPortStr, ok := os.LookupEnv("PORT")
	if ok {
		intValue := 8080
		_, err := fmt.Sscan(appPortStr, &intValue)
		if err != nil {
			log.Panic("PORT value is invalid")
		}
		if ok {
			cfg.AppPort = intValue
		}
	}

	fileStoragePath, ok := os.LookupEnv("FILE_STORAGE_PATH")
	if ok {
		cfg.FileStoragePath = fileStoragePath
	}

}
