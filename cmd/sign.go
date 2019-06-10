package main

import (
	"github.com/engineerd/signy/pkg/trust"

	"github.com/spf13/cobra"
)

type signCmd struct {
	gun          string
	file         string
	artifactType string
	rootKey      string
}

func newSignCmd() *cobra.Command {
	sign := signCmd{}

	cmd := &cobra.Command{
		Use:   "sign",
		Short: "Sign an artifact",
		RunE: func(cmd *cobra.Command, args []string) error {
			sign.file = args[0]
			sign.gun = args[1]
			return sign.run()
		},
	}
	cmd.Flags().StringVarP(&sign.artifactType, "type", "", "plaintext", "Type of the artifact")
	cmd.Flags().StringVarP(&sign.rootKey, "rootkey", "", "", "Root key to initialize the repository with")

	return cmd
}

func (s *signCmd) run() error {
	return trust.SignAndPublish(trustDir, trustServer, s.gun, s.file, tlscacert, s.rootKey)
}
