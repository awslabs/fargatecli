package cmd

import (
	"strconv"
	"strings"

	"github.com/jpignata/fargate/console"
	"github.com/spf13/cobra"
)

const version = "0.0.1"

type Port struct {
	Port     int64
	Protocol string
}

var verbose bool

var rootCmd = &cobra.Command{
	Use: "fargate",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if verbose {
			verbose = true
			console.Verbose = true
		}
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
}

func Execute() {
	rootCmd.Version = version
	rootCmd.Execute()
}

func inflatePort(src string) (port Port) {
	ports := inflatePorts([]string{src})
	return ports[0]
}

func inflatePorts(src []string) (ports []Port) {
	for _, portRaw := range src {
		if portRaw == "80" {
			ports = append(ports,
				Port{
					Port:     80,
					Protocol: "HTTP",
				},
			)
		} else if portRaw == "443" {
			ports = append(ports,
				Port{
					Port:     443,
					Protocol: "HTTPS",
				},
			)
		} else if strings.Index(portRaw, ":") > 1 {
			portRawContents := strings.Split(portRaw, ":")
			protocol := strings.ToUpper(portRawContents[0])
			port, err := strconv.ParseInt(portRawContents[1], 10, 64)

			if err != nil {
				console.ErrorExit(err, "Invalid command line flags")
			}

			ports = append(ports,
				Port{
					Port:     port,
					Protocol: protocol,
				},
			)
		} else {
			port, err := strconv.ParseInt(portRaw, 10, 64)

			if err != nil {
				console.ErrorExit(err, "Invalid command line flags")
			}

			ports = append(ports,
				Port{
					Port:     port,
					Protocol: "TCP",
				},
			)
		}
	}

	return
}
