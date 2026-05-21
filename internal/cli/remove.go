package cli

import (
	"github.com/chiragagg5k/docker-use/internal/accounts"
	"github.com/spf13/cobra"
)

func newRemoveCommand() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:     "remove <name>",
		Aliases: []string{"rm"},
		Short:   "Remove an account",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := accounts.NewStore()
			if err != nil {
				return err
			}
			return store.Remove(args[0], force)
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation")
	return cmd
}
