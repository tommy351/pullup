package cmd

import (
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

func RunManager(mgr manager.Manager, cleanup func(), err error) error {
	if err != nil {
		return err
	}

	defer cleanup()

	return mgr.Start(signals.SetupSignalHandler())
}
