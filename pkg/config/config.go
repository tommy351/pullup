package config

import (
	"github.com/ansel1/merry"
	"github.com/spf13/viper"
)

type Config struct {
	Server       ServerConfig     `mapstructure:"server"`
	Repositories []RepoConfig     `mapstructure:"repositories"`
	GitHub       GitHubConfig     `mapstructure:"github"`
	Kubernetes   KubernetesConfig `mapstructure:"kubernetes"`
	Log          LogConfig        `mapstructure:"log"`
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

	Patch []ResourcePatch `mapstructure:"patch"`
}

type ResourcePatch struct {
	Path  string      `mapstructure:"path"`
	Value interface{} `mapstructure:"value"`

	// Ops
	Add     string `mapstructure:"add"`
	Remove  string `mapstructure:"remove"`
	Replace string `mapstructure:"replace"`
	Copy    string `mapstructure:"copy"`
	Move    string `mapstructure:"move"`
	Test    string `mapstructure:"test"`
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

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
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
	v.SetDefault("log.level", "info")

	if err := v.ReadInConfig(); err != nil {
		return nil, merry.Wrap(err)
	}

	if err := v.Unmarshal(&config); err != nil {
		return nil, merry.Wrap(err)
	}

	return &config, nil
}

func MustReadConfig() *Config {
	conf, err := ReadConfig()

	if err != nil {
		panic(err)
	}

	return conf
}
