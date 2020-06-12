package tuf

import (
	"fmt"

	canonicaljson "github.com/docker/go/canonical/json"
	"github.com/theupdateframework/notary/client"
	"github.com/theupdateframework/notary/trustpinning"
	"github.com/theupdateframework/notary/tuf/data"
)

// SignAndPublish signs an artifact, then publishes the metadata to a trust server
func SignAndPublish(trustDir, trustServer, ref, file, tlscacert, rootKey, timeout string, custom *canonicaljson.RawMessage) (*client.Target, error) {
	if err := EnsureTrustDir(trustDir); err != nil {
		return nil, fmt.Errorf("cannot ensure trust directory: %v", err)
	}

	gun, err := getGUN(ref)
	if err != nil {
		return nil, fmt.Errorf("cannot get GUN reference: %v", err)
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
		getPassphraseRetriever(),
		trustpinning.TrustPinConfig{},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create new file cached repository: %v", err)
	}

	err = clearChangeList(repo)
	if err != nil {
		return nil, fmt.Errorf("cannot clear change list: %v", err)
	}
	defer clearChangeList(repo)

	err = reuseKeys(repo, rootKey)
	if err != nil {
		return nil, fmt.Errorf("cannot reuse keys: %v", err)
	}

	// NOTE: We use the full reference for the target filename of the bundle.
	target, err := client.NewTarget(ref, file, custom)
	if err != nil {
		return nil, err
	}

	// NOTE: And we add the bundle to the "targets/releases" instead of the top-level "targets" role.
	if err = repo.AddTarget(target, releasesRoleName); err != nil {
		return nil, err
	}

	err = repo.Publish()
	return target, err
}

// clearChangelist clears the notary staging changelist
func clearChangeList(notaryRepo client.Repository) error {
	cl, err := notaryRepo.GetChangelist()
	if err != nil {
		return err
	}
	return cl.Clear("")
}

// reuse root and top-level targets keys
func reuseKeys(repo client.Repository, rootKey string) error {
	if _, err := repo.ListTargets(); err != nil {
		switch err.(type) {
		case client.ErrRepoNotInitialized, client.ErrRepositoryNotExist:
			// Reuse root key.
			rootKeyIDs, err := importRootKey(rootKey, repo, getPassphraseRetriever())
			if err != nil {
				return err
			}

			// NOTE: 2nd variadic argument is to indicate that snapshot is managed remotely.
			// The impact of a timestamp + snapshot key compromise is not terrible:
			// https://docs.docker.com/notary/service_architecture/#threat-model
			if err = repo.Initialize(rootKeyIDs, data.CanonicalSnapshotRole); err != nil {
				return fmt.Errorf("cannot initialize repo: %v", err)
			}

			// Reuse targets key.
			if err = reuseTargetsKey(repo); err != nil {
				return fmt.Errorf("cannot reuse %s keys: %v", data.CanonicalTargetsRole, err)
			}

			// Reuse targets/releases key.
			releasesPublicKey, err := reuseReleasesKey(repo)
			if err != nil {
				return fmt.Errorf("cannot reuse %s keys: %v", releasesRoleName, err)
			}

			// Delegate to targets/releases.
			err = delegateToReleases(repo, releasesPublicKey)
			if err != nil {
				return fmt.Errorf("cannot delegate to %s: %v", releasesRoleName, err)
			}

		default:
			return fmt.Errorf("cannot list targets: %v", err)
		}
	}
	return nil
}
