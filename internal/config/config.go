package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Env      string
	Server   ServerConfig
	Pilot    PilotConfig
	Paths    PathsConfig
	OpenClaw OpenClawConfig
}

type ServerConfig struct {
	Host            string
	Port            int
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

func (s *ServerConfig) GetAddress() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

type PilotConfig struct {
	Username  string
	Password  string
	JWTSecret string        `mapstructure:"jwt_secret"`
	JWTExpiry time.Duration `mapstructure:"jwt_expiry"`
}

type PathsConfig struct {
	EngineeringTasks string `mapstructure:"engineering_tasks_path"`
	PrasMemory       string `mapstructure:"pras_memory_path"`
}

type OpenClawConfig struct {
	EngineeringLeadChannel string `mapstructure:"engineering_lead_channel"`
	EngineeringDevChannel  string `mapstructure:"engineering_dev_channel"`
}

func (c *Config) IsDevelopment() bool {
	return c.Env == "" || c.Env == "development" || c.Env == "dev"
}

func (c *Config) IsProduction() bool {
	return c.Env == "production" || c.Env == "prod"
}

func Load() (*Config, error) {
	_ = godotenv.Load(".env")

	v := viper.New()

	v.SetDefault("server.port", 8080)
	v.SetDefault("server.shutdown_timeout", "10s")
	v.SetDefault("pilot.jwt_expiry", "24h")

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Explicit bindings for env vars
	_ = v.BindEnv("pilot.username", "PILOT_USERNAME")
	_ = v.BindEnv("pilot.password", "PILOT_PASSWORD")
	_ = v.BindEnv("pilot.jwt_secret", "PILOT_JWT_SECRET")
	_ = v.BindEnv("paths.engineering_tasks_path", "ENGINEERING_TASKS_PATH")
	_ = v.BindEnv("paths.pras_memory_path", "PRAS_MEMORY_PATH")
	_ = v.BindEnv("openclaw.engineering_lead_channel", "OPENCLAW_ENGINEERING_LEAD_CHANNEL")
	_ = v.BindEnv("openclaw.engineering_dev_channel", "OPENCLAW_ENGINEERING_DEV_CHANNEL")

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func validate(cfg *Config) error {
	if cfg.Pilot.Username == "" {
		return fmt.Errorf("PILOT_USERNAME is required")
	}
	if cfg.Pilot.Password == "" {
		return fmt.Errorf("PILOT_PASSWORD is required")
	}
	if cfg.Pilot.JWTSecret == "" {
		return fmt.Errorf("PILOT_JWT_SECRET is required")
	}
	if cfg.Paths.EngineeringTasks == "" {
		return fmt.Errorf("ENGINEERING_TASKS_PATH is required")
	}
	if _, err := os.Stat(cfg.Paths.EngineeringTasks); os.IsNotExist(err) {
		return fmt.Errorf("ENGINEERING_TASKS_PATH does not exist: %s", cfg.Paths.EngineeringTasks)
	}
	return nil
}
