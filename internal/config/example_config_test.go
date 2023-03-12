package config_test

import (
	"fmt"
	"github.com/itksb/go-url-shortener/internal/config"
)

func ExampleNewConfig() {
	conf := config.Config{
		// AppPort http server port of the app
		AppPort: 8080,
		// ShortBaseURL base url
		ShortBaseURL: "http://localhost:8080",
		// AppHost where application runs
		AppHost: "localhost",
		// FileStoragePath path for the file storage dir.
		// Leave it empty if no file storage is used
		FileStoragePath: "",
		// SessionConfig session configuration
		SessionConfig: config.SessionConfig{
			HashKey:  "1234567890",
			BlockKey: "0123456701234567" + "0123456701234567",
		},
		// Dsn data source name. Fill it if db is used
		Dsn: "host=localhost port=5432 user=user password=password dbname=postgres sslmode=disable",
		// Debug enables debug routes
		Debug: false,
	}

	fmt.Println(fmt.Sprintf(
		"[AppPort:%d ShortBaseURL:%s AppHost:%s FileStoragePath:%s Dsn:%s]",
		conf.AppPort, conf.ShortBaseURL, conf.AppHost, conf.FileStoragePath,
		conf.Dsn,
	))

	// Output:
	// [AppPort:8080 ShortBaseURL:http://localhost:8080 AppHost:localhost FileStoragePath: Dsn:host=localhost port=5432 user=user password=password dbname=postgres sslmode=disable]
}
