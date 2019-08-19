package main

import (
	"github.com/engineerd/signy/pkg/trust"
	"github.com/spf13/cobra"
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
	return trust.SignAndPublish(i.ref, i.layout, i.linkDir, i.layoutKey, trustDir, trustServer, i.file, tlscacert)
}
