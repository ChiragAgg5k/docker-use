package cli

import (
	"fmt"

	"github.com/chiragagg5k/docker-use/internal/accounts"
	"github.com/spf13/cobra"
)

func newUseCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "use <name>",
		Short: "Switch to an account (shell wrapper intercepts this)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := accounts.NewStore()
			if err != nil {
				return err
			}
			line, err := store.Export(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), line)
			return nil
		},
	}
}
