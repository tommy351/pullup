package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/google/wire"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tommy351/pullup/internal/k8s"
	"github.com/tommy351/pullup/internal/log"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConfigSet provides configs.
// nolint: gochecknoglobals
var ConfigSet = wire.NewSet(NewLogConfig, NewKubernetesConfig)

type MetricsConfig struct {
	Address string `mapstructure:"address"`
}

type HealthProbeConfig struct {
	Address string `mapstructure:"address"`
}

type Config struct {
	Log        log.Config        `mapstructure:"log"`
	Kubernetes k8s.Config        `mapstructure:"kubernetes"`
	Metrics    MetricsConfig     `mapstructure:"metrics"`
	Health     HealthProbeConfig `mapstructure:"health"`
}

func NewLogConfig(conf Config) log.Config {
	return conf.Log
}

func NewKubernetesConfig(conf Config) k8s.Config {
	return conf.Kubernetes
}

func BindEnv(key, env string) {
	if v := os.Getenv(env); v != "" {
		viper.Set(key, v)
	}
}

func SetupCommand(cmd *cobra.Command) *cobra.Command {
	cmd.Version = fmt.Sprintf("%s, commit %s, built at %s", Version, Commit, Date)
	cmd.SilenceUsage = true

	cmd.SetVersionTemplate("{{ .Version }}")

	f := cmd.Flags()

	// Bind flags
	f.String("log-level", "", "log level")
	_ = viper.BindPFlag("log.level", f.Lookup("log-level"))
	viper.SetDefault("log.level", "info")

	f.Bool("log-dev", false, "enable dev mode for logging")
	_ = viper.BindPFlag("log.dev", f.Lookup("log-dev"))
	viper.SetDefault("log.dev", false)

	f.String("namespace", "", "kubernetes namespace")
	_ = viper.BindPFlag("kubernetes.namespace", f.Lookup("namespace"))
	viper.SetDefault("kubernetes.namespace", v1.NamespaceDefault)

	f.String("kubeconfig", "", "kubernetes config path")
	_ = viper.BindPFlag("kubernetes.config", f.Lookup("kubeconfig"))
	BindEnv("kubernetes.config", "KUBECONFIG")

	f.String("metrics-address", "", "metrics address")
	_ = viper.BindPFlag("metrics.address", f.Lookup("metrics-address"))
	viper.SetDefault("metrics.address", ":9100")

	f.String("health-address", "", "health probe address")
	_ = viper.BindPFlag("health.address", f.Lookup("health-address"))
	viper.SetDefault("health.address", ":9101")

	// Bind environment variables
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	return cmd
}
