package cmd

import (
	"github.com/spf13/cobra"
)

const logGroupFormat = "/fargate/service/%s"

var serviceCmd = &cobra.Command{
	Use: "service",
}

func init() {
	rootCmd.AddCommand(serviceCmd)
}
