package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/chiragagg5k/docker-use/internal/accounts"
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "docker-use",
		Short: "Manage multiple Docker Hub accounts",
		Long:  "docker-use lets you switch between Docker Hub accounts by isolating configs in ~/.docker-accounts.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := accountPath(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Account %q is available at %s\n", args[0], path)
			fmt.Fprintf(cmd.OutOrStdout(), "To switch this shell, run: eval \"$(%s init zsh)\" then docker-use %s\n", initCommand(), args[0])
			return nil
		},
	}
	root.AddCommand(
		newListCommand(),
		newWhoamiCommand(),
		newAddCommand(),
		newRemoveCommand(),
		newInitCommand(),
		newPathCommand(),
	)
	return root
}

func newPathCommand() *cobra.Command {
	return &cobra.Command{
		Use:    "__path <name>",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := accountPath(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), path)
			return nil
		},
	}
}

func accountPath(name string) (string, error) {
	store, err := accounts.NewStore()
	if err != nil {
		return "", err
	}
	return store.Export(name)
}

func initCommand() string {
	if strings.ContainsRune(os.Args[0], os.PathSeparator) {
		return os.Args[0]
	}
	return "command " + os.Args[0]
}
