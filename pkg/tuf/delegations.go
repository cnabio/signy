package tuf

import (
	"fmt"

	"github.com/theupdateframework/notary/client"
	"github.com/theupdateframework/notary/tuf/data"
)

// Delegate all paths ("*") to targets/releases.
// https://github.com/theupdateframework/notary/blob/f255ae779066dc28ae4aee196061e58bb38a2b49/cmd/notary/delegations.go
func delegateToReleases(repo client.Repository, publicKey data.PublicKey) error {
	// How Notary v1 denotes "*""
	// https://github.com/theupdateframework/notary/blob/f255ae779066dc28ae4aee196061e58bb38a2b49/cmd/notary/delegations.go#L367
	allPaths := []string{""}
	publicKeys := []data.PublicKey{publicKey}

	// Add the delegation to the repository
	err := repo.AddDelegation(releasesRoleName, publicKeys, allPaths)
	if err != nil {
		return fmt.Errorf("failed to create delegation: %v", err)
	}

	return nil
}
