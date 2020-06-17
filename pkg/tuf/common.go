// Most of the helper functions are adapted from github.com/theupdateframework/notary
//
// Figure out the proper way of making sure we are respecting the licensing from Notary
// While we are also vendoring Notary directly (see LICENSE in vendor/github.com/theupdateframework/notary/LICENSE),
// copying unexported functions could fall under different licensing, so we need to make sure.

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
