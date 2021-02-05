package tuf

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"

	"github.com/docker/go/canonical/json"
	log "github.com/sirupsen/logrus"
	"github.com/theupdateframework/notary/client"

	"github.com/cnabio/signy/pkg/cnab"
	"github.com/cnabio/signy/pkg/docker"
)

// GetTargetAndSHA returns the target with roles and the SHA256 of the target file
func GetTargetAndSHA(targetName, trustServer, tlscacert, trustDir, timeout string) (*client.TargetWithRole, string, error) {
	gun, err := docker.GetGUN(targetName)
	if err != nil {
		return nil, "", fmt.Errorf("cannot get repo and tag from reference: %v", err)
	}

	target, err := GetTargetWithRole(gun, targetName, trustServer, tlscacert, trustDir, timeout)
	if err != nil {
		return nil, "", err
	}

	trustedSHA := hex.EncodeToString(target.Hashes["sha256"])
	log.Infof("Pulled trust data for %v, with role %v - SHA256: %v", targetName, target.Role, trustedSHA)
	return target, trustedSHA, nil
}

// GetThickBundle reads the thick bundle from disk
func GetThickBundle(localFile string) ([]byte, error) {
	log.Infof("Reading thick bundle on disk: %v", localFile)
	return ioutil.ReadFile(localFile)
}

// GetThinBundle reads the thin bundle from the OCI registry
func GetThinBundle(ref string) ([]byte, error) {
	log.Infof("Pulling thin bundle from registry: %v", ref)
	bun, err := cnab.Pull(ref)
	if err != nil {
		return nil, err
	}
	return json.MarshalCanonical(bun)
}

// VerifyTrust ensures the trust metadata for a given GUN matches the metadata of the pushed bundle
func VerifyTrust(buf []byte, trustedSHA string) error {
	err := verifyTargetSHAFromBytes(buf, trustedSHA)
	if err == nil {
		log.Infof("The SHA sums are equal: %v\n", trustedSHA)
	}
	return err
}

func verifyTargetSHAFromBytes(buf []byte, trustedSHA string) error {
	hasher := sha256.New()
	_, err := hasher.Write(buf)
	if err != nil {
		return err
	}
	computedSHA := hex.EncodeToString(hasher.Sum(nil))

	log.Infof("Computed SHA: %v\n", computedSHA)
	if trustedSHA != computedSHA {
		return fmt.Errorf("the digest sum of the artifact from the trusted collection %v is not equal to the computed digest %v",
			trustedSHA, computedSHA)
	}
	return nil
}
