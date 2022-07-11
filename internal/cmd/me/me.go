package me

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewCmdMe is a me command.
func NewCmdMe() *cobra.Command {
	return &cobra.Command{
		Use:   "me",
		Short: "Displays configured bugsnag user",
		Long:  "Displays configured bugsnag user.",
		Run:   me,
	}
}

func me(*cobra.Command, []string) {
	fmt.Println(viper.GetString("login"))
}
