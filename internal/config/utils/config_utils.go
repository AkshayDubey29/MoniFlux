package utils

import (
    "github.com/spf13/viper"
    "github.com/AkshayDubey29/MoniFlux/internal/config/v1"
)

// LoadConfig loads configuration from config.yaml
func LoadConfig(path string) (*v1.Config, error) {
    var cfg v1.Config

    viper.SetConfigFile(path)
    viper.SetConfigType("yaml")

    if err := viper.ReadInConfig(); err != nil {
        return nil, err
    }

    if err := viper.Unmarshal(&cfg); err != nil {
        return nil, err
    }

    return &cfg, nil
}
