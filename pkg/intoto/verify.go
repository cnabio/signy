package intoto

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cnabio/signy/pkg/docker"
	log "github.com/sirupsen/logrus"
	"github.com/theupdateframework/notary/client"
)

const (
	// Fixed as per CNAB-100 spec
	BundleFilename = "bundle.json"
	ReadOnlyMask   = 0400
)

func Verify(verifyOnOS bool, verificationImage string, target *client.TargetWithRole, bundle []byte, logLevel string) error {
	m := &Metadata{}
	err := json.Unmarshal(*target.Custom, m)
	if err != nil {
		return err
	}

	verificationDir, err := ioutil.TempDir(os.TempDir(), "in-toto")
	if err != nil {
		return err
	}
	defer func() {
		os.RemoveAll(verificationDir)
		os.Remove(verificationDir)
	}()

	log.Infof("Writing in-toto metadata files into %v", verificationDir)
	err = WriteMetadataFiles(m, verificationDir)
	if err != nil {
		return err
	}

	bundleFilename := filepath.Join(verificationDir, BundleFilename)
	err = ioutil.WriteFile(bundleFilename, bundle, ReadOnlyMask)
	if err != nil {
		log.Fatal(err)
	}

	if verifyOnOS {
		return VerifyOnOS(verificationDir)
	}
	return docker.Run(verificationImage, verificationDir, logLevel)
}
