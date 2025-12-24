//go:build examples

package app

import (
	"path/filepath"
	"runtime"
)

var (
	_, callerFile, _, ok = runtime.Caller(0)
	examplesRoot         = func() string {
		if ok {
			return filepath.Dir(filepath.Dir(callerFile))
		}
		return "."
	}()
)

func resolvePath(p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(examplesRoot, p)
}

func composeFilePath() string {
	return filepath.Join(examplesRoot, "docker-compose.yml")
}
