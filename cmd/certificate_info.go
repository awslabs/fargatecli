package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	ACM "github.com/jpignata/fargate/acm"
	"github.com/jpignata/fargate/console"
	"github.com/jpignata/fargate/util"
	"github.com/spf13/cobra"
)

type CertificateInfoOperation struct {
	DomainName string
}

var certificateInfoCmd = &cobra.Command{
	Use:   "info <domain-name>",
	Short: "Inspect certificate",
	Long: `Inspect certificate

Show extended information for a certificate including each validation for the
certificate including any DNS records which must be created to validate
domain ownership.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		operation := &CertificateInfoOperation{
			DomainName: args[0],
		}

		getCertificateInfo(operation)
	},
}

func init() {
	certificateCmd.AddCommand(certificateInfoCmd)
}

func getCertificateInfo(operation *CertificateInfoOperation) {
	acm := ACM.New(sess)
	certificate := acm.DescribeCertificate(operation.DomainName)

	console.KeyValue("Domain Name", "%s\n", certificate.DomainName)
	console.KeyValue("Status", "%s\n", util.Humanize(certificate.Status))
	console.KeyValue("Type", "%s\n", util.Humanize(certificate.Type))
	console.KeyValue("Subject Alternative Names", "%s\n", strings.Join(certificate.SubjectAlternativeNames, ", "))

	if len(certificate.Validations) > 0 {
		console.Header("Validations")

		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 8, 1, '\t', 0)
		fmt.Fprintln(w, "Domain Name\tStatus\tRecord")

		for _, v := range certificate.Validations {
			fmt.Fprintf(w, "%s\t%s\t%s\n",
				v.DomainName,
				util.Humanize(v.Status),
				v.ResourceRecordString(),
			)
		}

		w.Flush()
	}
}
