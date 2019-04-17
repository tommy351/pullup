package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tommy351/pullup/pkg/apis/pullup/v1alpha1"
	"github.com/tommy351/pullup/pkg/controller/resourceset"
	webhookctrl "github.com/tommy351/pullup/pkg/controller/webhook"
	"github.com/tommy351/pullup/pkg/k8s"
	"github.com/tommy351/pullup/pkg/log"
	"github.com/tommy351/pullup/pkg/webhook"
	"golang.org/x/xerrors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// nolint: gochecknoglobals
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type Config struct {
	Log        log.Config     `mapstructure:"log"`
	Kubernetes k8s.Config     `mapstructure:"kubernetes"`
	Webhook    webhook.Config `mapstructure:"webhook"`
}

func loadConfig() *Config {
	var config Config

	if err := viper.Unmarshal(&config); err != nil {
		panic(err)
	}

	return &config
}

func bindEnv(key, env string) {
	if v := os.Getenv(env); v != "" {
		viper.Set(key, v)
	}
}

func newController(name string, mgr manager.Manager, logger logr.Logger, kind runtime.Object, reconciler reconcile.Reconciler) error {
	ctrl, err := controller.New(name, mgr, controller.Options{
		Reconciler: reconciler,
	})

	if err != nil {
		return xerrors.Errorf("failed to create a controller: %w", err)
	}

	logger = logger.WithName("controller").WithName(name)

	if _, err := inject.LoggerInto(logger, reconciler); err != nil {
		return xerrors.Errorf("failed to inject logger into reconciler: %w", err)
	}

	err = ctrl.Watch(&source.Kind{Type: kind}, &handler.EnqueueRequestForObject{})

	if err != nil {
		return xerrors.Errorf("failed to watch resource: %w", err)
	}

	return nil
}

func run(_ *cobra.Command, _ []string) error {
	conf := loadConfig()
	logger := log.New(&conf.Log)
	defer logger.Flush()

	kubeConf, err := k8s.LoadConfig(&conf.Kubernetes)

	if err != nil {
		return xerrors.Errorf("failed to load Kubernetes config: %w", err)
	}

	mgr, err := manager.New(kubeConf, manager.Options{
		Namespace: conf.Kubernetes.Namespace,
	})

	if err != nil {
		return xerrors.Errorf("failed to create a manager: %w", err)
	}

	sb := runtime.NewSchemeBuilder(v1alpha1.AddToScheme)

	if err := sb.AddToScheme(mgr.GetScheme()); err != nil {
		return xerrors.Errorf("failed to register scheme: %w", err)
	}

	eventRecorder := mgr.GetEventRecorderFor("pullup")

	err = newController("webhook", mgr, logger, &v1alpha1.Webhook{}, &webhookctrl.Reconciler{})

	if err != nil {
		return xerrors.Errorf("failed to create a webhook controller: %w", err)
	}

	err = newController("resource-set", mgr, logger, &v1alpha1.ResourceSet{}, &resourceset.Reconciler{
		EventRecorder: eventRecorder,
	})

	if err != nil {
		return xerrors.Errorf("failed to create a resource set controller: %w", err)
	}

	webhookServer := &webhook.Server{
		Config:    conf.Webhook,
		Namespace: conf.Kubernetes.Namespace,
	}

	if _, err := inject.LoggerInto(logger.WithName("webhook"), webhookServer); err != nil {
		return xerrors.Errorf("failed to inject logger into webhook server: %w", err)
	}

	err = mgr.Add(webhookServer)

	if err != nil {
		return xerrors.Errorf("failed to add webhook server to manager: %w", err)
	}

	return mgr.Start(signals.SetupSignalHandler())
}

func newCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "pullup",
		Short:        "Deploy pull requests before merged",
		Version:      fmt.Sprintf("%s, commit %s, built at %s", version, commit, date),
		RunE:         run,
		SilenceUsage: true,
	}

	cmd.SetVersionTemplate("{{ .Version }}")

	f := cmd.Flags()

	// Bind flags
	f.String("log-level", "", "log level")
	_ = viper.BindPFlag("log.level", f.Lookup("log-level"))
	viper.SetDefault("log.level", "info")

	f.String("namespace", "", "kubernetes namespace")
	_ = viper.BindPFlag("kubernetes.namespace", f.Lookup("namespace"))
	viper.SetDefault("kubernetes.namespace", metav1.NamespaceDefault)

	f.String("kubeconfig", "", "kubernetes config path")
	_ = viper.BindPFlag("kubernetes.config", f.Lookup("kubeconfig"))
	bindEnv("kubernetes.config", "KUBECONFIG")

	f.String("webhook-address", "", "webhook listening address")
	_ = viper.BindPFlag("webhook.address", f.Lookup("webhook-address"))
	viper.SetDefault("webhook.address", ":4000")

	// Bind environment variables
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	return cmd
}

func main() {
	if err := newCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
