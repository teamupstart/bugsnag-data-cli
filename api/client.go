package api

import (
	"time"

	"github.com/spf13/viper"
	"github.com/zalando/go-keyring"

	"github.com/teamupstart/bugsnag-data-cli/pkg/bugsnag"
	"github.com/teamupstart/bugsnag-data-cli/pkg/netrc"
)

const clientTimeout = 15 * time.Second

var bugsnagClient *bugsnag.Client

// Client initializes and returns bugsnag client.
func Client(config bugsnag.Config) *bugsnag.Client {
	if bugsnagClient != nil {
		return bugsnagClient
	}

	if config.APIEndpoint == "" {
		config.APIEndpoint = viper.GetString("api_endpoint")
	}
	if config.Login == "" {
		config.Login = viper.GetString("login")
	}
	if config.APIToken == "" {
		config.APIToken = viper.GetString("api_token")
	}
	if config.APIToken == "" {
		netrcConfig, _ := netrc.Read(config.APIEndpoint, config.Login)
		if netrcConfig != nil {
			config.APIToken = netrcConfig.Password
		}
	}
	if config.APIToken == "" {
		secret, _ := keyring.Get("bugsnag-data-cli", config.Login)
		config.APIToken = secret
	}

	if config.AuthType == "" {
		config.AuthType = bugsnag.AuthType(viper.GetString("auth_type"))
	}
	if config.AuthType == "" {
		config.AuthType = bugsnag.AuthTypeToken
	}

	bugsnagClient = bugsnag.NewClient(
		config,
		bugsnag.WithTimeout(clientTimeout),
	)

	return bugsnagClient
}
