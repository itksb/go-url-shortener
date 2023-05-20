// Package config application configuration
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

// SessionConfig application session configuration. All fields required
type SessionConfig struct {
	HashKey  string // secret key, used for hashing algo
	BlockKey string // secret block key
}

// Config application configuration structure
type Config struct {
	AppPort         int           `json:"-"`                 // application port
	AppHost         string        `json:"server_address"`    // application host
	ShortBaseURL    string        `json:"base_url"`          // short base url
	FileStoragePath string        `json:"file_storage_path"` // file storage path
	SessionConfig   SessionConfig `json:"-"`                 // session configuration
	Dsn             string        `json:"database_dsn"`      // data source name
	Debug           bool          `json:"-"`                 // is debug mode
	EnableHTTPS     bool          `json:"enable_https"`      // enable https
	Config          string        `json:"-"`                 // config file path
	TrustedSubnet   string        `json:"trusted_subnet"`    // CIDR, e.g.: 127.0.0.1/24
	GRPCAddr        string        `json:"grpc_address"`      // gRPC server address
}

// NewConfig  configuration constructor
func NewConfig() (Config, error) {
	// We can use environment parser here
	cfg := Config{
		AppPort:         8080,
		AppHost:         "localhost",
		ShortBaseURL:    "http://localhost:8080",
		FileStoragePath: "",
		SessionConfig: SessionConfig{
			HashKey:  "1234567890",
			BlockKey: "0123456701234567" + "0123456701234567",
		},
		Dsn:           "",
		Debug:         false,
		EnableHTTPS:   false,
		Config:        "",
		TrustedSubnet: "",
		GRPCAddr:      ":3200",
	}
	return cfg, nil
}

// UseOsEnv applies environment variables
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

	if debug, ok := os.LookupEnv("ENV"); ok {
		switch strings.ToLower(debug) {
		case "debug":
			cfg.Debug = true
		case "prod":
			cfg.Debug = false
		}
	}

	_, ok = os.LookupEnv("ENABLE_HTTPS")
	if ok {
		cfg.EnableHTTPS = true
	}

	configFile, ok := os.LookupEnv("CONFIG")
	if ok {
		cfg.Config = configFile
	}

	trustedSubnet, ok := os.LookupEnv("TRUSTED_SUBNET")
	if ok {
		cfg.TrustedSubnet = trustedSubnet
	}

	grpcAddr, ok := os.LookupEnv("GRPC_ADDRESS")
	if ok {
		cfg.GRPCAddr = grpcAddr
	}

}

// UseFlags applies run flags
func (cfg *Config) UseFlags() {
	appHost := flag.String("a", cfg.AppHost, "SERVER_ADDRESS")
	grpcAddr := flag.String("g", cfg.GRPCAddr, "GRPC_ADDRESS")
	shortBaseURL := flag.String("b", cfg.ShortBaseURL, "BASE_URL")
	fileStoragePath := flag.String("f", cfg.FileStoragePath, "FILE_STORAGE_PATH")
	curDebug := func() string {
		if cfg.Debug {
			return "debug"
		}
		return "prod"
	}()
	debugStr := flag.String("e", curDebug, "prod|debug")
	dsn := flag.String("d", cfg.Dsn, "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable")
	flag.Bool("s", cfg.EnableHTTPS, "EnableHTTPS")
	configFile := flag.String("c", cfg.Config, "CONFIG")
	configFile2 := flag.String("config", cfg.Config, "CONFIG")
	trustedSubnet := flag.String("t", cfg.TrustedSubnet, "Trusted Subnet address")
	flag.Parse()

	var err error
	parsedHost, parsedPort, err := makeAppHostPort(*appHost)
	if err != nil {
		log.Panic(err)
	}
	if parsedHost != "" {
		cfg.AppHost = parsedHost
	}
	if parsedPort != 0 {
		cfg.AppPort = parsedPort
	}

	cfg.ShortBaseURL = *shortBaseURL
	cfg.FileStoragePath = *fileStoragePath
	cfg.Dsn = *dsn
	cfg.Debug = strings.ToLower(*debugStr) == "debug"

	var enableHTTPS bool
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "s" {
			enableHTTPS = true
		}
	})
	cfg.EnableHTTPS = enableHTTPS
	if *configFile != "" {
		cfg.Config = *configFile
	}
	if *configFile2 != "" {
		cfg.Config = *configFile2
	}
	if *trustedSubnet != "" {
		cfg.TrustedSubnet = *trustedSubnet
	}

	if *grpcAddr != "" {
		cfg.GRPCAddr = *grpcAddr
	}
}

func makeAppHostPort(appHost string) (string, int, error) {
	addr := strings.SplitN(appHost, ":", 2)
	if len(addr) == 2 {
		host := addr[0]
		intValue := 8080
		_, err := fmt.Sscan(addr[1], &intValue)
		if err != nil {
			return "", 0, fmt.Errorf("PORT value is invalid: %s", addr[1])
		}
		return host, intValue, nil

	} else {
		return appHost, 0, nil
	}
}

// UseConfigFile applies config from json file
func (cfg *Config) UseConfigFile() {
	if cfg.Config != "" {
		// открытие файла с конфигурацией
		configFile, err := os.Open(cfg.Config)
		if err != nil {
			return
		}
		defer configFile.Close()

		// создание структуры для конфигурации
		config := Config{}

		// декодирование JSON файла в структуру Config
		jsonParser := json.NewDecoder(configFile)
		jsonParser.DisallowUnknownFields() // отклонение неизвестных полей
		err = jsonParser.Decode(&config)
		if err != nil {
			return
		}
		err = mergeConfigs(cfg, &config)
		if err != nil {
			return
		}
	}
}

// mergeConfigs merges configs into one
// first config values have priority
func mergeConfigs(result, cfg2 *Config) error {
	if result.AppHost == "" || result.AppHost == "localhost" {
		if cfg2.AppHost != "" {
			host, port, err := makeAppHostPort(cfg2.AppHost)
			if err != nil {
				return err
			}
			log.Printf("parsed host=%s port=%d", host, port)
			if host != "" {
				result.AppHost = host
			}
			if port != 0 {
				result.AppPort = port
			}
		}
	}

	if result.ShortBaseURL == "" || result.ShortBaseURL == "http://localhost:8080" {
		result.ShortBaseURL = cfg2.ShortBaseURL
	}
	if result.FileStoragePath == "" {
		result.FileStoragePath = cfg2.FileStoragePath
	}

	if result.Dsn == "" {
		result.Dsn = cfg2.Dsn
	}
	if !result.Debug {
		result.Debug = cfg2.Debug
	}
	if !result.EnableHTTPS {
		result.EnableHTTPS = cfg2.EnableHTTPS
	}

	if result.TrustedSubnet == "" {
		result.TrustedSubnet = cfg2.TrustedSubnet
	}

	if result.GRPCAddr == "" {
		result.GRPCAddr = cfg2.GRPCAddr
	}

	return nil
}
