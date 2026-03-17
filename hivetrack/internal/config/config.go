package config

import (
	"fmt"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	env "github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	Server       ServerConfig       `koanf:"server"`
	Database     DatabaseConfig     `koanf:"database"`
	OIDC         OIDCConfig         `koanf:"oidc"`
	Email        EmailConfig        `koanf:"email"`
	AI           AIConfig           `koanf:"ai"`
	InitialAdmin InitialAdminConfig `koanf:"initial_admin"`
}

type ServerConfig struct {
	Host           string   `koanf:"host"`
	Port           int      `koanf:"port"`
	ExternalURL    string   `koanf:"external_url"`
	AllowedOrigins []string `koanf:"allowed_origins"`
}

type DatabaseConfig struct {
	URL string `koanf:"url"`
}

type OIDCConfig struct {
	Authority string `koanf:"authority"`
	ClientID  string `koanf:"client_id"`
}

type EmailConfig struct {
	SMTPHost    string `koanf:"smtp_host"`
	SMTPPort    int    `koanf:"smtp_port"`
	SMTPUser    string `koanf:"smtp_user"`
	SMTPPass    string `koanf:"smtp_password"`
	FromAddress string `koanf:"from_address"`
}

type AIConfig struct {
	Enabled  bool   `koanf:"enabled"`
	Provider string `koanf:"provider"`
	Model    string `koanf:"model"`
	APIKey   string `koanf:"api_key"`
}

type InitialAdminConfig struct {
	Email string `koanf:"email"`
}

// Load reads config from configPath YAML file and HIVETRACK_ env vars.
func Load(configPath string) (*Config, error) {
	k := koanf.New(".")

	if err := k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
		return nil, fmt.Errorf("loading config file %s: %w", configPath, err)
	}

	if err := k.Load(env.Provider(".", env.Opt{
		Prefix: "HIVETRACK_",
		TransformFunc: func(k, v string) (string, any) {
			key := strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(k, "HIVETRACK_")), "_", ".")
			return key, v
		},
	}), nil); err != nil {
		return nil, fmt.Errorf("loading env config: %w", err)
	}

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	return &cfg, nil
}
