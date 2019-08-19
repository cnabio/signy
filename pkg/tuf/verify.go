package tuf

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"

	"github.com/docker/go/canonical/json"
	"github.com/theupdateframework/notary/client"

	"github.com/engineerd/signy/pkg/cnab"
)

// VerifyCNABTrust ensures the trust metadata for a given GUN matches the metadata of the pushed bundle
func VerifyCNABTrust(ref, localFile, trustServer, tlscacert, trustDir string) (*client.TargetWithRole, error) {
	gun, name := cnab.SplitTargetRef(ref)
	target, err := GetTargetWithRole(gun, name, trustServer, tlscacert, trustDir)
	if err != nil {
		return nil, err
	}

	trustedSHA := hex.EncodeToString(target.Hashes["sha256"])
	fmt.Printf("Pulled trust data for %v, with role %v - SHA256: %v", ref, target.Role, trustedSHA)

	fmt.Printf("\nPulling bundle from registry: %v", ref)
	bun, err := cnab.Pull(ref)
	if err != nil {
		return nil, fmt.Errorf("cannot pull bundle: %v", err)
	}
	buf, err := json.MarshalCanonical(bun)
	if err != nil {
		return nil, err
	}

	remotesErr := verifyTargetSHAFromBytes(target, buf)
	if localFile == "" {
		if remotesErr == nil {
			fmt.Printf("\nThe SHA sums are equal: %v\n", trustedSHA)
		}
		return target, remotesErr
	}

	lb, err := ioutil.ReadFile(localFile)
	if err != nil {
		return nil, err
	}

	if err := verifyTargetSHAFromBytes(target, lb); err == nil && remotesErr == nil {
		fmt.Printf("\nThe SHA sums are equal: %v\n", trustedSHA)
		return target, nil
	}
	return target, err
}

// VerifyPlainTextTrust ensures the trust metadata for a given GUN matches the computed metadata of the local file
func VerifyPlainTextTrust(ref, localFile, trustServer, tlscacert, trustDir string) error {
	gun, name := cnab.SplitTargetRef(ref)
	target, err := GetTargetWithRole(gun, name, trustServer, tlscacert, trustDir)
	if err != nil {
		return err
	}

	trustedSHA := hex.EncodeToString(target.Hashes["sha256"])
	fmt.Printf("Pulled trust data for %v, with role %v - SHA256: %v", ref, target.Role, trustedSHA)

	buf, err := ioutil.ReadFile(localFile)
	if err != nil {
		return err
	}

	err = verifyTargetSHAFromBytes(target, buf)
	if err != nil {
		return err
	}
	fmt.Printf("\nThe SHA sums are equal: %v\n", trustedSHA)
	return nil
}

func verifyTargetSHAFromBytes(target *client.TargetWithRole, buf []byte) error {
	trustedSHA := hex.EncodeToString(target.Hashes["sha256"])
	hasher := sha256.New()
	hasher.Write(buf)
	computedSHA := hex.EncodeToString(hasher.Sum(nil))
	fmt.Printf("\nComputed SHA: %v", computedSHA)
	if trustedSHA != computedSHA {
		return fmt.Errorf("the digest sum of the artifact from the trusted collection %v is not equal to the computed digest %v",
			trustedSHA, computedSHA)
	}

	return nil
}
