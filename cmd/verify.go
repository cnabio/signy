package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/docker/go/canonical/json"
	"github.com/engineerd/signy/pkg/trust"

	"github.com/engineerd/signy/pkg/cnab"
	"github.com/spf13/cobra"
)

type verifyCmd struct {
	gun          string
	output       string
	artifactType string
}

func newVerifyCmd() *cobra.Command {
	verify := verifyCmd{}

	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify trust data for an artifact",
		RunE: func(cmd *cobra.Command, args []string) error {
			verify.gun = args[0]
			return verify.run()
		},
	}

	cmd.Flags().StringVarP(&verify.artifactType, "type", "", "plaintext", "Type of the artifact.")
	cmd.Flags().StringVarP(&verify.output, "output", "o", "-", "Output destination of the pulled file. By default, to stdout.")

	return cmd
}

func (v *verifyCmd) run() error {

	parts := strings.Split(v.gun, ":")
	gun := parts[0]
	if len(parts) == 1 {
		parts = append(parts, "latest")
	}
	name := parts[1]

	target, err := trust.GetTargetWithRole(gun, name, trustServer, tlscacert, trustDir)
	if err != nil {
		return err
	}

	fmt.Printf("\nPulled trust data for %v - SHA256: %v", v.gun, hex.EncodeToString(target.Hashes["sha256"]))

	fmt.Printf("\nPulling bundle from registry: %v", gun)
	bun, err := cnab.Pull(v.gun)
	if err != nil {
		return fmt.Errorf("cannot pull bundle: %v", err)
	}

	buf, err := json.MarshalCanonical(bun)
	if err != nil {
		return err
	}

	sum := sha256.Sum256(buf)
	fmt.Printf("\nSHA256 of pulled bundle: %x\n", sum)

	return nil
}
