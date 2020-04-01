package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
)

var ErrConfigNotExist = errors.New("config does not exist")

type Configuration struct {
	Address string
	Repo    RepositoryConfig
	Token   TokenConfig
}

type RepositoryConfig struct {
	Address string
}

type TokenConfig struct {
	Refresh RefreshConfig
	Jwt     JWTConfig
}

type RefreshConfig struct {
	Length     int
	Expiration int
}

// SetDefaults will set the defaults for a refresh token if no configuration exists
func (rc *RefreshConfig) SetDefaults() {
	if rc.Length == 0 {
		rc.Length = 32
	} else if rc.Expiration == 0 {
		rc.Expiration = 24
	}
}

type JWTConfig struct {
	Expiration int
}

// SetDefaults will set the defaults for a jwt if no configuration exists
func (jc *JWTConfig) SetDefaults() {
	if jc.Expiration == 0 {
		jc.Expiration = 15
	}
}

// ParseConfig will look for a config file in a specified directory.
// Returns ErrConfigNotExist when the configuration file can't be found
func ParseConfig(path string) (*Configuration, error) {
	viper.SetConfigName("config")
	viper.AddConfigPath(path)

	config := &Configuration{}

	if err := viper.ReadInConfig(); err != nil {
		return nil, ErrConfigNotExist
	}

	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration file: %w", err)
	}

	// set defaults if don't exist
	config.Token.Refresh.SetDefaults()
	config.Token.Jwt.SetDefaults()

	return config, nil
}
