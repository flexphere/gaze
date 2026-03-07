package config

import "github.com/flexphere/gaze/internal/domain"

// Defaults returns the default application configuration.
func Defaults() *domain.Config {
	return domain.DefaultConfig()
}
