package cmd

import (
	"fmt"
	"net/http"

	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

func RunManager(mgr manager.Manager, cleanup func(), err error) error {
	if err != nil {
		return err
	}

	defer cleanup()

	if err := mgr.AddHealthzCheck("default", func(req *http.Request) error {
		return nil
	}); err != nil {
		return fmt.Errorf("failed to add default health check: %w", err)
	}

	if err := mgr.AddReadyzCheck("default", func(req *http.Request) error {
		return nil
	}); err != nil {
		return fmt.Errorf("failed to add default ready check: %w", err)
	}

	return mgr.Start(signals.SetupSignalHandler())
}
