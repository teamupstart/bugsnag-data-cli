package list

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/teamupstart/bugsnag-data-cli/api"
	"github.com/teamupstart/bugsnag-data-cli/internal/cmdutil"
	"github.com/teamupstart/bugsnag-data-cli/pkg/bugsnag"
)

// NewCmdList is a list command.
func NewCmdList() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Short:   "Lists Bugsnag organizations",
		Long:    "Lists Bugsnag organizations that a user has access to.",
		Aliases: []string{"lists", "ls"},
		Run:     List,
	}
}

// List displays a list view.
func List(cmd *cobra.Command, _ []string) {
	debug, err := cmd.Flags().GetBool("debug")
	cmdutil.ExitIfError(err)

	organizations, total, err := func() ([]*bugsnag.Organization, int, error) {
		s := cmdutil.Info("Fetching organizations...")
		defer s.Stop()

		organizations, err := api.Client(bugsnag.Config{Debug: debug}).Organization()
		if err != nil {
			return nil, 0, err
		}
		return organizations, len(organizations), nil
	}()
	cmdutil.ExitIfError(err)

	if total == 0 {
		cmdutil.Failed("No organizations found.")
		return
	}

	for _, org := range organizations {
		fmt.Printf("%+v\n", org)
	}
}
