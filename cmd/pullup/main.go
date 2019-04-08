package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tommy351/pullup/pkg/group"
	"github.com/tommy351/pullup/pkg/k8s"
	"github.com/tommy351/pullup/pkg/log"
	"github.com/tommy351/pullup/pkg/manager"
	"github.com/tommy351/pullup/pkg/signal"
	"github.com/tommy351/pullup/pkg/webhook"
)

type Config struct {
	Log        log.Config     `mapstructure:"log"`
	Kubernetes k8s.Config     `mapstructure:"kubernetes"`
	Webhook    webhook.Config `mapstructure:"webhook"`
}

// nolint: gochecknoglobals
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// nolint: gochecknoglobals
var rootCmd = &cobra.Command{
	Use:     "pullup",
	Short:   "Deploy pull requests before merged",
	Version: fmt.Sprintf("%s, commit %s, built at %s", version, commit, date),
	Run: func(cmd *cobra.Command, args []string) {
		conf := loadConfig()
		logger := log.New(&conf.Log)
		client, err := k8s.NewClient(&conf.Kubernetes)

		if err != nil {
			logger.Fatal().Stack().Err(err).Msg("Failed to create a Kubernetes client")
		}

		ctx := context.Background()
		ctx = logger.WithContext(ctx)
		ctx = signal.Context(ctx)

		server := &webhook.Server{
			Client: client,
			Config: conf.Webhook,
		}

		mgr := &manager.Manager{
			Client: client,
		}

		g := group.NewGroup(ctx)
		g.Go(server.Serve)
		g.Go(mgr.Run)

		if err := g.Wait(); err != nil {
			logger.Fatal().Stack().Err(err).Msg("Failed to start the server")
		}
	},
}

// nolint: gochecknoinits
func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.SetVersionTemplate("{{ .Version }}")

	f := rootCmd.Flags()

	// Bind flags
	f.String("log-level", "", "log level")
	_ = viper.BindPFlag("log.level", f.Lookup("log-level"))
	viper.SetDefault("log.level", "info")

	f.String("log-format", "", "log format")
	_ = viper.BindPFlag("log.format", f.Lookup("log-format"))

	f.String("namespace", "", "kubernetes namespace")
	_ = viper.BindPFlag("kubernetes.namespace", f.Lookup("namespace"))
	viper.SetDefault("kubernetes.namespace", "default")

	f.String("kubeconfig", "", "kubernetes config path	")
	_ = viper.BindPFlag("kubernetes.config", f.Lookup("kubeconfig"))
	viper.RegisterAlias("kubernetes.config", "kubeconfig")

	f.String("webhook-address", "", "webhook listening address")
	_ = viper.BindPFlag("webhook.address", f.Lookup("webhook-address"))
	viper.SetDefault("webhook.address", ":4000")

	// Bind environment variables
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

func initConfig() {}

func loadConfig() *Config {
	var config Config

	if err := viper.Unmarshal(&config); err != nil {
		panic(err)
	}

	return &config
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
