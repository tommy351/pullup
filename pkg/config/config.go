package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server       ServerConfig `mapstructure:"server"`
	Repositories []RepoConfig `mapstructure:"repositories"`
	GitHub       GitHubConfig `mapstructure:"github"`
}

type ServerConfig struct {
	Addr string `mapstructure:"addr"`
}

type RepoConfig struct {
	Name      string           `mapstructure:"name"`
	Resources []ResourceConfig `mapstructure:"resources"`
}

type ResourceConfig struct {
	APIVersion string `mapstructure:"apiVersion"`
	Kind       string `mapstructure:"kind"`
	Name       string `mapstructure:"name"`
	Type       string `mapstructure:"type"`
}

type GitHubConfig struct {
	Secret string `mapstructure:"secret"`
}

func ReadConfig() (*Config, error) {
	var config Config
	v := viper.New()
	v.SetConfigName("pullup")
	v.AddConfigPath("/etc/pullup")
	v.AddConfigPath("$HOME/.pullup")
	v.AddConfigPath(".")
	v.AutomaticEnv()
	v.SetEnvPrefix("pullup")
	v.SetDefault("server.addr", ":4000")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	if err := v.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
