package main

import (
	"github.com/spf13/cobra"

	"github.com/engineerd/signy/pkg/trust"
)

type intotoVerifyCmd struct {
	verificationDir   string
	ref               string
	verificationImage string
	targetFiles       []string
	keepTempDir       bool
}

func newIntotoVerifyCmd() *cobra.Command {
	i := intotoVerifyCmd{}
	cmd := &cobra.Command{
		Use:   "intoto-verify",
		Short: "execute the in-toto verification",
		RunE: func(cmd *cobra.Command, args []string) error {
			i.ref = args[0]
			return i.run()
		},
	}
	cmd.Flags().StringVarP(&i.verificationDir, "verifications", "", "", "directory to run verifications")
	cmd.Flags().StringArrayVarP(&i.targetFiles, "target", "", nil, "target files to copy in container")
	cmd.Flags().StringVarP(&i.verificationImage, "image", "", "", "container image to run the verification")
	cmd.Flags().BoolVarP(&i.keepTempDir, "keep", "", false, "if passed, the temporary directory where the metadata is pulled is not deleted")

	return cmd
}

func (i *intotoVerifyCmd) run() error {
	return trust.Validate(i.ref, "", trustServer, tlscacert, trustDir, i.verificationImage, i.targetFiles, i.keepTempDir)
}
