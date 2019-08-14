package main

import (
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/engineerd/signy/pkg/cnab"
	"github.com/engineerd/signy/pkg/trust"
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
	switch s.artifactType {
	case "plaintext":
		_, err := trust.SignAndPublish(trustDir, trustServer, s.gun, s.file, tlscacert, s.rootKey)
		return err
	case "cnab":
		target, err := trust.SignAndPublish(trustDir, trustServer, s.gun, s.file, tlscacert, s.rootKey)
		if err != nil {
			return fmt.Errorf("cannot sign and publish trust data: %v", err)
		}
		fmt.Printf("\nPushed trust data for %v: %v\n", s.gun, hex.EncodeToString(target.Hashes["sha256"]))
		return cnab.Push(s.file, s.gun)
	default:
		return fmt.Errorf("unknown type")
	}
}
