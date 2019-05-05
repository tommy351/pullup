package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tommy351/pullup/cmd"
	"github.com/tommy351/pullup/pkg/webhook"
	"github.com/tommy351/pullup/pkg/webhook/github"
)

type Config struct {
	cmd.Config `mapstructure:",squash"`

	Webhook webhook.Config `mapstructure:"webhook"`
	GitHub  github.Config  `mapstructure:"github"`
}

func NewConfig(conf Config) cmd.Config {
	return conf.Config
}

func run(_ *cobra.Command, _ []string) error {
	var conf Config

	if err := viper.Unmarshal(&conf); err != nil {
		return err
	}

	return cmd.RunManager(InitializeManager(conf))
}

func main() {
	cmd := cmd.SetupCommand(&cobra.Command{
		Use:  "pullup-webhook",
		RunE: run,
	})

	f := cmd.Flags()

	f.String("address", "", "listening address")
	_ = viper.BindPFlag("webhook.address", f.Lookup("address"))
	viper.SetDefault("webhook.address", ":8080")

	f.String("github-secret", "", "GitHub secret")
	_ = viper.BindPFlag("github.secret", f.Lookup("github-secret"))

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
