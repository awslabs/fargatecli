package cmd

import (
	"context"
	"sort"
	"strings"
	"sync"

	"github.com/jpignata/fargate/acm"
	"github.com/spf13/cobra"
	"golang.org/x/time/rate"
)

type certificateListOperation struct {
	acm    acm.Client
	output Output
}

func (o certificateListOperation) execute() {
	certificates, err := o.find()

	if err != nil {
		o.output.Fatal(err, "Could not list certificates")
		return
	}

	if len(certificates) == 0 {
		o.output.Info("No certificates found")
		return
	}

	o.display(certificates)
}

func (o certificateListOperation) find() (acm.Certificates, error) {
	var wg sync.WaitGroup

	o.output.Debug("Listing certificates [API=acm Action=ListCertificates]")
	certificates, err := o.acm.ListCertificates()

	if err != nil {
		return acm.Certificates{}, err
	}

	errs := make(chan error)
	done := make(chan bool)
	limiter := rate.NewLimiter(describeRequestLimitRate, 1)

	for i := 0; i < len(certificates); i++ {
		wg.Add(1)

		go func(index int) {
			defer wg.Done()

			if err := limiter.Wait(context.Background()); err == nil {
				o.output.Debug("Describing certificate [API=acm Action=DescribeCertificate ARN=%s]", certificates[index].ARN)
				if err := o.acm.InflateCertificate(&certificates[index]); err != nil {
					errs <- err
				}
			}
		}(i)
	}

	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case err := <-errs:
		return acm.Certificates{}, err
	case <-done:
		return certificates, nil
	}
}

func (o certificateListOperation) display(certificates []acm.Certificate) {
	rows := [][]string{
		[]string{"CERTIFICATE", "TYPE", "STATUS", "SUBJECT ALTERNATIVE NAMES"},
	}

	sort.Slice(certificates, func(i, j int) bool {
		return certificates[i].DomainName < certificates[j].DomainName
	})

	for _, certificate := range certificates {
		rows = append(rows,
			[]string{
				certificate.DomainName,
				Titleize(certificate.Type),
				Titleize(certificate.Status),
				strings.Join(certificate.SubjectAlternativeNames, ", "),
			},
		)
	}

	o.output.Table("", rows)
}

var certificateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List certificates",
	Run: func(cmd *cobra.Command, args []string) {
		certificateListOperation{
			acm:    acm.New(sess),
			output: output,
		}.execute()
	},
}

func init() {
	certificateCmd.AddCommand(certificateListCmd)
}
