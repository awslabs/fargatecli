package cmd

import (
	"errors"

	"github.com/jpignata/fargate/acm"
	"github.com/spf13/cobra"
)

type certificateOperation struct {
	acm acm.Client
}

func (o certificateOperation) findCertificate(domainName string, output Output) (acm.Certificate, error) {
	output.Debug("Listing certificates [API=acm Action=ListCertificate]")
	certificates, err := o.acm.ListCertificates()

	if err != nil {
		return acm.Certificate{}, err
	}

	certificates = certificates.GetCertificates(domainName)

	switch {
	case len(certificates) == 0:
		return acm.Certificate{}, errCertificateNotFound
	case len(certificates) > 1:
		return acm.Certificate{}, errCertificateTooManyFound
	}

	output.Debug("Describing certificate [API=acm Action=DescribeCertificate ARN=%s]", certificates[0].Arn)
	return o.acm.InflateCertificate(certificates[0])
}

var (
	errCertificateNotFound     = errors.New("Certificate not found")
	errCertificateTooManyFound = errors.New("Too many certificates found")
)

var certificateCmd = &cobra.Command{
	Use:   "certificate",
	Short: "Manage certificates",
	Long: `Manages certificate

Certificates are TLS certificates issued by or imported into AWS Certificate
Manager for use in securing traffic between load balancers and end users. ACM
provides TLS certificates free of charge for use within AWS resources.`,
}

func init() {
	rootCmd.AddCommand(certificateCmd)
}
