package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct{}

func New() *Config {
	return &Config{}
}

func (c *Config) LoadFromFile(file string, data any) error {
	viper.SetConfigFile(file)
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	return viper.Unmarshal(data)
}
