package cmd

import (
	"errors"

	"github.com/jpignata/fargate/acm"
	"github.com/spf13/cobra"
)

type certificateOperation struct {
	acm    acm.Client
	output Output
}

func (o certificateOperation) findCertificate(domainName string) (acm.Certificate, error) {
	o.output.Debug("Listing certificates [API=acm Action=ListCertificate]")
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

	o.output.Debug("Describing certificate [API=acm Action=DescribeCertificate ARN=%s]", certificates[0].ARN)

	if err := o.acm.InflateCertificate(&certificates[0]); err != nil {
		return acm.Certificate{}, err
	}

	return certificates[0], nil
}

var (
	errCertificateNotFound     = errors.New("certificate not found")
	errCertificateTooManyFound = errors.New("too many certificates found")

	certificateCmd = &cobra.Command{
		Use:   "certificate",
		Short: "Manage certificates",
		Long: `Manages certificate

Certificates are TLS certificates issued by or imported into AWS Certificate
Manager for use in securing traffic between load balancers and end users. ACM
provides TLS certificates free of charge for use within AWS resources.`,
	}
)

func init() {
	rootCmd.AddCommand(certificateCmd)
}
