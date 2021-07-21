package tuf

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/cnabio/signy/pkg/intoto"
	"github.com/theupdateframework/notary/client"
	"github.com/theupdateframework/notary/trustpinning"
	"github.com/theupdateframework/notary/tuf/data"
)

// GetTargetWithRole returns a single target by name from the trusted collection
func GetTargetWithRole(gun, targetName, trustServer, tlscacert, trustDir, timeout string) (*client.TargetWithRole, error) {
	targets, err := GetTargets(gun, trustServer, tlscacert, trustDir, timeout)
	if err != nil {
		return nil, fmt.Errorf("cannot list targets:%v", err)
	}

	for _, target := range targets {
		if target.Name == targetName {
			return target, nil
		}
	}

	return nil, fmt.Errorf("cannot find target %v in trusted collection %v", targetName, gun)
}

// VerifyInToto ensures that the in-toto root layout, pubkeys, and links match the TUF metadata
func VerifyInTotoMetadata(gun, trustServer, tlscacert, trustDir, timeout string) error {
	// get targets from releases
	targets, err := GetTargets(gun, trustServer, tlscacert, trustDir, timeout, releasesRoleName)
	if err != nil {
		return fmt.Errorf("cannot list %v :%v", releasesRoleName, err)
	}

	// verify that releases signs ONLY links
	// the target paths entrusted to the delegatee
	paths := getReleasesPathPattern(gun)
	for _, target := range targets {
		if target.Role.String() != releasesRoleName.String() {
			return fmt.Errorf("expected %v but got :%v", releasesRoleName, target.Role.String())
		}
		// check the target name matches links
		match := false
		for _, path := range paths {
			if strings.HasPrefix(target.Name, path) {
				match = true
				break
			}
		}
		if !match {
			return fmt.Errorf("%v does not match %v", target.Name, paths)
		}
		// check the hash and length matches
		custom := intoto.Custom{}
		custom.ReadRawMessage(target.Custom)
		linkMeta, err := data.NewFileMeta(bytes.NewBuffer(custom.InToto.Data), data.NotaryDefaultHashes...)
		if err != nil {
			return err
		}
		if linkMeta.Length != target.Length {
			return fmt.Errorf("%v has observed length %v but expected %v", target.Name, linkMeta.Length, target.Length)
		}
		err = data.CompareMultiHashes(linkMeta.Hashes, target.Hashes)
		if err != nil {
			return fmt.Errorf("%v has observed hashes %v but expected %v", target.Name, linkMeta.Hashes, target.Hashes)
		}
	}

	// get targets from the top-level targets role
	targets, err = GetTargets(gun, trustServer, tlscacert, trustDir, timeout)
	if err != nil {
		return fmt.Errorf("cannot list %v :%v", data.CanonicalTargetsRole, err)
	}
	// TODO:  verify that targets signs ONLY layouts and pubkeys
	for _, target := range targets {
		if target.Role.String() != data.CanonicalTargetsRole.String() {
			return fmt.Errorf("expected %v but got :%v", data.CanonicalTargetsRole, target.Role.String())
		}
		// check the target name matches links
		match := false
		for _, path := range paths {
			if strings.HasPrefix(target.Name, path) {
				match = true
				break
			}
		}
		if !match {
			return fmt.Errorf("%v does not match %v", target.Name, paths)
		}

	}

	return nil
}

// GetTargets returns all targets for a given gun from the trusted collection
func GetTargets(gun, trustServer, tlscacert, trustDir, timeout string, roles ...data.RoleName) ([]*client.TargetWithRole, error) {
	if err := EnsureTrustDir(trustDir); err != nil {
		return nil, fmt.Errorf("cannot ensure trust directory: %v", err)
	}

	transport, err := makeTransport(trustServer, gun, tlscacert, timeout)
	if err != nil {
		return nil, fmt.Errorf("cannot make transport: %v", err)
	}

	repo, err := client.NewFileCachedRepository(
		trustDir,
		data.GUN(gun),
		trustServer,
		transport,
		nil,
		trustpinning.TrustPinConfig{},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create new file cached repository: %v", err)
	}

	return repo.ListTargets(roles...)
}

// PrintTargets prints all the targets for a specific GUN from a trust server
func PrintTargets(gun, trustServer, tlscacert, trustDir, timeout string) error {
	targets, err := GetTargets(gun, trustServer, tlscacert, trustDir, timeout)
	if err != nil {
		return fmt.Errorf("cannot list targets:%v", err)
	}

	for _, tgt := range targets {
		fmt.Printf("%s\t%s\n", tgt.Name, hex.EncodeToString(tgt.Hashes["sha256"]))
	}
	return nil
}
