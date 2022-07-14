package root

import (
	"fmt"
	"os"

	"github.com/teamupstart/bugsnag-data-cli/pkg/netrc"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	initCmd "github.com/teamupstart/bugsnag-data-cli/internal/cmd/init"
	"github.com/teamupstart/bugsnag-data-cli/internal/cmd/me"
	"github.com/teamupstart/bugsnag-data-cli/internal/cmd/organization"
	"github.com/teamupstart/bugsnag-data-cli/internal/cmd/version"
	"github.com/teamupstart/bugsnag-data-cli/internal/cmdutil"
	bugsnagConfig "github.com/teamupstart/bugsnag-data-cli/internal/config"

	"github.com/zalando/go-keyring"
)

const (
	bugsnagCLIHelpLink  = "https://github.com/teamupstart/bugsnag-data-cli#getting-started"
	bugsnagAPITokenLink = "https://bugsnagapiv2.docs.apiary.io/#introduction/authentication"
)

var (
	config string
	debug  bool
)

func init() {
	cobra.OnInitialize(func() {
		if config != "" {
			viper.SetConfigFile(config)
		} else {
			home, err := cmdutil.GetConfigHome()
			if err != nil {
				cmdutil.Failed("Error: %s", err)
				return
			}

			viper.AddConfigPath(fmt.Sprintf("%s/%s", home, bugsnagConfig.Dir))
			viper.SetConfigName(bugsnagConfig.FileName)
			viper.SetConfigType(bugsnagConfig.FileType)
		}

		viper.AutomaticEnv()
		viper.SetEnvPrefix("bugsnag")

		if err := viper.ReadInConfig(); err == nil && debug {
			fmt.Printf("Using config file: %s\n", viper.ConfigFileUsed())
		}
	})
}

// NewCmdRoot is a root command.
func NewCmdRoot() *cobra.Command {
	cmd := cobra.Command{
		Use:   "bugsnag",
		Short: "Interactive Bugsnag CLI",
		Long:  "Interactive Bugsnag CLI.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			subCmd := cmd.Name()
			if !cmdRequireToken(subCmd) {
				return
			}

			checkForBugsnagToken(viper.GetString("api_endpoint"), viper.GetString("login"))

			configFile := viper.ConfigFileUsed()
			if !bugsnagConfig.Exists(configFile) {
				cmdutil.Failed("Missing configuration file.\nRun 'bugsnag init' to configure the tool.")
			}
		},
	}

	configHome, err := cmdutil.GetConfigHome()
	if err != nil {
		cmdutil.Failed("Error: %s", err)
	}

	cmd.PersistentFlags().StringVarP(
		&config, "config", "c", "",
		fmt.Sprintf("Config file (default is %s/%s/%s.yml)", configHome, bugsnagConfig.Dir, bugsnagConfig.FileName),
	)
	cmd.PersistentFlags().StringP(
		"organization", "o", "",
		fmt.Sprintf(
			"Bugsnag organization to look into (defaults to %s/%s/%s.yml)",
			configHome, bugsnagConfig.Dir, bugsnagConfig.FileName,
		),
	)
	cmd.PersistentFlags().BoolVar(&debug, "debug", false, "Turn on debug output")

	cmd.SetHelpFunc(helpFunc)

	_ = viper.BindPFlag("config", cmd.PersistentFlags().Lookup("config"))
	_ = viper.BindPFlag("organization.id", cmd.PersistentFlags().Lookup("organization"))
	_ = viper.BindPFlag("debug", cmd.PersistentFlags().Lookup("debug"))

	addChildCommands(&cmd)

	return &cmd
}

func addChildCommands(cmd *cobra.Command) {
	cmd.AddCommand(
		initCmd.NewCmdInit(),
		me.NewCmdMe(),
		organization.NewCmdOrganization(),
		version.NewCmdVersion(),
	)
}

func cmdRequireToken(cmd string) bool {
	allowList := []string{
		"init",
		"help",
		"bugsnag",
		"version",
	}

	for _, item := range allowList {
		if item == cmd {
			return false
		}
	}

	return true
}

func checkForBugsnagToken(api_endpoint string, login string) {
	if os.Getenv("BUGSNAG_API_TOKEN") != "" {
		return
	}

	netrcConfig, _ := netrc.Read(api_endpoint, login)
	if netrcConfig != nil {
		return
	}

	secret, _ := keyring.Get("bugsnag-data-cli", login)
	if secret != "" {
		return
	}

	msg := fmt.Sprintf(`The tool needs a Bugsnag API token to function.

You can generate the token using this link: %s

After generating the token, you can either:
  - Export API token to your shell as a BUGSNAG_API_TOKEN env variable
  - Or, you can use a .netrc file to define required machine details

Once you are done with the above steps, run 'bugsnag init' to generate the config if you haven't already.

For more details, see: %s
`, bugsnagAPITokenLink, bugsnagCLIHelpLink)

	cmdutil.Warn(msg)
	os.Exit(1)
}
