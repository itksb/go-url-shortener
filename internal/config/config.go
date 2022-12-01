package config

type (
	// Config - Application configuration
	Config struct {
		AppPort      int
		ShortBaseURL string
	}
)

// NewConfig - Constructor
func NewConfig() (Config, error) {
	// We can use environment parser here
	cfg := Config{
		AppPort:      8080,
		ShortBaseURL: "http://localhost:8080",
	}
	return cfg, nil
}
