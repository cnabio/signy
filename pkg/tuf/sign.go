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
	if err := ensureTrustDir(trustDir); err != nil {
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
			rootKeyIDs, err := importRootKey(rootKey, repo, getPassphraseRetriever())
			if err != nil {
				return nil, err
			}

			// 2nd variadic argument is to indicate that snapshot is managed remotely.
			if err = repo.Initialize(rootKeyIDs, data.CanonicalSnapshotRole); err != nil {
				return nil, fmt.Errorf("cannot initialize repo: %v", err)
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
