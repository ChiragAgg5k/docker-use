package cli

import (
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "docker-use",
		Short: "Manage multiple Docker Hub accounts",
		Long:  "docker-use lets you switch between Docker Hub accounts by isolating configs in ~/.docker-accounts.",
	}
	root.AddCommand(
		newListCommand(),
		newWhoamiCommand(),
		newAddCommand(),
		newRemoveCommand(),
		newUseCommand(),
		newInitCommand(),
	)
	return root
}
