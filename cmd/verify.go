package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/cnabio/signy/pkg/docker"
	"github.com/cnabio/signy/pkg/intoto"
	"github.com/cnabio/signy/pkg/tuf"
)

type verifyCmd struct {
	ref       string
	thick     bool
	localFile string

	intoto            bool
	verifyOnOS        bool
	verificationImage string
}

func newVerifyCmd() *cobra.Command {
	const verifyDesc = `
Pulls the metadata for a target from a trusted collection and checks that the trusted digest
equals the digest of the existing artifact.
For canonical CNAB bundes, the bundle is pulled from the OCI registry, and the two digests are compared.

For thick bundles, the --thick flag is required, together with the --local <path-to-thick-bundle>.

Example: verifies the metadata in the trusted collection for a CNAB bundle against the bundle pushed to an OCI registry

$ signy --tlscacert=$NOTARY_CA --server https://localhost:4443 verify localhost:5000/thin-bundle:v1
INFO[0000] Pulled trust data for localhost:5000/thin-bundle:v1, with role targets - SHA256: c7e92bd51f059d60b15ad456edf194648997d739f60799b37e08edafd88a81b5
INFO[0000] Pulling bundle from registry: localhost:5000/thin-bundle:v1
INFO[0000] Computed SHA: c7e92bd51f059d60b15ad456edf194648997d739f60799b37e08edafd88a81b5
INFO[0000] The SHA sums are equal: c7e92bd51f059d60b15ad456edf194648997d739f60799b37e08edafd88a81b5

Example: verifies the metadata for a local thick bundle:

$ signy --tlscacert=$NOTARY_CA --server https://localhost:4443 verify --thick --local testdata/cnab/helloworld-0.1.1.tgz localhost:5000/thick-bundle:v1
INFO[0000] Pulled trust data for localhost:5000/thick-bundle:v1, with role targets - SHA256: 540cc4dc213548ebbdffb2ab0ef58729e089d1887edbcde6eeca851de624da70
INFO[0000] Computed SHA: 540cc4dc213548ebbdffb2ab0ef58729e089d1887edbcde6eeca851de624da70
INFO[0000] The SHA sums are equal: 540cc4dc213548ebbdffb2ab0ef58729e089d1887edbcde6eeca851de624da70

In order to also verify  in-toto metadata from the TUF collection, use the --in-toto flag (and, if the verification requires, --target, to indicate target files used by the verification).

Example:

$ signy --tlscacert=$NOTARY_CA --server https://localhost:4443 verify localhost:5000/thin-intoto:v2 --in-toto
INFO[0000] Pulled trust data for localhost:5000/thin-intoto:v2, with role targets - SHA256: c7e92bd51f059d60b15ad456edf194648997d739f60799b37e08edafd88a81b5
INFO[0000] Pulling bundle from registry: localhost:5000/thin-intoto:v2
INFO[0000] Computed SHA: c7e92bd51f059d60b15ad456edf194648997d739f60799b37e08edafd88a81b5
INFO[0000] The SHA sums are equal: c7e92bd51f059d60b15ad456edf194648997d739f60799b37e08edafd88a81b5
INFO[0000] Writing In-Toto metadata files into /tmp/intoto-verification169227773
INFO[0000] copying file /in-toto/layout.template in container for verification...
INFO[0000] Loading layout...
INFO[0000] Loading layout key(s)...
INFO[0001] The software product passed all verification.
`
	verify := verifyCmd{}
	cmd := &cobra.Command{
		Use:   "verify [target reference]",
		Short: "Verifies the trust data for an artifact",
		Long:  verifyDesc,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			verify.ref = args[0]
			return verify.run()
		},
	}
	cmd.Flags().BoolVarP(&verify.thick, "thick", "", false, "Verifies a thick bundle. If passed, only the signature is pulled from the trust server, and is verified against a local thick bundle")
	cmd.Flags().StringVarP(&verify.localFile, "local", "", "", "Local file to validate the SHA256 against (mandatory for thick bundles)")

	cmd.Flags().BoolVarP(&verify.intoto, "in-toto", "", false, "If passed, will try to fetch in-toto metadata from TUF and perform the verification")
	cmd.Flags().BoolVarP(&verify.verifyOnOS, "verify-on-os", "", false, "If passed, will run in-toto inspections on the OS instead of in container")
	cmd.Flags().StringVarP(&verify.verificationImage, "image", "", docker.VerificationImage, "container image to run the in-toto verification")

	return cmd
}

func (v *verifyCmd) run() error {
	if v.thick && v.localFile == "" {
		return fmt.Errorf("no local file provided for thick bundle verification")
	}

	target, bundle, err := tuf.VerifyTrust(v.ref, v.localFile, trustServer, tlscacert, trustDir, timeout)

	if err == nil && v.intoto {
		if v.verifyOnOS {
			log.Warn("Running in-toto inspections on the OS instead of in container...")
		}
		err = intoto.Verify(v.verifyOnOS, v.verificationImage, target, bundle, logLevel)
	}

	return err
}
