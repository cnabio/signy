package tuf

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"

	"github.com/docker/go/canonical/json"
	log "github.com/sirupsen/logrus"
	"github.com/theupdateframework/notary/client"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/signy/pkg/cnab"
)

// VerifyTrust ensures the trust metadata for a given GUN matches the metadata of the pushed bundle
func VerifyTrust(ref, localFile, trustServer, tlscacert, trustDir, timeout string) (*client.TargetWithRole, []byte, error) {
	var bun *bundle.Bundle
	var buf []byte

	target, trustedSHA, err := GetTargetAndSHA(ref, trustServer, tlscacert, trustDir, timeout)
	if err != nil {
		return target, buf, err
	}
	log.Infof("Pulled trust data for %v, with role %v - SHA256: %v", ref, target.Role, trustedSHA)

	if localFile == "" {
		log.Infof("Pulling bundle from registry: %v", ref)
		bun, err = cnab.Pull(ref)
		if err != nil {
			return target, buf, fmt.Errorf("cannot pull bundle: %v", err)
		}
		buf, err = json.MarshalCanonical(bun)
	} else {
		buf, err = ioutil.ReadFile(localFile)
	}
	if err != nil {
		return target, buf, err
	}

	err = verifyTargetSHAFromBytes(trustedSHA, buf)
	if err == nil {
		log.Infof("The SHA sums are equal: %v\n", trustedSHA)
	}

	return target, buf, err
}

func verifyTargetSHAFromBytes(trustedSHA string, buf []byte) error {
	hasher := sha256.New()
	hasher.Write(buf)
	computedSHA := hex.EncodeToString(hasher.Sum(nil))

	log.Infof("Computed SHA: %v\n", computedSHA)
	if trustedSHA != computedSHA {
		return fmt.Errorf("the digest sum of the artifact from the trusted collection %v is not equal to the computed digest %v",
			trustedSHA, computedSHA)
	}
	return nil
}

// GetTargetAndSHA returns the target with roles and the SHA256 of the target file
func GetTargetAndSHA(ref, trustServer, tlscacert, trustDir, timeout string) (*client.TargetWithRole, string, error) {
	repoInfo, tag, err := getRepoAndTag(ref)
	if err != nil {
		return nil, "", fmt.Errorf("cannot get repo and tag from reference: %v", err)
	}

	target, err := GetTargetWithRole(repoInfo.Name.Name(), tag, trustServer, tlscacert, trustDir, timeout)
	if err != nil {
		return nil, "", err
	}

	return target, hex.EncodeToString(target.Hashes["sha256"]), nil
}
