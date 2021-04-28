package intoto

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/theupdateframework/notary/client"

	"github.com/cnabio/signy/pkg/docker"
)

const (
	BundleFilename = "bundle.json"
	ReadOnlyMask   = 0400
)

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
	m := &Metadata{}
	err := json.Unmarshal(*target.Custom, m)
	if err != nil {
		return "", err
	}

	verificationDir, err := ioutil.TempDir(os.TempDir(), "in-toto")
	if err != nil {
		return "", err
	}

	/*
		TODO: Figure out a better way to do this.
		For example, if we have a tar.gz file that was not pushed to notary that needs verified, we need to figure out a way to run in-toto verification on it.
		For a lot of in-toto demo example, it uses a .tar.gz example. In the future since we're using docker images, it should pull all data for verification from the image itself.
	*/
	log.Infof("Copy all files to the temp directory %v for verification", verificationDir)
	err = copy("/Users/scottbuckel/Work/signy/signy/bin", verificationDir)
	if err != nil {
		return "", err
	}

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

func copy(source, destination string) error {
	var err error = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		var relPath string = strings.Replace(path, source, "", 1)
		if relPath == "" {
			return nil
		}
		if info.IsDir() {
			return os.Mkdir(filepath.Join(destination, relPath), 0755)
		} else {
			var data, err1 = ioutil.ReadFile(filepath.Join(source, relPath))
			if err1 != nil {
				return err1
			}
			return ioutil.WriteFile(filepath.Join(destination, relPath), data, 0777)
		}
	})
	return err
}
