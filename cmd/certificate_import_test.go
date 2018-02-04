package cmd

import (
	"errors"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jpignata/fargate/acm/mock/client"
	"github.com/jpignata/fargate/cmd/mock"
)

func TestCertificateImportOperation(t *testing.T) {
	certificateArn := "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012"
	certificate := readFile("testdata/certificate.crt", t)
	privateKey := readFile("testdata/private.key", t)
	certificateChain := readFile("testdata/chain.crt", t)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockClient := client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	mockClient.EXPECT().ImportCertificate(certificate, privateKey, certificateChain).Return(certificateArn, nil)

	certificateImportOperation{
		acm:                  mockClient,
		certificateChainFile: "testdata/chain.crt",
		certificateFile:      "testdata/certificate.crt",
		output:               mockOutput,
		privateKeyFile:       "testdata/private.key",
	}.execute()

	if len(mockOutput.FatalMsgs) > 0 {
		for _, fatal := range mockOutput.FatalMsgs {
			t.Errorf(fatal.Msg, fatal.Errors)
		}
	}

	if len(mockOutput.InfoMsgs) == 0 {
		t.Errorf("Expected info output from operation, got none")
	}

	if !strings.Contains(mockOutput.InfoMsgs[0], "Imported certificate") {
		t.Errorf("Expected info output to say 'Imported certificate [ARN=%s]', got: %s", certificateArn, mockOutput.InfoMsgs[0])
	}
}

func TestCertificateImportOperationSansChain(t *testing.T) {
	var certificateChain []byte

	certificateArn := "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012"
	certificate := readFile("testdata/certificate.crt", t)
	privateKey := readFile("testdata/private.key", t)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockClient := client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	mockClient.EXPECT().ImportCertificate(certificate, privateKey, certificateChain).Return(certificateArn, nil)

	certificateImportOperation{
		acm:             mockClient,
		certificateFile: "testdata/certificate.crt",
		output:          mockOutput,
		privateKeyFile:  "testdata/private.key",
	}.execute()

	if len(mockOutput.FatalMsgs) > 0 {
		for _, fatal := range mockOutput.FatalMsgs {
			t.Errorf(fatal.Msg, fatal.Errors)
		}
	}

	if len(mockOutput.InfoMsgs) == 0 {
		t.Errorf("Expected info output from operation, got none")
	}

	if !strings.Contains(mockOutput.InfoMsgs[0], "Imported certificate") {
		t.Errorf("Expected info output to say 'Imported certificate [ARN=%s]', got: %s", certificateArn, mockOutput.InfoMsgs[0])
	}
}

func TestCertificateImportOperationMissingParameters(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockClient := client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	certificateImportOperation{
		acm:    mockClient,
		output: mockOutput,
	}.execute()

	if len(mockOutput.FatalMsgs) == 0 {
		t.Errorf("Expected fatal output from operation, got none")
	}

	if !mockOutput.Exited {
		t.Errorf("Expected premature exit, didn't")
	}

	if mockOutput.FatalMsgs[0].Msg != "Invalid certificate import parameters" {
		t.Errorf("Expected fatal output 'Invalid certificate import parameters', got: %s", mockOutput.FatalMsgs[0].Msg)
	}
}

func TestCertificateImportOperationBadFiles(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockClient := client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	certificateImportOperation{
		acm:                  mockClient,
		certificateChainFile: "pretend",
		certificateFile:      "pretend",
		output:               mockOutput,
		privateKeyFile:       "pretend",
	}.execute()

	if len(mockOutput.FatalMsgs) == 0 {
		t.Errorf("Expected fatal output from operation, got none")
	}

	if !mockOutput.Exited {
		t.Errorf("Expected premature exit, didn't")
	}

	if mockOutput.FatalMsgs[0].Msg != "Could not read file(s)" {
		t.Errorf("Expected fatal output 'Could not read file(s)', got: %s", mockOutput.FatalMsgs[0].Msg)
	}
}

func readFile(fileName string, t *testing.T) []byte {
	contents, err := ioutil.ReadFile(fileName)

	if err != nil {
		t.Errorf(err.Error())
	}

	return contents
}

func TestCertificateImportOperationError(t *testing.T) {
	certificate := readFile("testdata/certificate.crt", t)
	privateKey := readFile("testdata/private.key", t)
	certificateChain := readFile("testdata/chain.crt", t)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockClient := client.NewMockClient(mockCtrl)
	mockOutput := &mock.Output{}

	mockClient.EXPECT().ImportCertificate(certificate, privateKey, certificateChain).Return("", errors.New(":-("))

	certificateImportOperation{
		acm:                  mockClient,
		certificateChainFile: "testdata/chain.crt",
		certificateFile:      "testdata/certificate.crt",
		output:               mockOutput,
		privateKeyFile:       "testdata/private.key",
	}.execute()

	if len(mockOutput.FatalMsgs) == 0 {
		t.Errorf("Expected fatal output from operation, got none")
	}

	if !mockOutput.Exited {
		t.Errorf("Expected premature exit, didn't")
	}

	if mockOutput.FatalMsgs[0].Msg != "Could not import certificate" {
		t.Errorf("Expected fatal output 'Could not import certificate', got: %s", mockOutput.FatalMsgs[0].Msg)
	}
}
