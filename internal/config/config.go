package config

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	HTTP       HTTPConfig
	Postgres   PostgresConfig
	Telegram   TelegramConfig
	JWT        JWTConfig
	Log        LogConfig
	AdminToken string `mapstructure:"admin_token"`
	UploadsDir string `mapstructure:"uploads_dir"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`  // debug|info|warn|error
	Format string `mapstructure:"format"` // json|text
}

type JWTConfig struct {
	Secret string `mapstructure:"secret"`
}

type HTTPConfig struct {
	Port         string        `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

type PostgresConfig struct {
	DSN string `mapstructure:"dsn"`
}

type TelegramConfig struct {
	Token       string `mapstructure:"token"`
	Debug       bool   `mapstructure:"debug"`
	APIEndpoint string `mapstructure:"api_endpoint"`
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		slog.Debug("no .env file found, relying on environment", "err", err)
	}

	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")

	// Environment variable bindings
	v.SetEnvPrefix("")
	v.AutomaticEnv()

	// Defaults
	v.SetDefault("http.port", "8080")
	v.SetDefault("http.read_timeout", "15s")
	v.SetDefault("http.write_timeout", "15s")
	v.SetDefault("telegram.debug", false)
	v.SetDefault("uploads_dir", "./uploads")
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")

	// Bind explicit env keys to nested config paths
	_ = v.BindEnv("http.port", "HTTP_PORT")
	_ = v.BindEnv("http.read_timeout", "HTTP_READ_TIMEOUT")
	_ = v.BindEnv("http.write_timeout", "HTTP_WRITE_TIMEOUT")
	_ = v.BindEnv("postgres.dsn", "POSTGRES_DSN")
	_ = v.BindEnv("telegram.token", "TELEGRAM_TOKEN")
	_ = v.BindEnv("telegram.debug", "TELEGRAM_DEBUG")
	_ = v.BindEnv("telegram.api_endpoint", "TELEGRAM_API_ENDPOINT")
	_ = v.BindEnv("admin_token", "ADMIN_TOKEN")
	_ = v.BindEnv("jwt.secret", "JWT_SECRET")
	_ = v.BindEnv("uploads_dir", "UPLOADS_DIR")
	_ = v.BindEnv("log.level", "LOG_LEVEL")
	_ = v.BindEnv("log.format", "LOG_FORMAT")

	// Config file is optional — ignore not-found errors
	if err := v.ReadInConfig(); err != nil {
		var notFoundErr viper.ConfigFileNotFoundError
		if !errors.As(err, &notFoundErr) {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if err := validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func validate(cfg *Config) error {
	if cfg.Postgres.DSN == "" {
		return fmt.Errorf("POSTGRES_DSN is required")
	}
	if cfg.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	return nil
}
