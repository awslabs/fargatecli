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

var certificateInfoCmd = &cobra.Command{
	Use:   "info <domain name>",
	Short: "Display information about an SSL certificate",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		infoCertificate(args[0])
	},
}

func init() {
	certificateCmd.AddCommand(certificateInfoCmd)
}

func infoCertificate(domainName string) {
	acm := ACM.New(sess)
	certificate := acm.DescribeCertificate(domainName)

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
				fmt.Sprintf("%s %s -> %s",
					v.ResourceRecord.Type,
					v.ResourceRecord.Name,
					v.ResourceRecord.Value,
				),
			)
		}

		w.Flush()
	}
}
