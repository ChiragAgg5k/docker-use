package cli

import (
	"github.com/chiragagg5k/docker-use/internal/accounts"
	"github.com/spf13/cobra"
)

func newAddCommand() *cobra.Command {
	var username string
	var force bool
	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add a new account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := accounts.NewStore()
			if err != nil {
				return err
			}
			return store.Add(cmd.Context(), args[0], username, force)
		},
	}
	cmd.Flags().StringVarP(&username, "username", "u", "", "Docker Hub username")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Replace an existing account")
	_ = cmd.MarkFlagRequired("username")
	return cmd
}
