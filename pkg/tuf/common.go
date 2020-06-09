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

	"github.com/docker/distribution/reference"
	"github.com/docker/docker/registry"
	"github.com/theupdateframework/notary/tuf/data"
)

const (
	dockerConfigDir  = ".docker"
	releasesRoleName = data.RoleName("targets/releases")
)

func DefaultTrustDir() string {
	homeEnvPath := os.Getenv("HOME")
	if homeEnvPath == "" && runtime.GOOS == "windows" {
		homeEnvPath = os.Getenv("USERPROFILE")
	}

	return filepath.Join(homeEnvPath, ".signy")
}

func DefaultDockerCfgDir() string {
	homeEnvPath := os.Getenv("HOME")
	if homeEnvPath == "" && runtime.GOOS == "windows" {
		homeEnvPath = os.Getenv("USERPROFILE")
	}

	return filepath.Join(homeEnvPath, dockerConfigDir)
}

// ensures the trust directory exists
func EnsureTrustDir(trustDir string) error {
	return os.MkdirAll(trustDir, 0700)
}

func getRepoAndTag(name string) (*registry.RepositoryInfo, string, error) {
	r, err := reference.ParseNormalizedNamed(name)
	if err != nil {
		return nil, "", err
	}
	repo, err := registry.ParseRepositoryInfo(r)
	if err != nil {
		return nil, "", err
	}

	return repo, getTag(r), nil
}

func getTag(ref reference.Named) string {
	switch x := ref.(type) {
	case reference.Canonical, reference.Digested:
		return ""
	case reference.NamedTagged:
		return x.Tag()
	default:
		return ""
	}
}
