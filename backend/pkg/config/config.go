package config

import (
	"github.com/AkshayDubey29/MoniFlux/backend/internal/config/utils"
	"github.com/AkshayDubey29/MoniFlux/backend/internal/config/v1"
)

// LoadConfig loads the application configuration
func LoadConfig(path string) (*v1.Config, error) {
	return utils.LoadConfig(path)
}
