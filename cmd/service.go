package cmd

import (
	"github.com/spf13/cobra"
)

const serviceLogGroupFormat = "/fargate/service/%s"

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Manage services",
}

func init() {
	rootCmd.AddCommand(serviceCmd)
}
