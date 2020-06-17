package tuf

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/cnabio/signy/pkg/docker"
	"github.com/cnabio/signy/pkg/intoto"

	"github.com/theupdateframework/notary/client"
	"github.com/theupdateframework/notary/trustpinning"
	"github.com/theupdateframework/notary/tuf/data"
)

// SignAndPublish signs an artifact, then publishes the metadata to a trust server
func SignAndPublish(trustDir, trustServer, ref, file, tlscacert, rootKey, timeout string, rootLayout intoto.RootLayout, publicKeys intoto.PublicKeys, links intoto.Links, bundleCustom intoto.Custom) (string, error) {
	if err := EnsureTrustDir(trustDir); err != nil {
		return "", fmt.Errorf("cannot ensure trust directory: %v", err)
	}

	gun, err := docker.GetGUN(ref)
	if err != nil {
		return "", fmt.Errorf("cannot get GUN reference: %v", err)
	}

	transport, err := makeTransport(trustServer, gun, tlscacert, timeout)
	if err != nil {
		return "", fmt.Errorf("cannot make transport: %v", err)
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
		return "", fmt.Errorf("cannot create new file cached repository: %v", err)
	}

	err = clearChangeList(repo)
	if err != nil {
		return "", fmt.Errorf("cannot clear change list: %v", err)
	}
	defer clearChangeList(repo)

	err = reuseKeys(repo, rootKey)
	if err != nil {
		return "", fmt.Errorf("cannot reuse keys: %v", err)
	}

	err = addRootLayout(repo, rootLayout, publicKeys)
	if err != nil {
		return "", fmt.Errorf("cannot add root layout: %v", err)
	}

	bundleDigest, err := addBundle(repo, ref, file, links, bundleCustom)
	if err != nil {
		return "", fmt.Errorf("cannot add bundle: %v", err)
	}

	err = repo.Publish()
	return bundleDigest, err
}

func addBundle(repo client.Repository, ref, file string, links intoto.Links, bundleCustom intoto.Custom) (string, error) {
	for linkFilename, linkCustom := range links {
		linkMeta, err := data.NewFileMeta(bytes.NewBuffer(linkCustom.InToto.Data), data.NotaryDefaultHashes...)
		if err != nil {
			return "", fmt.Errorf("cannot get link meta: %v", err)
		}

		linkCustomRawMessage, err := linkCustom.GetRawMessage()
		if err != nil {
			return "", fmt.Errorf("cannot get raw message for link custom: %v", err)
		}

		linkTarget := client.Target{Name: linkFilename, Hashes: linkMeta.Hashes, Length: linkMeta.Length, Custom: &linkCustomRawMessage}
		if err = repo.AddTarget(&linkTarget, releasesRoleName); err != nil {
			return "", fmt.Errorf("cannot add link target: %v", err)
		}

	}

	bundleCustomRawMessage, err := bundleCustom.GetRawMessage()
	if err != nil {
		return "", fmt.Errorf("cannot get raw message for bundle custom: %v", err)
	}

	// NOTE: We use the full reference for the target filename of the bundle.
	bundleTarget, err := client.NewTarget(ref, file, &bundleCustomRawMessage)
	if err != nil {
		return "", fmt.Errorf("cannot get bundle target: %v", err)
	}

	// NOTE: And we add the bundle to the "targets/releases" instead of the top-level "targets" role.
	if err = repo.AddTarget(bundleTarget, releasesRoleName); err != nil {
		return "", fmt.Errorf("cannot add bundle target: %v", err)
	}

	return hex.EncodeToString(bundleTarget.Hashes["sha256"]), nil
}

func addRootLayout(repo client.Repository, rootLayout intoto.RootLayout, publicKeys intoto.PublicKeys) error {
	for publicKeyFilename, publicKeyCustom := range publicKeys {
		meta, err := data.NewFileMeta(bytes.NewBuffer(publicKeyCustom.InToto.Data), data.NotaryDefaultHashes...)
		if err != nil {
			return fmt.Errorf("cannot get public key meta: %v", err)
		}

		publicKeyCustomRawMessage, err := publicKeyCustom.GetRawMessage()
		if err != nil {
			return fmt.Errorf("cannot get raw message for public key custom: %v", err)
		}

		linkTarget := client.Target{Name: publicKeyFilename, Hashes: meta.Hashes, Length: meta.Length, Custom: &publicKeyCustomRawMessage}
		if err = repo.AddTarget(&linkTarget, data.CanonicalTargetsRole); err != nil {
			return fmt.Errorf("cannot add public key target: %v", err)
		}
	}

	rootLayoutMeta, err := data.NewFileMeta(bytes.NewBuffer(rootLayout.Custom.InToto.Data), data.NotaryDefaultHashes...)
	if err != nil {
		return fmt.Errorf("cannot get root layout meta: %v", err)
	}

	rootLayoutCustomRawMessage, err := rootLayout.Custom.GetRawMessage()
	if err != nil {
		return fmt.Errorf("cannot get raw message for root layout custom: %v", err)
	}

	linkTarget := client.Target{Name: rootLayout.Filename, Hashes: rootLayoutMeta.Hashes, Length: rootLayoutMeta.Length, Custom: &rootLayoutCustomRawMessage}
	if err = repo.AddTarget(&linkTarget, data.CanonicalTargetsRole); err != nil {
		return fmt.Errorf("cannot add root layout target: %v", err)
	}

	return nil
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
