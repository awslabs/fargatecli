package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/jpignata/fargate/acm"
	"github.com/spf13/cobra"
)

type certificateImportOperation struct {
	acm                  acm.Client
	certificate          []byte
	certificateChain     []byte
	certificateChainFile string
	certificateFile      string
	output               Output
	privateKey           []byte
	privateKeyFile       string
}

func (o certificateImportOperation) execute() {
	if errs := o.validate(); len(errs) > 0 {
		o.output.Fatals(errs, "Invalid certificate import parameters")
		return
	}

	if errs := o.readFiles(); len(errs) > 0 {
		o.output.Fatals(errs, "Could not read file(s)")
		return
	}

	o.output.Debug("Importing certificate [API=acm Action=ImportCertificate]")
	arn, err := o.acm.ImportCertificate(o.certificate, o.privateKey, o.certificateChain)

	if err != nil {
		o.output.Fatal(err, "Could not import certificate")
		return
	}

	o.output.Info("Imported certificate [ARN=%s]", arn)
}

func (o certificateImportOperation) validate() []error {
	var errs []error

	if o.certificateFile == "" {
		errs = append(errs, fmt.Errorf("--certificate is required"))
	}

	if o.privateKeyFile == "" {
		errs = append(errs, fmt.Errorf("--key is required"))
	}

	return errs
}

func (o *certificateImportOperation) readFiles() []error {
	var errs []error

	o.output.Debug("Reading certificate [File=%s]", o.certificateFile)
	if certificate, err := ioutil.ReadFile(o.certificateFile); err == nil {
		o.certificate = certificate
	} else {
		errs = append(errs, err)
	}

	o.output.Debug("Reading private key [File=%s]", o.privateKeyFile)
	if privateKey, err := ioutil.ReadFile(o.privateKeyFile); err == nil {
		o.privateKey = privateKey
	} else {
		errs = append(errs, err)
	}

	if o.certificateChainFile != "" {
		o.output.Debug("Reading certificate chain [File=%s]", o.certificateChainFile)
		if certificateChain, err := ioutil.ReadFile(o.certificateChainFile); err == nil {
			o.certificateChain = certificateChain
		} else {
			errs = append(errs, err)
		}
	}

	return errs
}

var certificateImportCmd = &cobra.Command{
	Use:   "import --certificate <certificate-file> --key <key-file> [--chain <chain-file>]",
	Short: "Import a certificate",
	Long: `Import a certificate

Upload a certificate from a certificate file, a private key file, and optionally
an intermediate certificate chain file. The files must be PEM-encoded and the
private key must not be encrypted or protected by a passphrase. See
http://docs.aws.amazon.com/acm/latest/APIReference/API_ImportCertificate.html
for more details.`,
	Run: func(cmd *cobra.Command, args []string) {
		certificateImportOperation{
			acm:                  acm.New(sess),
			certificateChainFile: certificateImportFlags.chain,
			certificateFile:      certificateImportFlags.certificate,
			output:               output,
			privateKeyFile:       certificateImportFlags.key,
		}.execute()
	},
}

var certificateImportFlags struct {
	certificate, key, chain string
}

func init() {
	certificateImportCmd.Flags().StringVarP(&certificateImportFlags.certificate, "certificate", "c", "",
		"Filename of the certificate to import")
	certificateImportCmd.Flags().StringVarP(&certificateImportFlags.key, "key", "k", "",
		"Filename of the private key used to generate the certificate")
	certificateImportCmd.Flags().StringVar(&certificateImportFlags.chain, "chain", "",
		"Filename of intermediate certificate chain")

	certificateCmd.AddCommand(certificateImportCmd)
}
