package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
)

var ErrConfigNotExist = errors.New("config does not exist")

type Configuration struct {
	ServerAddr string
	Repo       RepositoryConfig
	Token      TokenConfig
}

type RepositoryConfig struct {
	Address string
}

type TokenConfig struct {
	RefreshLength     int
	RefreshExpiration int
	JWTExpiration     int
}

// ParseConfig will look for a config file in a specified directory.
// Returns ErrConfigNotExist when the configuration file can't be found
func ParseConfig(path string) (*Configuration, error) {
	viper.SetConfigName("config")
	viper.AddConfigPath(path)

	config := &Configuration{}

	if err := viper.ReadRemoteConfig(); err != nil {
		return nil, ErrConfigNotExist
	}

	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration file: %w", err)
	}

	return config, nil
}
