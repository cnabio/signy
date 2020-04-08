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
		"root":     os.Getenv("SIGNY_ROOT_PASSPHRASE"),
		"targets":  os.Getenv("SIGNY_TARGETS_PASSPHRASE"),
		"releases": os.Getenv("SIGNY_RELEASES_PASSPHRASE"),
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
func readKey(role data.RoleName, keyFilename string, retriever notary.PassRetriever) (data.PrivateKey, error) {
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
		privKey, err := readKey(data.CanonicalRootRole, rootKey, retriever)
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
		log.Debugf("Signy found root key, using: %s\n", rootKeyID)

		return []string{rootKeyID}, nil
	}

	return []string{}, nil
}

// Try to reuse a single targets key across repositories.
// FIXME: Unfortunately, short of forking Notary or sending a PR upstream, there isn't an easy way to prevent it
// from automagically creating a new, local targets key per TUF metadata repository. We fix this here by undoing
// more than one new, local targets key, and reusing any existing local targets key, just like the way Notary
// reuses the root key.
func reuseTargetsKey(r client.Repository) error {
	var (
		err                                error
		thisTargetsKeyID, thatTargetsKeyID string
	)

	// Get all known targets keys.
	targetsKeyList := r.GetCryptoService().ListKeys(data.CanonicalTargetsRole)
	// Try to extract a single targets key we can reuse.
	switch len(targetsKeyList) {
	case 0:
		err = fmt.Errorf("no targets key despite having initialized a repo")
	case 1:
		log.Debug("Nothing to do, only one targets key available")
	case 2:
		// First, we publish current changes to repository in order to list roles.
		// FIXME: Find a find better way to list roles w/o publishing changes first.
		publishErr := r.Publish()
		if publishErr != nil {
			err = publishErr
			break
		}

		// Get the current top-level roles.
		roleWithSigs, listRolesErr := r.ListRoles()
		if listRolesErr != nil {
			err = listRolesErr
			break
		}

		// Get the current targets key.
		// NOTE: We do not delete it, in case the user wants to keep it.
		for _, roleWithSig := range roleWithSigs {
			role := roleWithSig.Role
			if role.Name == data.CanonicalTargetsRole {
				if len(role.KeyIDs) == 1 {
					thisTargetsKeyID = role.KeyIDs[0]
					log.Debugf("This targets keyid: %s", thisTargetsKeyID)
				} else {
					return fmt.Errorf("this targets role has more than 1 key")
				}
			}
		}

		// Get and reuse the other targets key.
		for _, keyID := range targetsKeyList {
			if keyID != thisTargetsKeyID {
				thatTargetsKeyID = keyID
				break
			}
		}
		log.Debugf("That targets keyID: %s", thatTargetsKeyID)
		log.Debugf("Before rotating targets key from %s to %s", thisTargetsKeyID, thatTargetsKeyID)
		err = r.RotateKey(data.CanonicalTargetsRole, false, []string{thatTargetsKeyID})
		log.Debugf("After targets key rotation")
	default:
		err = fmt.Errorf("there are more than 2 targets keys")
	}

	return err
}
