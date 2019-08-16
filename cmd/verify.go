package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/engineerd/signy/pkg/trust"
)

type verifyCmd struct {
	ref          string
	artifactType string
	localFile    string
}

func newVerifyCmd() *cobra.Command {
	const verifyDesc = `
Pulls the metadata for a target from a trusted collection and checks that the trusted digest
equals the digest of the existing artifact.
For CNAB, the bundle is pulled from the OCI registry, and the two digests are compared. Optionally, the digests 
can also be validated against a local bundle.json file in canonical form, passed through the --local flag.

For plaintext artifacts, the target from the trusted collection must be validated against a local file passed 
through the --local flag.

Example: verifies the metadata for a plaintext file (must be already pushed):

$ signy sign --type plaintext file.txt docker.io/<user>/<repo>:<tag>                                                             
Pushed trust data for docker.io/<user>/<repo>:<tag> : cf8916940c7f8b5eb747b9e056c32895176da9f0136033659929310540bef672
$ signy verify --type plaintext --local file.txt docker.io/<user>/<repo>:<tag>
Pulled trust data for docker.io/<user>/<repo>:<tag>, with role targets - SHA256: cf8916940c7f8b5eb747b9e056c32895176da9f0136033659929310540bef672
Computed SHA: cf8916940c7f8b5eb747b9e056c32895176da9f0136033659929310540bef672
The SHA sums are equal: cf8916940c7f8b5eb747b9e056c32895176da9f0136033659929310540bef672

Example: verifies the metadata in the trusted collection for a CNAB bundle against the bundle pushed to an OCI registry

$ signy verify --type cnab docker.io/<user>/<repo>:<tag>
Pulled trust data for docker.io/<user>/<repo>:<tag>, with role targets - SHA256: 607ddb1d998e2155104067f99065659b202b0b19fa9ae52349ba3e9248635475
Pulling bundle from registry: docker.io/<user>/<repo>:<tag>
Relocation map map[cnab/helloworld:0.1.1:radumatei/signed-cnab@sha256:a59a4e74d9cc89e4e75dfb2cc7ea5c108e4236ba6231b53081a9e2506d1197b6]

Computed SHA: 607ddb1d998e2155104067f99065659b202b0b19fa9ae52349ba3e9248635475
The SHA sums are equal: 607ddb1d998e2155104067f99065659b202b0b19fa9ae52349ba3e9248635475

Example: verifies the metadata in the trusted collection for a CNAB bundle against the bundle pushed to an OCI registry and against a local file

$ signy verify --type cnab --local bundle.json docker.io/<user>/<repo>:<tag>
Pulled trust data for docker.io/<user>/<repo>:<tag>, with role targets - SHA256: 607ddb1d998e2155104067f99065659b202b0b19fa9ae52349ba3e9248635475
Pulling bundle from registry: docker.io/<user>/<repo>:<tag>
Relocation map map[cnab/helloworld:0.1.1:radumatei/signed-cnab@sha256:a59a4e74d9cc89e4e75dfb2cc7ea5c108e4236ba6231b53081a9e2506d1197b6]

Computed SHA: 607ddb1d998e2155104067f99065659b202b0b19fa9ae52349ba3e9248635475
Computed SHA: 607ddb1d998e2155104067f99065659b202b0b19fa9ae52349ba3e9248635475
The SHA sums are equal: 607ddb1d998e2155104067f99065659b202b0b19fa9ae52349ba3e9248635475
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
	cmd.Flags().StringVarP(&verify.artifactType, "type", "", "plaintext", "Type of the artifact.")
	cmd.Flags().StringVarP(&verify.localFile, "local", "", "", "Local file to validate the SHA256 against (mandatory for plaintext artifacts).")

	return cmd
}

func (v *verifyCmd) run() error {
	switch v.artifactType {
	case "plaintext":
		if v.localFile == "" {
			return fmt.Errorf("no local file provided for plain text verification")
		}
		return trust.VerifyPlainTextTrust(v.ref, v.localFile, trustServer, tlscacert, trustDir)
	case "cnab":
		_, err := trust.VerifyCNABTrust(v.ref, v.localFile, trustServer, tlscacert, trustDir)
		return err
	default:
		return fmt.Errorf("unknown type")
	}
}
