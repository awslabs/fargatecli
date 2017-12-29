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

type CertificateImportOperation struct {
	CertificateFile      string
	PrivateKeyFile       string
	CertificateChainFile string
}

func (o *CertificateImportOperation) Validate() {
	var msgs []string

	if o.CertificateFile == "" {
		msgs = append(msgs, "--certificate is required")
	}

	if o.PrivateKeyFile == "" {
		msgs = append(msgs, "--key is required")
	}

	if len(msgs) > 0 {
		console.ErrorExit(fmt.Errorf(strings.Join(msgs, ", ")), "Invalid command line flags")
	}
}

var (
	flagCertificateImportCertificate string
	flagCertificateImportKey         string
	flagCertificateImportChain       string
)

var certificateImportCmd = &cobra.Command{
	Use:   "import --certificate <certificate file> --key <key file>",
	Short: "Import a certificate",
	Long: `Import a certificate

Upload a certificate from a certificate file, a private key file, an optionally
an intermediate certificate chain file. The files must be PEM-encoded and the
private key must not be encrypted or protected by a passphrase. See
http://docs.aws.amazon.com/acm/latest/APIReference/API_ImportCertificate.html 
for more details.`,
	Run: func(cmd *cobra.Command, args []string) {
		operation := &CertificateImportOperation{
			CertificateFile:      flagCertificateImportCertificate,
			PrivateKeyFile:       flagCertificateImportKey,
			CertificateChainFile: flagCertificateImportChain,
		}

		operation.Validate()

		importCertificate(operation)
	},
}

func init() {
	certificateImportCmd.Flags().StringVarP(&flagCertificateImportCertificate, "certificate", "c", "", "A file containing the certificate to import")
	certificateImportCmd.Flags().StringVarP(&flagCertificateImportKey, "key", "k", "", "A file containing the private key used to generate the certificate")
	certificateImportCmd.Flags().StringVar(&flagCertificateImportChain, "chain", "", "A file containing intermediate certificates")

	certificateCmd.AddCommand(certificateImportCmd)
}

func importCertificate(operation *CertificateImportOperation) {
	var (
		certificate      string
		privateKey       string
		certificateChain string
	)

	acm := ACM.New(sess)

	certificateData, err := ioutil.ReadFile(operation.CertificateFile)

	if err != nil {
		console.ErrorExit(err, "Could not read certificate from file %s", operation.CertificateFile)
	}

	privateKeyData, err := ioutil.ReadFile(operation.PrivateKeyFile)

	if err != nil {
		console.ErrorExit(err, "Could not read key from file %s", operation.PrivateKeyFile)
	}

	certificate = base64.StdEncoding.EncodeToString(certificateData)
	privateKey = base64.StdEncoding.EncodeToString(privateKeyData)

	if operation.CertificateChainFile != "" {
		certificateChainData, err := ioutil.ReadFile(operation.CertificateChainFile)

		if err != nil {
			console.ErrorExit(err, "Could not read certificate chain from file %s", operation.CertificateChainFile)
		}

		certificateChain = base64.StdEncoding.EncodeToString(certificateChainData)
	}

	acm.ImportCertificate([]byte(certificate), []byte(privateKey), []byte(certificateChain))
	console.Info("Imported certificate from %s", operation.CertificateFile)
}
