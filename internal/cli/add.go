package cli

import (
	"fmt"

	"github.com/chiragagg5k/docker-use/internal/accounts"
	"github.com/spf13/cobra"
)

func newAddCommand() *cobra.Command {
	var username string
	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add a new account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if username == "" {
				return fmt.Errorf("--username (-u) is required")
			}
			store, err := accounts.NewStore()
			if err != nil {
				return err
			}
			return store.Add(cmd.Context(), args[0], username)
		},
	}
	cmd.Flags().StringVarP(&username, "username", "u", "", "Docker Hub username")
	_ = cmd.MarkFlagRequired("username")
	return cmd
}
