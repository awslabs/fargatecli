package sdk

import (
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/aws/aws-sdk-go/service/acm/acmiface"
)

type MockCertificateListPagesClient struct {
	acmiface.ACMAPI
	Resp  *acm.ListCertificatesOutput
	Error error
}

func (m MockCertificateListPagesClient) ListCertificatesPages(in *acm.ListCertificatesInput, fn func(*acm.ListCertificatesOutput, bool) bool) error {
	fn(m.Resp, true)

	if m.Error != nil {
		return m.Error
	}

	return nil
}
