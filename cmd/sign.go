package main

import (
	"encoding/hex"
	"fmt"

	canonicaljson "github.com/docker/go/canonical/json"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/engineerd/signy/pkg/cnab"
	"github.com/engineerd/signy/pkg/intoto"
	"github.com/engineerd/signy/pkg/tuf"
)

type signCmd struct {
	ref     string
	thick   bool
	file    string
	rootKey string

	intoto bool
	layout string
	// TODO: figure out a way to pass layout root key to TUF (not in the custom object)
	layoutKey string
	linkDir   string
}

func newSignCmd() *cobra.Command {
	const signDesc = `
Pushes the metadata of a CNAB bundle to a trust collection on a remote server. If the artifact is a thin bundle, this command
also pushes it to a repository using CNAB-TO-OCI.
On the first push to a repository, this command generates the signing keys.
To avoid introducing the passphrases every time, set the following environment variables with the corresponding passphrases:

export SIGNY_ROOT_PASSPHRASE
export SIGNY_TARGETS_PASSPHRASE
export SIGNY_SNAPSHOT_PASSPHRASE
export SIGNY_DELEGATION_PASSPHRASE

For more info on managing the signing keys, see https://docs.docker.com/notary/advanced_usage/

Example: computes the SHA256 digest of a canonical CNAB bundle, pushes it to the trust server, then pushes the bundle using CNAB-TO-OCI

$ signy sign bundle.json docker.io/<user>/<repo>:<tag>
Root key found, using: d701ba005e6d217c7eb6cb56dbc6cf0bd81f41347927acbca1318131cc693fc9

Pushed trust data for docker.io/<user>/<repo>:<tag>: 607ddb1d998e2155104067f99065659b202b0b19fa9ae52349ba3e9248635475
Starting to copy image cnab/helloworld:0.1.1...
Completed image cnab/helloworld:0.1.1 copy

Generated relocation map: bundle.ImageRelocationMap{"cnab/helloworld:0.1.1":"docker.io/radumatei/signed-cnab-bundle@sha256:a59a4e74d9cc89e4e75dfb2cc7ea5c108e4236ba6231b53081a9e2506d1197b6"}
Pushed successfully, with digest "sha256:086ef83113475d4582a7431b4b9bc98634d4f71ad1289cca45e661153fc9a46e"

Example: computes the SHA256 digest of a thick bundle, pushes it to a trust sever

$ signy --tlscacert=$NOTARY_CA --server https://localhost:4443 sign helloworld-0.1.1.tgz --thick  localhost:5000/thick-bundle-signature:v1

You are about to create a new root signing key passphrase. This passphrase
will be used to protect the most sensitive key in your signing system. Please
choose a long, complex passphrase and be careful to keep the password and the
key file itself secure and backed up. It is highly recommended that you use a
password manager to generate the passphrase and keep it safe. There will be no
way to recover this key. You can find the key in your config directory.

Enter passphrase for new root key with ID d701ba0:
Repeat passphrase for new root key with ID d701ba0:
Enter passphrase for new targets key with ID 5113934:
Repeat passphrase for new targets key with ID 5113934:
Enter passphrase for new snapshot key with ID d12e8e4:
Repeat passphrase for new snapshot key with ID d12e8e4:

Pushed trust data for localhost:5000/thick-bundle-signature:v1: cd205919129bff138a3402b4de5abbbc1d310ec982e83a780ffee1879adda678
`
	sign := signCmd{}
	cmd := &cobra.Command{
		Use:   "sign [file] [target reference]",
		Short: "Signs an artifact",
		Long:  signDesc,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			sign.file = args[0]
			sign.ref = args[1]
			return sign.run()
		},
	}
	cmd.Flags().StringVarP(&sign.rootKey, "root-key", "", "", "Root key to initialize the repository with")
	cmd.Flags().BoolVarP(&sign.thick, "thick", "", false, "Signs a thick bundle. If passed, only the signature is pushed to the trust server, not the bundle file")

	cmd.Flags().BoolVarP(&sign.intoto, "in-toto", "", false, "Adds in-toto metadata to TUF. If passed, the root layout, links directory, and root kyes must be supplied")
	cmd.Flags().StringVarP(&sign.layout, "layout", "", "", "Path to the in-toto root layout file")
	cmd.Flags().StringVarP(&sign.linkDir, "links", "", "", "Path to the in-toto links directory")
	cmd.Flags().StringVarP(&sign.layoutKey, "layout-key", "", "", "Path to the in-toto root layout public keys")

	return cmd
}

func (s *signCmd) run() error {
	var cm *canonicaljson.RawMessage
	if s.intoto {
		log.Infof("Adding In-Toto layout and links metadata to TUF")
		err := intoto.ValidateFromPath(s.layout)
		if err != nil {
			return fmt.Errorf("validation for in-toto metadata failed: %v", err)
		}
		custom, err := intoto.GetMetadataRawMessage(s.layout, s.linkDir, s.layoutKey)
		if err != nil {
			return fmt.Errorf("cannot get metadata message: %v", err)
		}
		// TODO: Radu M
		// Refactor GetMatedataRawMessage to return a pointer to a raw message
		cm = &custom
	}

	target, err := tuf.SignAndPublish(trustDir, trustServer, s.ref, s.file, tlscacert, s.rootKey, cm)
	if err != nil {
		return fmt.Errorf("cannot sign and publish trust data: %v", err)
	}

	log.Infof("Pushed trust data for %v: %v\n", s.ref, hex.EncodeToString(target.Hashes["sha256"]))

	if s.thick {
		return nil
	}

	return cnab.Push(s.file, s.ref)
}
