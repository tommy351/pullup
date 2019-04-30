package testutil

import (
	"path/filepath"
	"runtime"
)

func ProjectDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..")
}

func AssetBinPath(name string) string {
	return filepath.Join(ProjectDir(), "assets", "bin", name)
}
