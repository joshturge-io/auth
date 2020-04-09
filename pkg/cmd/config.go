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
	Cipher  CipherConfig
	Token   TokenConfig
}

// SetDefaults will set the defaults for our config struct
func (c *Configuration) SetDefaults() {
	switch {
	case c.Repo.FlushInterval == 0:
		c.Repo.FlushInterval = 15
	case c.Cipher.SaltLength == 0:
		c.Cipher.SaltLength = 16
	case c.Token.Refresh.Expiration == 0:
		c.Token.Refresh.Expiration = 24
	case c.Token.Refresh.Length == 0:
		c.Token.Refresh.Length = 32
	case c.Token.Jwt.Expiration == 0:
		c.Token.Jwt.Expiration = 15
	}
}

type RepositoryConfig struct {
	Address       string
	FlushInterval int
}

type CipherConfig struct {
	SaltLength int
	Keys       []string
}

type TokenConfig struct {
	Refresh RefreshConfig
	Jwt     JWTConfig
}

type RefreshConfig struct {
	Length     int
	Expiration int
}

type JWTConfig struct {
	Expiration int
}

// ParseConfig will look for a config file in a specified directory.
// Returns ErrConfigNotExist when the configuration file can't be found
func ParseConfig(path string) (*Configuration, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.AddConfigPath(path)

	if err := v.ReadInConfig(); err != nil {
		return nil, ErrConfigNotExist
	}

	config := &Configuration{}

	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration file: %w", err)
	}

	config.SetDefaults()

	return config, nil
}
