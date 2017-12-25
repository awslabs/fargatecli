package cmd

import (
	"github.com/spf13/cobra"
)

var serviceEnvCmd = &cobra.Command{
	Use: "env",
}

func init() {
	serviceCmd.AddCommand(serviceEnvCmd)
}
