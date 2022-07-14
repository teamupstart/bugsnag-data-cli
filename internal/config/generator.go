package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/viper"

	"github.com/teamupstart/bugsnag-data-cli/api"
	"github.com/teamupstart/bugsnag-data-cli/internal/cmdutil"
	"github.com/teamupstart/bugsnag-data-cli/pkg/bugsnag"
)

const (
	// Dir is a bugsnag-data-cli config directory.
	Dir = ".bugsnag"
	// FileName is a bugsnag-data-cli config file name.
	FileName = ".config"
	// FileType is a bugsnag-data-cli config file extension.
	FileType = "yml"
)

var (
	// ErrSkip is returned when a user skips the config generation.
	ErrSkip = fmt.Errorf("skipping config generation")
	// ErrUnexpectedResponseFormat is returned if the response data is in unexpected format.
	ErrUnexpectedResponseFormat = fmt.Errorf("unexpected response format")
)

type organizationConf struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

// BugsnagCLIConfig is a Bugsnag CLI config.
type BugsnagCLIConfig struct {
	APIEndpoint  string
	Login        string
	Organization string
	Force        bool
}

// BugsnagCLIConfigGenerator is a Bugsnag CLI config generator.
type BugsnagCLIConfigGenerator struct {
	usrCfg *BugsnagCLIConfig
	value  struct {
		api_endpoint string
		login        string
		authType     bugsnag.AuthType
		organization *organizationConf
	}
	bugsnagClient           *bugsnag.Client
	organizationsMap        map[string]*organizationConf
	organizationSuggestions []string
}

// NewBugsnagCLIConfigGenerator creates a new Bugsnag CLI config.
func NewBugsnagCLIConfigGenerator(cfg *BugsnagCLIConfig) *BugsnagCLIConfigGenerator {
	gen := BugsnagCLIConfigGenerator{
		usrCfg:           cfg,
		organizationsMap: make(map[string]*organizationConf),
	}

	return &gen
}

// Generate generates the config file.
func (c *BugsnagCLIConfigGenerator) Generate() (string, error) {
	ce := func() bool {
		s := cmdutil.Info("Checking configuration...")
		defer s.Stop()

		return Exists(viper.ConfigFileUsed())
	}()

	if !c.usrCfg.Force && ce && !shallOverwrite() {
		return "", ErrSkip
	}
	if err := c.configureEndpointAndLoginDetails(); err != nil {
		return "", err
	}
	if err := c.configureOrganizationDetails(); err != nil {
		return "", err
	}

	home, err := cmdutil.GetConfigHome()
	if err != nil {
		return "", err
	}
	cfgDir := fmt.Sprintf("%s/%s", home, Dir)

	if err := func() error {
		s := cmdutil.Info("Creating new configuration...")
		defer s.Stop()

		return create(cfgDir, fmt.Sprintf("%s.%s", FileName, FileType))
	}(); err != nil {
		return "", err
	}

	return c.write(cfgDir)
}

func (c *BugsnagCLIConfigGenerator) configureEndpointAndLoginDetails() error {
	var qs []*survey.Question

	c.value.api_endpoint = c.usrCfg.APIEndpoint
	c.value.login = c.usrCfg.Login

	if c.usrCfg.APIEndpoint == "" {
		qs = append(qs, &survey.Question{
			Name: "api_endpoint",
			Prompt: &survey.Input{
				Message: "Link to Bugsnag API Endpoint:",
				Help:    "This is a link to your bugsnag api endpoint, eg: https://api.bugsnag.com",
				Default: "https://api.bugsnag.com",
			},
			Validate: func(val interface{}) error {
				errInvalidURL := fmt.Errorf("not a valid URL")

				str, ok := val.(string)
				if !ok {
					return errInvalidURL
				}
				u, err := url.Parse(str)
				if err != nil || u.Scheme == "" || u.Host == "" {
					return errInvalidURL
				}
				if u.Scheme != "https" {
					return errInvalidURL
				}

				return nil
			},
		})
	}

	if c.usrCfg.Login == "" {
		qs = append(qs, &survey.Question{
			Name: "login",
			Prompt: &survey.Input{
				Message: "Login username:",
				Help:    "This is the username you use to login to your bugsnag account.",
			},
			Validate: func(val interface{}) error {
				var (
					errInvalidUser = fmt.Errorf("not a valid user")
				)

				str, ok := val.(string)
				if !ok {
					return errInvalidUser
				}
				if len(str) < 3 || len(str) > 254 {
					return errInvalidUser
				}

				return nil
			},
		})

	}

	if len(qs) > 0 {
		ans := struct {
			APIEndpoint string `survey:"api_endpoint"`
			Login       string `survey:"login"`
		}{}

		if err := survey.Ask(qs, &ans); err != nil {
			return err
		}

		c.value.api_endpoint = ans.APIEndpoint
		c.value.login = ans.Login
	}

	return c.verifyLoginDetails(c.value.api_endpoint, c.value.login)
}

