package cmd

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jpignata/fargate/acm/mock/client"
)

func TestCertificateRequestOperation(t *testing.T) {
	certificateArn := "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012"
	domainName := "exmaple.com"
	aliases := []string{"www.example.com"}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockACMClient := client.NewMockACMClient(mockCtrl)

	operation := CertificateRequestOperation{
		ACM:        mockACMClient,
		DomainName: domainName,
		Aliases:    aliases,
	}

	mockACMClient.EXPECT().RequestCertificate(domainName, aliases).Return(certificateArn, nil)

	operation.Validate()
	operation.Execute()
}
