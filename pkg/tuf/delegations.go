package tuf

import (
	"fmt"

	"github.com/theupdateframework/notary/client"
	"github.com/theupdateframework/notary/tuf/data"
)

// Delegate all paths ("*") to targets/releases.
// https://github.com/theupdateframework/notary/blob/f255ae779066dc28ae4aee196061e58bb38a2b49/cmd/notary/delegations.go
func delegateToReleases(repo client.Repository, publicKey data.PublicKey) error {
	// the public keys used to verify the delegatee
	publicKeys := []data.PublicKey{publicKey}
	// the target paths entrusted to the delegatee
	paths := make([]string, 2)
	gun := repo.GetGUN().String()
	tags := gun + ":"
	links := gun + "/in-toto-links/"
	paths = append(paths, tags, links)

	// Add the delegation to the repository
	err := repo.AddDelegation(releasesRoleName, publicKeys, paths)
	if err != nil {
		return fmt.Errorf("failed to create delegation: %v", err)
	}

	return nil
}
