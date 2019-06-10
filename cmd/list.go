package main

import (
	"github.com/engineerd/signy/pkg/trust"

	"github.com/spf13/cobra"
)

type listCmd struct {
	gun string
}

func newListCmd() *cobra.Command {
	list := listCmd{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lists targets for a remote trusted collection",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			list.gun = args[0]
			return list.run()
		},
	}

	return cmd
}

func (l *listCmd) run() error {
	return trust.PrintTargets(l.gun, trustServer, tlscacert, trustDir)
}
