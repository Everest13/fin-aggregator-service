package config

import (
	"fmt"
	"github.com/spf13/viper"
	"time"
)

const (
	MonzoClientID       = "MONZO_CLIENT_ID"
	MonzoClientSecret   = "MONZO_CLIENT_SECRET"
	MonzoRedirectURI    = "monzo.redirect_uri"
	GRPCPort            = "grpc.port"
	GRPCNetwork         = "grpc.network"
	HTTPPort            = "http.port"
	HTTPHost            = "http.host"
	HTTPClientTimeout   = "http.client_timeout"
	HTTPGracefulTimeout = "http.graceful_timeout"
	DBName              = "database.name"
	DBUser              = "database.user"
	DBPassword          = "database.password"
	DBHost              = "database.host"
	DBPort              = "database.port"
	SSLMode             = "database.ssl_mode"
	DBMaxCons           = "database.max_cons"
	DBMinCons           = "database.min_cons"
	DBMaxConLifetime    = "database.max_con_lifetime"
)

type Monzo struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

type GRPCConfig struct {
	Port    string
	Network string
}

type HTTPConfig struct {
	Port            string
	Host            string
	GracefulTimeout time.Duration
	ClientTimeout   time.Duration
}

type DBConfig struct {
	Name           string
	User           string
	Password       string
	Host           string
	Port           uint16
	SSLMode        string
	MaxCons        int32
	MinCons        int32
	MaxConLifetime time.Duration
}

type Config struct {
	GRPC  GRPCConfig
	HTTP  HTTPConfig
	Monzo Monzo
	DB    DBConfig
}

func LoadValues() error {
	viper.SetConfigName(".env_local")
	viper.SetConfigType("env")
	viper.AddConfigPath("config")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to load config/.env_local: %w", err)
	}

	viper.SetConfigName("values_local")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	if err := viper.MergeInConfig(); err != nil {
		return fmt.Errorf("failed to merge config/values_local.yaml: %w", err)
	}

	return nil
}
