package tuf

import (
	"encoding/hex"
	"fmt"

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
	// TODO: get targets from releases
	targets, err := GetTargets(gun, trustServer, tlscacert, trustDir, timeout, releasesRoleName)
	if err != nil {
		return fmt.Errorf("cannot list %v :%v", releasesRoleName, err)
	}
	// TODO:  verify that releases signs ONLY links
	for _, target := range targets {
		if target.Role.String() != releasesRoleName.String() {
			return fmt.Errorf("expected %v but got :%v", releasesRoleName, target.Role.String())
		}
		// TODO: check the target name matches links
		// TODO: check the hash and length matches
	}

	// TODO: get targets from targets
	targets, err = GetTargets(gun, trustServer, tlscacert, trustDir, timeout)
	if err != nil {
		return fmt.Errorf("cannot list targets: %v", err)
	}
	// TODO:  verify that targets signs ONLY layouts and pubkeys
	for _, target := range targets {
		if target.Name == targetName {
			return target, nil
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
