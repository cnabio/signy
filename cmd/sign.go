package main

import (
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/engineerd/signy/pkg/cnab"
	"github.com/engineerd/signy/pkg/trust"
)

type signCmd struct {
	ref          string
	file         string
	artifactType string
	rootKey      string
}

func newSignCmd() *cobra.Command {
	const signDesc = `
Pushes the metadata of an artifact to a trust collection on a remote server, and based on the type, 
it pushes the actual artifact to a repository. Currently, the supported artifact types are plaintext and cnab.

On the first push to a repository, it also generates the signing keys.
To avoid introducing the passphrases every time, set the following environment variables with the corresponding passphrases:

export SIGNY_ROOT_PASSPHRASE
export SIGNY_TARGETS_PASSPHRASE
export SIGNY_SNAPSHOT_PASSPHRASE
export SIGNY_DELEGATION_PASSPHRASE

Example: computes the SHA256 digest of plaintext file, then pushes it to the trust server and display the computed SHA256 digest.

For more info on managing the signing keys, see https://docs.docker.com/notary/advanced_usage/

$ signy sign --type plaintext file.txt docker.io/<user>/<repo>:<tag>
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

Pushed trust data for docker.io/<user>/<repo>:tag: cf8916940c7f8b5eb747b9e056c32895176da9f0136033659929310540bef672


Example: computes the SHA256 digest of a canonical CNAB bundle, pushes it to the trust server, then pushes the bundle using CNAB-TO-OCI

$ signy sign --type cnab bundle.json docker.io/<user>/<repo>:<tag>
Root key found, using: d701ba005e6d217c7eb6cb56dbc6cf0bd81f41347927acbca1318131cc693fc9

Pushed trust data for docker.io/<user>/<repo>:<tag>: 607ddb1d998e2155104067f99065659b202b0b19fa9ae52349ba3e9248635475
Starting to copy image cnab/helloworld:0.1.1...
Completed image cnab/helloworld:0.1.1 copy

Generated relocation map: bundle.ImageRelocationMap{"cnab/helloworld:0.1.1":"docker.io/radumatei/signed-cnab-bundle@sha256:a59a4e74d9cc89e4e75dfb2cc7ea5c108e4236ba6231b53081a9e2506d1197b6"}
Pushed successfully, with digest "sha256:086ef83113475d4582a7431b4b9bc98634d4f71ad1289cca45e661153fc9a46e"
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
	cmd.Flags().StringVarP(&sign.artifactType, "type", "", "plaintext", "Type of the artifact")
	cmd.Flags().StringVarP(&sign.rootKey, "rootkey", "", "", "Root key to initialize the repository with")

	return cmd
}

func (s *signCmd) run() error {
	switch s.artifactType {
	case "plaintext":
		target, err := trust.SignAndPublish(trustDir, trustServer, s.ref, s.file, tlscacert, s.rootKey)
		fmt.Printf("\nPushed trust data for %v: %v \n", s.ref, hex.EncodeToString(target.Hashes["sha256"]))
		return err
	case "cnab":
		target, err := trust.SignAndPublish(trustDir, trustServer, s.ref, s.file, tlscacert, s.rootKey)
		if err != nil {
			return fmt.Errorf("cannot sign and publish trust data: %v", err)
		}
		fmt.Printf("\nPushed trust data for %v: %v\n", s.ref, hex.EncodeToString(target.Hashes["sha256"]))
		return cnab.Push(s.file, s.ref)
	default:
		return fmt.Errorf("unknown type")
	}
}
