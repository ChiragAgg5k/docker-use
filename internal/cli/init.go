package cli

import (
	"fmt"

	"github.com/chiragagg5k/docker-use/internal/shell"
	"github.com/spf13/cobra"
)

func newInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init <shell>",
		Short: "Print shell integration script",
		Args:  cobra.ExactArgs(1),
		ValidArgs: []string{"zsh", "bash", "fish"},
		RunE: func(cmd *cobra.Command, args []string) error {
			script, err := shell.InitScript(args[0])
			if err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), script)
			return nil
		},
	}
}