func (c *BugsnagCLIConfigGenerator) verifyLoginDetails(api_endpoint, login string) error {
	s := cmdutil.Info("Verifying login details...")
	defer s.Stop()

	api_endpoint = strings.TrimRight(api_endpoint, "/")

	c.bugsnagClient = api.Client(bugsnag.Config{
		APIEndpoint: api_endpoint,
		Login:       login,
		AuthType:    c.value.authType,
		Debug:       viper.GetBool("debug"),
	})

	if ret, err := c.bugsnagClient.Me(); err != nil {
		return err
	} else if c.value.authType == bugsnag.AuthTypeToken {
		login = ret.Login
	}

	c.value.api_endpoint = api_endpoint
	c.value.login = login

	return nil
}

func (c *BugsnagCLIConfigGenerator) configureOrganizationDetails() error {
	organization := c.usrCfg.Organization

	if err := c.getOrganizationSuggestions(); err != nil {
		return err
	}

	if c.usrCfg.Organization == "" {
		organizationPrompt := survey.Select{
			Message: "Default organization:",
			Help:    "This is your organization id that you want to access by default when using the cli.",
			Options: c.organizationSuggestions,
		}
		if err := survey.AskOne(&organizationPrompt, &organization, survey.WithValidator(survey.Required)); err != nil {
			return err
		}
	}
	c.value.organization = c.organizationsMap[strings.ToLower(organization)]

	if c.value.organization == nil {
		return fmt.Errorf("organization not found\n  Please check the organization id and try again")
	}

	return nil
}

func (c *BugsnagCLIConfigGenerator) getOrganizationSuggestions() error {
	s := cmdutil.Info("Fetching organizations...")
	defer s.Stop()

	organizations, err := c.bugsnagClient.Organization()
	if err != nil {
		return err
	}
	for _, organization := range organizations {
		c.organizationsMap[strings.ToLower(organization.Name)] = &organizationConf{
			Id:   organization.Id,
			Name: organization.Name,
		}
		c.organizationSuggestions = append(c.organizationSuggestions, organization.Name)
	}

	return nil
}

func (c *BugsnagCLIConfigGenerator) write(path string) (string, error) {
	config := viper.New()
	config.AddConfigPath(path)
	config.SetConfigName(FileName)
	config.SetConfigType(FileType)

	config.Set("api_endpoint", c.value.api_endpoint)
	config.Set("login", c.value.login)
	config.Set("organization", c.value.organization)

	if err := config.WriteConfig(); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s.%s", path, FileName, FileType), nil
}

// Exists checks if the file exist.
func Exists(file string) bool {
	if file == "" {
		return false
	}
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return false
	}
	return true
}

func shallOverwrite() bool {
	var ans bool

	prompt := &survey.Confirm{
		Message: "Config already exist. Do you want to overwrite?",
	}
	if err := survey.AskOne(prompt, &ans); err != nil {
		return false
	}

	return ans
}

func create(path, name string) error {
	const perm = 0o700

	if !Exists(path) {
		if err := os.MkdirAll(path, perm); err != nil {
			return err
		}
	}

	file := fmt.Sprintf("%s/%s", path, name)
	if Exists(file) {
		if err := os.Rename(file, file+".bkp"); err != nil {
			return err
		}
	}
	_, err := os.Create(file)

	return err
}
