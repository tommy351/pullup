package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tommy351/pullup/cmd"
)

func run(_ *cobra.Command, _ []string) error {
	var conf cmd.Config

	if err := viper.Unmarshal(&conf); err != nil {
		return err
	}

	return cmd.RunManager(InitializeManager(conf))
}

func main() {
	cmd := cmd.SetupCommand(&cobra.Command{
		Use:  "pullup-controller",
		RunE: run,
	})

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
