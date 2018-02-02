package cmd

import (
	"context"
	"sort"
	"strings"
	"sync"

	"github.com/jpignata/fargate/acm"
	"github.com/jpignata/fargate/console"
	"github.com/spf13/cobra"
	"golang.org/x/time/rate"
)

const describeRequestLimitRate = 10

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
		console.Info("No certificates found")
		return
	}

	o.display(certificates)
}

func (o certificateListOperation) find() ([]acm.Certificate, error) {
	o.output.Debug("Listing certificates [API=acm Action=ListCertificates")
	certificates, err := o.acm.ListCertificates()

	if err != nil {
		return []acm.Certificate{}, err
	}

	var wg sync.WaitGroup
	var inflatedCertificates []acm.Certificate

	ctx := context.Background()
	ch := make(chan acm.Certificate, len(certificates))
	limiter := rate.NewLimiter(describeRequestLimitRate, 1)

	for _, certificate := range certificates {
		wg.Add(1)

		go func(c acm.Certificate) {
			defer wg.Done()

			if err := limiter.Wait(ctx); err == nil {
				o.output.Debug("Describing certificate [API=acm Action=DescribeCertificate ARN=%s]", c.Arn)
				certificate, _ := o.acm.InflateCertificate(c)

				ch <- certificate
			}
		}(certificate)
	}

	wg.Wait()
	close(ch)

	for c := range ch {
		inflatedCertificates = append(inflatedCertificates, c)
	}

	return inflatedCertificates, nil
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
				Humanize(certificate.Type),
				Humanize(certificate.Status),
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
