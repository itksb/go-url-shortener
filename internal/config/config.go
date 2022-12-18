package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

type (
	// SessionConfig - Application session configuration. All fields required
	SessionConfig struct {
		HashKey  string
		BlockKey string
	}

	// Config - Application configuration
	Config struct {
		AppPort         int
		AppHost         string
		ShortBaseURL    string
		FileStoragePath string
		SessionConfig   SessionConfig
		Dsn             string
	}
)

// NewConfig - Constructor
func NewConfig() (Config, error) {
	// We can use environment parser here
	cfg := Config{
		AppPort:         8080,
		ShortBaseURL:    "http://localhost:8080",
		AppHost:         "localhost",
		FileStoragePath: "",
		SessionConfig: SessionConfig{
			HashKey:  "1234567890",
			BlockKey: "0123456701234567" + "0123456701234567",
		},
		Dsn: "",
	}
	return cfg, nil
}

// UseOsEnv - apply environment variables
func (cfg *Config) UseOsEnv() {
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

	sessionHashKey, ok := os.LookupEnv("SESSION_HASHKEY")
	if ok {
		cfg.SessionConfig.HashKey = sessionHashKey
	}

	sessionBlockKey, ok := os.LookupEnv("SESSION_BLOCKKEY")
	if ok {
		cfg.SessionConfig.BlockKey = sessionBlockKey
	}

	dsn, ok := os.LookupEnv("DATABASE_DSN")
	if ok {
		cfg.Dsn = dsn
	}

}

// UseFlags - scan flags
func (cfg *Config) UseFlags() {
	appHost := flag.String("a", cfg.AppHost, "SERVER_ADDRESS")
	shortBaseURL := flag.String("b", cfg.ShortBaseURL, "BASE_URL")
	fileStoragePath := flag.String("f", cfg.FileStoragePath, "FILE_STORAGE_PATH")
	dsn := flag.String("d", cfg.Dsn, "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable")
	flag.Parse()

	addr := strings.SplitN(*appHost, ":", 2)
	if len(addr) == 2 {
		cfg.AppHost = addr[0]
		intValue := 8080
		_, err := fmt.Sscan(addr[1], &intValue)
		if err != nil {
			log.Panic("PORT value is invalid")
		}
		cfg.AppPort = intValue

	} else {
		cfg.AppHost = *appHost
	}

	cfg.ShortBaseURL = *shortBaseURL
	cfg.FileStoragePath = *fileStoragePath
	cfg.Dsn = *dsn
}
