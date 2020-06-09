// Most of the helper functions are adapted from github.com/theupdateframework/notary
//
// Figure out the proper way of making sure we are respecting the licensing from Notary
// While we are also vendoring Notary directly (see LICENSE in vendor/github.com/theupdateframework/notary/LICENSE),
// copying unexported functions could fall under different licensing, so we need to make sure.

package tuf

import (
	"fmt"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/theupdateframework/notary"
	"github.com/theupdateframework/notary/client"
	"github.com/theupdateframework/notary/cryptoservice"
	"github.com/theupdateframework/notary/passphrase"
	"github.com/theupdateframework/notary/trustmanager"
	"github.com/theupdateframework/notary/tuf/data"
	"github.com/theupdateframework/notary/tuf/utils"
)

func getPassphraseRetriever() notary.PassRetriever {
	baseRetriever := passphrase.PromptRetriever()
	env := map[string]string{
		"root":             os.Getenv("SIGNY_ROOT_PASSPHRASE"),
		"targets":          os.Getenv("SIGNY_TARGETS_PASSPHRASE"),
		"targets/releases": os.Getenv("SIGNY_RELEASES_PASSPHRASE"),
	}

	return func(keyName string, alias string, createNew bool, numAttempts int) (string, bool, error) {
		if v := env[alias]; v != "" {
			return v, numAttempts > 1, nil
		}
		return baseRetriever(keyName, alias, createNew, numAttempts)
	}
}

// Attempt to read a role key from a file, and return it as a data.PrivateKey
// If key is for the Root role, it must be encrypted
func readPrivateKey(role data.RoleName, keyFilename string, retriever notary.PassRetriever) (data.PrivateKey, error) {
	pemBytes, err := ioutil.ReadFile(keyFilename)
	if err != nil {
		return nil, fmt.Errorf("Error reading input root key file: %v", err)
	}

	isEncrypted := true
	if err = cryptoservice.CheckRootKeyIsEncrypted(pemBytes); err != nil {
		if role == data.CanonicalRootRole {
			return nil, err
		}
		isEncrypted = false
	}

	var privKey data.PrivateKey
	if isEncrypted {
		privKey, _, err = trustmanager.GetPasswdDecryptBytes(retriever, pemBytes, "", data.CanonicalRootRole.String())
	} else {
		privKey, err = utils.ParsePEMPrivateKey(pemBytes, "")
	}
	if err != nil {
		return nil, err
	}

	return privKey, nil
}

// importRootKey imports the root key from path then adds the key to repo
// returns key ids
// https://github.com/theupdateframework/notary/blob/f255ae779066dc28ae4aee196061e58bb38a2b49/cmd/notary/tuf.go#L413
func importRootKey(rootKey string, nRepo client.Repository, retriever notary.PassRetriever) ([]string, error) {
	var rootKeyList []string

	if rootKey != "" {
		privKey, err := readPrivateKey(data.CanonicalRootRole, rootKey, retriever)
		if err != nil {
			return nil, err
		}
		// add root key to repo
		err = nRepo.GetCryptoService().AddKey(data.CanonicalRootRole, "", privKey)
		if err != nil {
			return nil, fmt.Errorf("Error importing key: %v", err)
		}
		rootKeyList = []string{privKey.ID()}
	} else {
		rootKeyList = nRepo.GetCryptoService().ListKeys(data.CanonicalRootRole)
	}

	if len(rootKeyList) > 0 {
		// Chooses the first root key available, which is initialization specific
		// but should return the HW one first.
		rootKeyID := rootKeyList[0]
		log.Debugf("found root key: %s\n", rootKeyID)
		return []string{rootKeyID}, nil
	}

	return []string{}, nil
}

// Try to reuse a single targets/releases key across repositories.
func reuseReleasesKey(r client.Repository) (data.PublicKey, error) {
	// Get all known targets keys.
	cryptoService := r.GetCryptoService()
	keyList := cryptoService.ListKeys(releasesRoleName)

	// Try to extract a single targets/releases key we can reuse.
	switch len(keyList) {
	case 0:
		log.Debugf("No %s key available, need to make one", releasesRoleName)
		return cryptoService.Create(releasesRoleName, r.GetGUN(), data.ECDSAKey)
	case 1:
		log.Debugf("Nothing to do, only one %s key available", releasesRoleName)
		return cryptoService.GetKey(keyList[0]), nil
	default:
		return nil, fmt.Errorf("there is more than one %s keys", releasesRoleName)
	}
}

// Try to reuse a single targets key across repositories.
// FIXME: Unfortunately, short of forking Notary or sending a PR upstream, there isn't an easy way to prevent it
// from automagically creating a new, local targets key per TUF metadata repository. We fix this here by undoing
// more than one new, local targets key, and reusing any existing local targets key, just like the way Notary
// reuses the root key.
func reuseTargetsKey(r client.Repository) error {
	// Get all known targets keys.
	keyList := r.GetCryptoService().ListKeys(data.CanonicalTargetsRole)

	// Try to extract a single targets key we can reuse.
	switch len(keyList) {
	case 0:
		return fmt.Errorf("no targets key despite having initialized a repo")
	case 1:
		log.Debug("Nothing to do, only one targets key available")
		return nil
	case 2:
		// First, we publish current changes to repository in order to list roles.
		// FIXME: Find a find better way to list roles w/o publishing changes first.
		err := r.Publish()
		if err != nil {
			return err
		}

		// Get the current top-level roles.
		roleWithSigs, err := r.ListRoles()
		if err != nil {
			return err
		}

		// Get the current targets key.
		// NOTE: We do not delete it, in case the user wants to keep it.
		var thisKeyID string
		for _, roleWithSig := range roleWithSigs {
			role := roleWithSig.Role
			if role.Name == data.CanonicalTargetsRole {
				if len(role.KeyIDs) == 1 {
					thisKeyID = role.KeyIDs[0]
					log.Debugf("This targets keyid: %s", thisKeyID)
				} else {
					return fmt.Errorf("this targets role has more than 1 key")
				}
			}
		}

		// Get and reuse the other targets key.
		var thatKeyID string
		for _, keyID := range keyList {
			if keyID != thisKeyID {
				thatKeyID = keyID
				break
			}
		}
		log.Debugf("That targets keyID: %s", thatKeyID)
		log.Debugf("Before rotating targets key from %s to %s", thisKeyID, thatKeyID)
		err = r.RotateKey(data.CanonicalTargetsRole, false, []string{thatKeyID})
		log.Debugf("After targets key rotation")
		return err
	default:
		return fmt.Errorf("there are more than two targets keys")
	}
}
