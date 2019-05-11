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

type Config struct {
	Log        log.Config `mapstructure:"log"`
	Kubernetes k8s.Config `mapstructure:"kubernetes"`
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

	f.String("namespace", "", "kubernetes namespace")
	_ = viper.BindPFlag("kubernetes.namespace", f.Lookup("namespace"))
	viper.SetDefault("kubernetes.namespace", v1.NamespaceDefault)

	f.String("kubeconfig", "", "kubernetes config path")
	_ = viper.BindPFlag("kubernetes.config", f.Lookup("kubeconfig"))
	BindEnv("kubernetes.config", "KUBECONFIG")

	// Bind environment variables
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	return cmd
}
