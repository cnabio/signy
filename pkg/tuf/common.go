package tuf

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/theupdateframework/notary/tuf/data"
)

const (
	dockerConfigDir  = ".docker"
	releasesRoleName = data.RoleName("targets/releases")
)

// DefaultTrustDir returns where the Signy trust data lives
func DefaultTrustDir() string {
	homeEnvPath := os.Getenv("HOME")
	if homeEnvPath == "" && runtime.GOOS == "windows" {
		homeEnvPath = os.Getenv("USERPROFILE")
	}

	return filepath.Join(homeEnvPath, ".signy")
}

// DefaultDockerCfgDir returns where the Docker config directory lives
func DefaultDockerCfgDir() string {
	homeEnvPath := os.Getenv("HOME")
	if homeEnvPath == "" && runtime.GOOS == "windows" {
		homeEnvPath = os.Getenv("USERPROFILE")
	}

	return filepath.Join(homeEnvPath, dockerConfigDir)
}

// EnsureTrustDir ensures the trust directory exists
func EnsureTrustDir(trustDir string) error {
	return os.MkdirAll(trustDir, 0700)
}
