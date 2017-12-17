package cmd

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"strings"

	ACM "github.com/jpignata/fargate/acm"
	"github.com/jpignata/fargate/console"
	"github.com/spf13/cobra"
)

var (
	certificateFile      string
	privateKeyFile       string
	certificateChainFile string
	certificate          string
	privateKey           string
	certificateChain     string
)

var certificateImportCmd = &cobra.Command{
	Use: "import --certificate <certificate file> --key <key file>",
	Run: func(cmd *cobra.Command, args []string) {
		importCertificate()
	},
	Short: "Import an SSL certificate from local files",
	PreRun: func(cmd *cobra.Command, args []string) {
		validateCertificateAndKeyFiles()
	},
}

func init() {
	certificateCmd.AddCommand(certificateImportCmd)

	certificateImportCmd.Flags().StringVarP(&certificateFile, "certificate", "c", "", "A file containing the certificate to import")
	certificateImportCmd.Flags().StringVarP(&privateKeyFile, "key", "k", "", "A file containing the private key used to generate the certificate")
	certificateImportCmd.Flags().StringVar(&certificateChainFile, "chain", "", "A file containing intermediate certificates")
}

func validateCertificateAndKeyFiles() {
	var msgs []string

	if certificateFile == "" {
		msgs = append(msgs, "--certificate is required")
	}

	if privateKeyFile == "" {
		msgs = append(msgs, "--key is required")
	}

	if len(msgs) > 0 {
		console.ErrorExit(fmt.Errorf(strings.Join(msgs, ", ")), "Invalid command line flags")
	}
}

func importCertificate() {
	console.Info("Importing certificate")

	acm := ACM.New()

	certificateData, err := ioutil.ReadFile(certificateFile)

	if err != nil {
		console.ErrorExit(err, "Could not read certificate from file %s", certificateFile)
	}

	privateKeyData, err := ioutil.ReadFile(privateKeyFile)

	if err != nil {
		console.ErrorExit(err, "Could not read key from file %s", privateKeyFile)
	}

	certificate = base64.StdEncoding.EncodeToString(certificateData)
	privateKey = base64.StdEncoding.EncodeToString(privateKeyData)

	if certificateChainFile != "" {
		certificateChainData, err := ioutil.ReadFile(privateKeyFile)

		if err != nil {
			console.ErrorExit(err, "Could not read certificate chain from file %s", certificateChainFile)
		}

		certificateChain = base64.StdEncoding.EncodeToString(certificateChainData)
	}

	acm.ImportCertificate([]byte(certificate), []byte(privateKey), []byte(certificateChain))
}
