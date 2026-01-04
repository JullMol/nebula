package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server ServerConfig `mapstructure:"server"`
	Worker WorkerConfig `mapstructure:"worker"`
}

type ServerConfig struct {
	Port    string   `mapstructure:"port"`
	Workers []string `mapstructure:"workers"`
	RedisAddr string `mapstructure:"redis_addr"`
}

type WorkerConfig struct {
	Port string `mapstructure:"port"`
	Name string `mapstructure:"name"`
}

func LoadConfig() (*Config, error) {
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}