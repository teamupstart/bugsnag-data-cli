package init

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/teamupstart/bugsnag-data-cli/internal/cmdutil"
	bugsnagConfig "github.com/teamupstart/bugsnag-data-cli/internal/config"
	"github.com/teamupstart/bugsnag-data-cli/internal/query"
	"github.com/teamupstart/bugsnag-data-cli/pkg/bugsnag"
)

type initParams struct {
	api_endpoint string
	login        string
	organization string
	force        bool
}

// NewCmdInit is an init command.
func NewCmdInit() *cobra.Command {
	cmd := cobra.Command{
		Use:     "init",
		Short:   "Init initializes bugsnag config",
		Long:    "Init initializes bugsnag configuration required for the tool to work properly.",
		Aliases: []string{"initialize", "configure", "config", "setup"},
		Run:     initialize,
	}

	cmd.Flags().SortFlags = false

	cmd.Flags().String("api_endpoint", "", "Link to your bugsnag api endpoint")
	cmd.Flags().String("login", "", "Bugsnag api token")
	cmd.Flags().String("organization", "", "Your default organization id")
	cmd.Flags().Bool("force", false, "Forcefully override existing config if it exists")

	return &cmd
}

func parseFlags(flags query.FlagParser) *initParams {
	api_endpoint, err := flags.GetString("api_endpoint")
	cmdutil.ExitIfError(err)

	login, err := flags.GetString("login")
	cmdutil.ExitIfError(err)

	organization, err := flags.GetString("organization")
	cmdutil.ExitIfError(err)

	force, err := flags.GetBool("force")
	cmdutil.ExitIfError(err)

	return &initParams{
		api_endpoint: api_endpoint,
		login:        login,
		organization: organization,
		force:        force,
	}
}

func initialize(cmd *cobra.Command, _ []string) {
	params := parseFlags(cmd.Flags())

	c := bugsnagConfig.NewBugsnagCLIConfigGenerator(
		&bugsnagConfig.BugsnagCLIConfig{
			APIEndpoint:  params.api_endpoint,
			Login:        params.login,
			Organization: params.organization,
			Force:        params.force,
		},
	)

	file, err := c.Generate()

	if err != nil {
		if e, ok := err.(*bugsnag.ErrUnexpectedResponse); ok {
			fmt.Println()
			cmdutil.Failed("Received unexpected response '%s' from bugsnag. Please try again.", e.Status)
		} else {
			switch err {
			case bugsnagConfig.ErrSkip:
				cmdutil.Success("Skipping config generation. Current config: %s", viper.ConfigFileUsed())
			case bugsnagConfig.ErrUnexpectedResponseFormat:
				fmt.Println()
				cmdutil.Failed("Got response in unexpected format when fetching metadata. Please try again.")
			default:
				fmt.Println()
				cmdutil.Failed("Unable to generate configuration: %s", err.Error())
			}
		}
		os.Exit(1)
	}

	cmdutil.Success("Configuration generated: %s", file)
}
