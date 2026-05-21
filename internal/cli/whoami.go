package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/chiragagg5k/docker-use/internal/accounts"
	"github.com/spf13/cobra"
)

func newWhoamiCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show current account",
		RunE: func(cmd *cobra.Command, args []string) error {
			config := os.Getenv("DOCKER_CONFIG")
			if config == "" {
				fmt.Fprintln(cmd.OutOrStdout(), "DOCKER_CONFIG is not set.")
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "DOCKER_CONFIG=%s\n", config)

			store, err := accounts.NewStore()
			if err != nil {
				return err
			}
			acc, err := accounts.CurrentFromEnv(store)
			if err != nil {
				return err
			}
			if acc == "" {
				acc = "unknown"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Account: %s\n", acc)

			configPath := filepath.Join(config, "config.json")
			user, err := accounts.DockerHubUsername(configPath)
			if err != nil {
				if os.IsNotExist(err) {
					fmt.Fprintln(cmd.OutOrStdout(), "Docker Hub user: not logged in")
					return nil
				}
				return err
			}
			if user == "" {
				if acc != "unknown" {
					user, err = store.Username(acc)
					if err != nil {
						return err
					}
				}
			}
			if user == "" {
				fmt.Fprintln(cmd.OutOrStdout(), "Docker Hub user: not logged in")
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Docker Hub user: %s\n", user)
			return nil
		},
	}
}
