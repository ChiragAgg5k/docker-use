package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

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
			if !force {
				fmt.Fprintf(cmd.ErrOrStderr(), "Remove account %q? [y/N]: ", args[0])
				answer, err := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
				if err != nil && !errors.Is(err, io.EOF) {
					return fmt.Errorf("failed to read confirmation: %w", err)
				}
				if err != nil && errors.Is(err, io.EOF) && answer == "" {
					return fmt.Errorf("aborted")
				}
				if strings.ToLower(strings.TrimSpace(answer)) != "y" {
					return fmt.Errorf("aborted")
				}
			}
			return store.Remove(args[0])
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation")
	return cmd
}
