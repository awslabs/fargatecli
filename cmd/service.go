package cmd

import (
	"github.com/spf13/cobra"
)

var serviceCmd = &cobra.Command{
	Use: "service",
}

func init() {
	rootCmd.AddCommand(serviceCmd)
}
