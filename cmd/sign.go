package main

import (
	"github.com/spf13/cobra"
)

type signCmd struct {
	gun          string
	file         string
	artifactType string
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

	return cmd
}

func (s *signCmd) run() error {
	return nil
}
