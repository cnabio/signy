package intoto

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/theupdateframework/notary/client"

	"github.com/cnabio/signy/pkg/docker"
)

const (
	// BundleFilename is the name written to disk during verification
	BundleFilename = "bundle.json"
	// ReadOnlyMask is applied to files written to disk for verification
	ReadOnlyMask = 0400
)

// VerifyOnOS verifies a bundle directly on the OS
func VerifyOnOS(target *client.TargetWithRole, bundle []byte) error {
	verificationDir, err := getVerificationDir(target, bundle)
	if err != nil {
		return err
	}
	defer func() {
		os.RemoveAll(verificationDir)
		os.Remove(verificationDir)
	}()
	return verifyOnOS(verificationDir)
}

// VerifyInContainer verifies a bundle within a container
func VerifyInContainer(target *client.TargetWithRole, bundle []byte, verificationImage string, logLevel string) error {
	verificationDir, err := getVerificationDir(target, bundle)
	if err != nil {
		return err
	}
	defer func() {
		os.RemoveAll(verificationDir)
		os.Remove(verificationDir)
	}()
	return docker.Run(verificationImage, verificationDir, logLevel)
}

func getVerificationDir(target *client.TargetWithRole, bundle []byte) (string, error) {
	// FIXME
	m := &Metadata{}
	err := json.Unmarshal(*target.Custom, m)
	if err != nil {
		return "", err
	}

	verificationDir, err := ioutil.TempDir(os.TempDir(), "in-toto")
	if err != nil {
		return "", err
	}

	// FIXME
	log.Infof("Writing in-toto metadata files into %v", verificationDir)
	err = WriteMetadataFiles(m, verificationDir)
	if err != nil {
		return "", err
	}

	bundleFilename := filepath.Join(verificationDir, BundleFilename)
	err = ioutil.WriteFile(bundleFilename, bundle, ReadOnlyMask)
	if err != nil {
		return "", err
	}

	return verificationDir, nil
}
