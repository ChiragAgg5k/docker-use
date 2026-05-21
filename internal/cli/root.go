package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chiragagg5k/docker-use/internal/accounts"
	"github.com/spf13/cobra"
)

const defaultVersion = "dev"

func NewRootCommand(version ...string) *cobra.Command {
	cmdVersion := defaultVersion
	if len(version) > 0 && version[0] != "" {
		cmdVersion = version[0]
	}
	root := &cobra.Command{
		Use:          "docker-use [account]",
		Short:        "Manage multiple Docker Hub accounts",
		Long:         "docker-use lets you switch between Docker Hub accounts by isolating configs in ~/.docker-accounts.",
		Version:      cmdVersion,
		SilenceUsage: true,
		Args:         cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No account selected. Run `docker-use list` to see accounts or `docker-use <account>` to switch.")
				return nil
			}
			path, err := accountPath(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Account %q is available at %s\n", args[0], path)
			if shell := shellName(); shell != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "To switch this shell, run: eval \"$(%s init %s)\" then docker-use %s\n", initCommand(), shell, args[0])
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "To switch this shell, run: eval \"$(%s init <bash|fish|zsh>)\" then docker-use %s\n", initCommand(), args[0])
			}
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
		newSwitchCommand(),
		newCurrentCommand(),
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

func newSwitchCommand() *cobra.Command {
	return &cobra.Command{
		Use:    "__switch <name>",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := accounts.OpenStore()
			if err != nil {
				return err
			}
			path, err := store.SaveCurrent(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), path)
			return nil
		},
	}
}

func newCurrentCommand() *cobra.Command {
	return &cobra.Command{
		Use:    "__current",
		Hidden: true,
		Args:   cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := accounts.OpenStore()
			if err != nil {
				return err
			}
			_, path, err := store.Current()
			if err != nil {
				return err
			}
			if path != "" {
				fmt.Fprintln(cmd.OutOrStdout(), path)
			}
			return nil
		},
	}
}

func accountPath(name string) (string, error) {
	store, err := accounts.OpenStore()
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

func shellName() string {
	name := filepath.Base(os.Getenv("SHELL"))
	switch name {
	case "bash", "fish", "zsh":
		return name
	default:
		return ""
	}
}
