package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/engineerd/signy/pkg/tuf"
)

type verifyCmd struct {
	ref       string
	thick     bool
	localFile string
}

func newVerifyCmd() *cobra.Command {
	const verifyDesc = `
Pulls the metadata for a target from a trusted collection and checks that the trusted digest
equals the digest of the existing artifact.
For CNAB, the bundle is pulled from the OCI registry, and the two digests are compared. Optionally, the digests
can also be validated against a local bundle.json file in canonical form, passed through the --local flag.

For plaintext artifacts, the target from the trusted collection must be validated against a local file passed
through the --local flag.

Example: verifies the metadata in the trusted collection for a CNAB bundle against the bundle pushed to an OCI registry

$ signy verify docker.io/<user>/<repo>:<tag>
Pulled trust data for docker.io/<user>/<repo>:<tag>, with role targets - SHA256: 607ddb1d998e2155104067f99065659b202b0b19fa9ae52349ba3e9248635475
Pulling bundle from registry: docker.io/<user>/<repo>:<tag>
Relocation map map[cnab/helloworld:0.1.1:radumatei/signed-cnab@sha256:a59a4e74d9cc89e4e75dfb2cc7ea5c108e4236ba6231b53081a9e2506d1197b6]

Computed SHA: 607ddb1d998e2155104067f99065659b202b0b19fa9ae52349ba3e9248635475
The SHA sums are equal: 607ddb1d998e2155104067f99065659b202b0b19fa9ae52349ba3e9248635475

Example: verifies the metadata for a local thick bundle:

$ signy --tlscacert=$NOTARY_CA --server https://localhost:4443 verify --thick --local helloworld-0.1.1.tgz localhost:5000/thick-bundle-signature:v1
Pulled trust data for localhost:5000/thick-bundle-signature:v1, with role targets - SHA256: cd205919129bff138a3402b4de5abbbc1d310ec982e83a780ffee1879adda678
Computed SHA: cd205919129bff138a3402b4de5abbbc1d310ec982e83a780ffee1879adda678
The SHA sums are equal: cd205919129bff138a3402b4de5abbbc1d310ec982e83a780ffee1879adda678
`
	verify := verifyCmd{}
	cmd := &cobra.Command{
		Use:   "verify [target reference]",
		Short: "Verifies the trust data for an artifact",
		Long:  verifyDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			verify.ref = args[0]
			return verify.run()
		},
	}
	cmd.Flags().BoolVarP(&verify.thick, "thick", "", false, "Verifies a thick bundle. If passed, only the signature is pulled from the trust server, and is verified against a local thick bundle.")
	cmd.Flags().StringVarP(&verify.localFile, "local", "", "", "Local file to validate the SHA256 against (mandatory for thick bundles).")

	return cmd
}

func (v *verifyCmd) run() error {
	if v.thick {
		if v.localFile == "" {
			return fmt.Errorf("no local file provided for thick bundle verification")
		}
		return tuf.VerifyFileTrust(v.ref, v.localFile, trustServer, tlscacert, trustDir)
	}

	return tuf.VerifyCNABTrust(v.ref, trustServer, tlscacert, trustDir)
}
