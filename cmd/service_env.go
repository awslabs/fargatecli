package cmd

import (
	"github.com/spf13/cobra"
)

var serviceEnvCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage environment variables",
}

func init() {
	serviceCmd.AddCommand(serviceEnvCmd)
}
