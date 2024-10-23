package config

import (
	"github.com/AkshayDubey29/MoniFlux/backend/internal/common"
	"github.com/AkshayDubey29/MoniFlux/backend/internal/config/utils"
)

// LoadConfig loads the application configuration
func LoadConfig(path string) (*common.Config, error) {
	return utils.LoadConfig(path)
}
