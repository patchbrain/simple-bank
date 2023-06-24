package util

import (
	"github.com/spf13/viper"
	"time"
)

type Config struct {
	Environment          string        `mapstructure:"ENVIRONMENT"`
	RedisAddress         string        `mapstructure:"REDIS_ADDRESS"`
	DbDriver             string        `mapstructure:"DB_DRIVER"`
	DbSource             string        `mapstructure:"DB_SOURCE"`
	MigrateUrl           string        `mapstructure:"MIGRATE_URL"`
	HTTPServerAddress    string        `mapstructure:"HTTP_SERVER_ADDRESS"`
	GRPCServerAddress    string        `mapstructure:"GRPC_SERVER_ADDRESS"`
	TokenSecretKey       string        `mapstructure:"TOKEN_SECRET_KEY"`
	TokenDuration        time.Duration `mapstructure:"TOKEN_DURATION"`
	RefreshTokenDuration time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
}

func LoadConfig(path string) (Config, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return Config{}, err
	}

	cfg := Config{}
	if err := viper.Unmarshal(&cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
