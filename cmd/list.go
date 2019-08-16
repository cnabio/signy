package main

import (
	"github.com/spf13/cobra"

	"github.com/engineerd/signy/pkg/trust"
)

type listCmd struct {
	gun string
}

func newListCmd() *cobra.Command {
	const listDesc = `
Lists all targets for a given trust collection on a remote server.

Example:
$ signy list docker.io/library/alpine

3.5     66952b313e51c3bd1987d7c4ddf5dba9bc0fb6e524eed2448fa660246b3e76ec
3.8     04696b491e0cc3c58a75bace8941c14c924b9f313b03ce5029ebbc040ed9dcd9
3.2     e9a2035f9d0d7cee1cdd445f5bfa0c5c646455ee26f14565dce23cf2d2de7570
3.6     66790a2b79e1ea3e1dabac43990c54aca5d1ddf268d9a5a0285e4167c8b24475
3.10    6a92cd1fcdc8d8cdec60f33dda4db2cb1fcdcacf3410a8e05b3741f44a9b5998
3.9.4   7746df395af22f04212cd25a92c1d6dbc5a06a0ca9579a229ef43008d4d1302a
`
	list := listCmd{}
	cmd := &cobra.Command{
		Use:   "list [GUN]",
		Short: "Lists all targets for a given remote collection on a trust server.",
		Long:  listDesc,
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
