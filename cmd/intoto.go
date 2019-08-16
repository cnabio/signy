package main

import (
	"github.com/engineerd/signy/pkg/intoto"
	"github.com/spf13/cobra"
)

type intotoCmd struct {
	layout    string
	layoutKey string
	linkDir   string
}

func newIntotoCmd() *cobra.Command {
	i := intotoCmd{}
	cmd := &cobra.Command{
		Use:   "intoto",
		Short: "execute the in-toto verification",
		RunE: func(cmd *cobra.Command, args []string) error {
			return i.run()
		},
	}
	cmd.Flags().StringVarP(&i.layout, "layout", "", "", "path to the root layout file")
	cmd.Flags().StringVarP(&i.layoutKey, "layout-key", "", "", "path to the root layout public key")
	cmd.Flags().StringVarP(&i.linkDir, "links", "", "", "path to the links directory")

	return cmd
}

func (i *intotoCmd) run() error {
	return intoto.Verify(i.layout, i.linkDir, i.layoutKey)
}
