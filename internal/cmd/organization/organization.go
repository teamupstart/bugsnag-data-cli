package organization

import (
	"github.com/spf13/cobra"

	"github.com/teamupstart/bugsnag-data-cli/internal/cmd/organization/list"
)

const helpText = `Organization manages Bugsnag Organizations. See available commands below.`

func NewCmdOrganization() *cobra.Command {
	cmd := cobra.Command{
		Use:         "organization",
		Short:       "Organization manages Bugsnag Organizations",
		Long:        helpText,
		Aliases:     []string{"organizations", "orgs", "org"},
		Annotations: map[string]string{"cmd:main": "true"},
		RunE:        organizations,
	}

	cmd.AddCommand(list.NewCmdList())

	return &cmd
}

func organizations(cmd *cobra.Command, _ []string) error {
	return cmd.Help()
}