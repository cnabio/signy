package tuf

import (
	"fmt"

	canonicaljson "github.com/docker/go/canonical/json"
	"github.com/theupdateframework/notary/client"
	"github.com/theupdateframework/notary/trustpinning"
	"github.com/theupdateframework/notary/tuf/data"
)

// clearChangelist clears the notary staging changelist
func clearChangeList(notaryRepo client.Repository) error {
	cl, err := notaryRepo.GetChangelist()
	if err != nil {
		return err
	}
	return cl.Clear("")
}

// SignAndPublish signs an artifact, then publishes the metadata to a trust server
func SignAndPublish(trustDir, trustServer, ref, file, tlscacert, rootKey, timeout string, custom *canonicaljson.RawMessage) (*client.Target, error) {
	if err := EnsureTrustDir(trustDir); err != nil {
		return nil, fmt.Errorf("cannot ensure trust directory: %v", err)
	}

	repoInfo, tag, err := getRepoAndTag(ref)
	if err != nil {
		return nil, fmt.Errorf("cannot get repo and tag from reference: %v", err)
	}

	transport, err := makeTransport(trustServer, repoInfo.Name.Name(), tlscacert, timeout)
	if err != nil {
		return nil, fmt.Errorf("cannot make transport: %v", err)
	}

	repo, err := client.NewFileCachedRepository(
		trustDir,
		data.GUN(repoInfo.Name.Name()),
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

	if _, err = repo.ListTargets(); err != nil {
		switch err.(type) {
		case client.ErrRepoNotInitialized, client.ErrRepositoryNotExist:
			// Reuse root key.
			rootKeyIDs, err := importRootKey(rootKey, repo, getPassphraseRetriever())
			if err != nil {
				return nil, err
			}

			// NOTE: 2nd variadic argument is to indicate that snapshot is managed remotely.
			// The impact of a timestamp + snapshot key compromise is not terrible:
			// https://docs.docker.com/notary/service_architecture/#threat-model
			if err = repo.Initialize(rootKeyIDs, data.CanonicalSnapshotRole); err != nil {
				return nil, fmt.Errorf("cannot initialize repo: %v", err)
			}

			// Reuse targets key.
			if err = reuseTargetsKey(repo); err != nil {
				return nil, fmt.Errorf("cannot reuse targets keys: %v", err)
			}

		default:
			return nil, fmt.Errorf("cannot list targets: %v", err)
		}
	}

	target, err := client.NewTarget(tag, file, custom)
	if err != nil {
		return nil, err
	}

	// TODO - Radu M
	// decide whether to allow actually passing roles as flags

	// If roles is empty, we default to adding to targets
	if err = repo.AddTarget(target, data.NewRoleList([]string{})...); err != nil {
		return nil, err
	}

	err = repo.Publish()
	return target, err
}
