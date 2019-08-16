package main

import (
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/engineerd/signy/pkg/cnab"
	"github.com/engineerd/signy/pkg/intoto"
	"github.com/engineerd/signy/pkg/trust"
)

type intotoSignCmd struct {
	layout    string
	layoutKey string
	linkDir   string

	ref  string
	file string
}

func newIntotoSignCmd() *cobra.Command {
	i := intotoSignCmd{}
	cmd := &cobra.Command{
		Use:   "intoto-sign",
		Short: "execute the in-toto verification",
		RunE: func(cmd *cobra.Command, args []string) error {
			i.file = args[0]
			i.ref = args[1]
			return i.run()
		},
	}
	cmd.Flags().StringVarP(&i.layout, "layout", "", "", "path to the root layout file")
	cmd.Flags().StringVarP(&i.layoutKey, "layout-key", "", "", "path to the root layout public key")
	cmd.Flags().StringVarP(&i.linkDir, "links", "", "", "path to the links directory")

	return cmd
}

func (i *intotoSignCmd) run() error {
	err := intoto.Verify(i.layout, i.linkDir, i.layoutKey)
	if err != nil {
		return fmt.Errorf("validation for in-toto metadata failed: %v", err)
	}
	r, err := intoto.GetMetadataRawMessage(i.layout, i.linkDir, i.layoutKey)
	if err != nil {
		return fmt.Errorf("cannot get metadata message: %v", err)
	}

	fmt.Printf("\nAdding In-Toto layout and links metadata to TUF")

	target, err := trust.SignAndPublish(trustDir, trustServer, i.ref, i.file, tlscacert, "", &r)
	if err != nil {
		return fmt.Errorf("cannot sign and publish trust data: %v", err)
	}
	fmt.Printf("\nPushed trust data for %v: %v\n", i.ref, hex.EncodeToString(target.Hashes["sha256"]))
	return cnab.Push(i.file, i.ref)
}
