package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server       ServerConfig     `mapstructure:"server"`
	Repositories []RepoConfig     `mapstructure:"repositories"`
	GitHub       GitHubConfig     `mapstructure:"github"`
	Kubernetes   KubernetesConfig `mapstructure:"kubernetes"`
}

type ServerConfig struct {
	Addr string `mapstructure:"addr"`
}

type RepoConfig struct {
	Name      string           `mapstructure:"name"`
	Resources []ResourceConfig `mapstructure:"resources"`
}

type ResourceConfig struct {
	ResourceReference `mapstructure:",squash"`

	Merge interface{} `mapstructure:"merge"`
}

type ResourceReference struct {
	APIVersion string `mapstructure:"apiVersion"`
	Kind       string `mapstructure:"kind"`
	Name       string `mapstructure:"name"`
}

type GitHubConfig struct {
	Secret string `mapstructure:"secret"`
}

type KubernetesConfig struct {
	Namespace string `mapstructure:"namespace"`
}

func ReadConfig() (*Config, error) {
	var config Config
	v := viper.New()

	// Set paths to config files
	v.SetConfigName("pullup")
	v.AddConfigPath("/etc/pullup")
	v.AddConfigPath("$HOME/.pullup")
	v.AddConfigPath(".")

	// Bind to environment variables
	v.AutomaticEnv()
	v.SetEnvPrefix("pullup")

	// Default values
	v.SetDefault("server.addr", ":4000")
	v.SetDefault("kubernetes.namespace", "default")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	if err := v.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
