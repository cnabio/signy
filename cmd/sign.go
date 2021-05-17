package main

import (
	"encoding/hex"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/cnabio/signy/pkg/canonical_json"
	"github.com/cnabio/signy/pkg/cnab"
	"github.com/cnabio/signy/pkg/intoto"
	"github.com/cnabio/signy/pkg/tuf"
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
export SIGNY_RELEASES_PASSPHRASE

For more info on managing the signing keys, see https://docs.docker.com/notary/advanced_usage/

Example: computes the SHA256 digest of a canonical CNAB bundle, pushes it to the trust server, then pushes the bundle using CNAB-TO-OCI

$ signy --tlscacert=$NOTARY_CA --server https://localhost:4443 sign testdata/cnab/bundle.json localhost:5000/thin-bundle:v1
INFO[0000] Pushed trust data for localhost:5000/thin-bundle:v1: c7e92bd51f059d60b15ad456edf194648997d739f60799b37e08edafd88a81b5
INFO[0000] Starting to copy image cnab/helloworld:0.1.1
INFO[0002] Completed image cnab/helloworld:0.1.1 copy
INFO[0002] Generated relocation map: relocation.ImageRelocationMap{"cnab/helloworld:0.1.1":"localhost:5000/thin-bundle@sha256:a59a4e74d9cc89e4e75dfb2cc7ea5c108e4236ba6231b53081a9e2506d1197b6"}
INFO[0002] Pushed successfully, with digest "sha256:b4936e42304c184bafc9b06dde9ea1f979129e09a021a8f40abc07f736de9268"

Example: computes the SHA256 digest of a thick bundle, pushes it to a trust sever

$ signy --tlscacert=$NOTARY_CA --server https://localhost:4443 sign --thick testdata/cnab/helloworld-0.1.1.tgz localhost:5000/thick-bundle:v1
INFO[0000] Pushed trust data for localhost:5000/thick-bundle:v1: 540cc4dc213548ebbdffb2ab0ef58729e089d1887edbcde6eeca851de624da70

In order to also push in-toto metadata to the TUF collection, use the --in-toto flag, together with --layout, --links, and (temporarily?) --layout-key.

Example:

$ signy --tlscacert=$NOTARY_CA --server https://localhost:4443 sign testdata/cnab/bundle.json localhost:5000/thin-intoto:v2 --in-toto --layout testdata/intoto/root.layout --links testdata/intoto --layout-key testdata/intoto/alice.pub
INFO[0000] Adding In-Toto layout and links metadata to TUF
INFO[0000] Pushed trust data for localhost:5000/thin-intoto:v2: c7e92bd51f059d60b15ad456edf194648997d739f60799b37e08edafd88a81b5
INFO[0000] Starting to copy image cnab/helloworld:0.1.1
INFO[0001] Completed image cnab/helloworld:0.1.1 copy
INFO[0001] Generated relocation map: relocation.ImageRelocationMap{"cnab/helloworld:0.1.1":"localhost:5000/thin-intoto@sha256:a59a4e74d9cc89e4e75dfb2cc7ea5c108e4236ba6231b53081a9e2506d1197b6"}
INFO[0001] Pushed successfully, with digest "sha256:b4936e42304c184bafc9b06dde9ea1f979129e09a021a8f40abc07f736de9268"
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
	var cm *canonical_json.RawMessage
	if s.intoto {
		if s.layout == "" || s.layoutKey == "" || s.linkDir == "" {
			return fmt.Errorf("required in-toto metadata not found")
		}
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

	// NOTE: We first push to the Registry, and then Notary. This is so that if we modify the bundle locally,
	// we will not invalidate its signature by first pushing to Notary, and then the Registry.

	// We push only thin bundles to the Registry.
	if !s.thick {
		if err := cnab.Push(s.file, s.ref); err != nil {
			return err
		}
	}

	target, err := tuf.SignAndPublish(trustDir, trustServer, s.ref, s.file, tlscacert, s.rootKey, timeout, cm)
	if err != nil {
		return fmt.Errorf("cannot sign and publish trust data: %v", err)
	}

	log.Infof("Pushed trust data for %v: %v\n", s.ref, hex.EncodeToString(target.Hashes["sha256"]))
	return nil
}
