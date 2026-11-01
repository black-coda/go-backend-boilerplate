package config

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog"
)

type Config struct {
	Primary       Primary              `koanf:"primary" validate:"required"`
	Server        ServerConfig         `koanf:"server" validate:"required"`
	Database      DatabaseConfig       `koanf:"database" validate:"required"`
	Auth          AuthConfig           `koanf:"auth" validate:"required"`
	Redis         RedisConfig          `koanf:"redis" validate:"required"`
	Integration   IntegrationConfig    `koanf:"integration" validate:"required"`
	Observability *ObservabilityConfig `koanf:"observability"`
}

type Primary struct {
	Env string `koanf:"env" validate:"required"`
}

type ServerConfig struct {
	Port               string   `koanf:"port" validate:"required"`
	ReadTimeout        int      `koanf:"read_timeout" validate:"required"`
	WriteTimeout       int      `koanf:"write_timeout" validate:"required"`
	IdleTimeout        int      `koanf:"idle_timeout" validate:"required"`
	CORSAllowedOrigins []string `koanf:"cors_allowed_origins" validate:"required"`
}

type DatabaseConfig struct {
	Host            string `koanf:"host" validate:"required"`
	Port            int    `koanf:"port" validate:"required"`
	User            string `koanf:"user" validate:"required"`
	Password        string `koanf:"password"`
	Name            string `koanf:"name" validate:"required"`
	SSLMode         string `koanf:"ssl_mode" validate:"required"`
	MaxOpenConns    int    `koanf:"max_open_conns" validate:"required"`
	MaxIdleConns    int    `koanf:"max_idle_conns" validate:"required"`
	ConnMaxLifetime int    `koanf:"conn_max_lifetime" validate:"required"`
	ConnMaxIdleTime int    `koanf:"conn_max_idle_time" validate:"required"`
}
type RedisConfig struct {
	Address string `koanf:"address" validate:"required"`
}

type IntegrationConfig struct {
	ResendAPIKey string `koanf:"resend_api_key" validate:"required"`
}

type AuthConfig struct {
	SecretKey string `koanf:"secret_key" validate:"required"`
}

func LoadConfig() (*Config, error) {
	buildInfo, _ := debug.ReadBuildInfo()
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).
		Level(zerolog.TraceLevel).
		With().
		Timestamp().
		Caller().
		Int("pid", os.Getpid()).
		Str("go_version", buildInfo.GoVersion).
		Logger()
	logger.WithLevel(zerolog.PanicLevel).Msg("panic message")

	k := koanf.New(".")

	// Loading YAML Config
	if err := k.Load(
		file.Provider("config.yaml"),
		yaml.Parser(),
	); err != nil {
		logger.Error().Err(err).Msg("Load config Error")
		return nil, fmt.Errorf("load config: %w", err)
	}

	// Loading from ENV Config
	// Load only environment variables with prefix "BOILER_" and merge into config.
	// Transform var names by:
	// 1. Converting to lowercase
	// 2. Removing "BOILER_" prefix
	// 3. Replacing "_" with "." to representing nesting using the . delimiter.
	// Example: BOILER_PARENT1_CHILD1_NAME becomes "parent1.child1.name"
	if err := k.Load(env.Provider(".", env.Opt{
		Prefix: "BOILER_",
		TransformFunc: func(k, v string) (string, any) {
			// Transform the key.
			k = strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(k, "BOILER_")), "_", ".")

			// Transform the value into slices, if they contain spaces.
			// Eg: MYVAR_TAGS="foo bar baz" -> tags: ["foo", "bar", "baz"]
			// This is to demonstrate that string values can be transformed to any type
			// where necessary.
			if strings.Contains(v, " ") {
				return k, strings.Split(v, " ")
			}

			return k, v
		},
	}), nil); err != nil {
		logger.Error().Err(err).Msg("Load env Error")
	}

	mainConfig := &Config{}
	err := k.Unmarshal("", mainConfig)
	if err != nil {
		logger.Fatal().Err(err).Msg("Error unmarshal config")
	}

	if mainConfig.Observability == nil {
		mainConfig.Observability = &ObservabilityConfig{}
	}
	mainConfig.Observability.Environment = mainConfig.Primary.Env

	if err := mainConfig.Observability.Validate(); err != nil {
		logger.Fatal().Err(err).Msg("Error validate config")
	}
	return mainConfig, nil
}
