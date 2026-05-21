package cli

import (
	"fmt"

	"github.com/chiragagg5k/docker-use/internal/accounts"
	"github.com/spf13/cobra"
)

func newListCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List accounts",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := accounts.NewStore()
			if err != nil {
				return err
			}
			accs, err := store.List()
			if err != nil {
				return err
			}
			if len(accs) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No accounts found.")
				return nil
			}
			for _, a := range accs {
				fmt.Fprintln(cmd.OutOrStdout(), a.Name)
			}
			return nil
		},
	}
}
