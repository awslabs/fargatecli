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

var certificateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List requested and imported certificates",
	Long: `List requested and imported certificates

Shows all certificates within AWS Certificate Manager and their attributes
such as type, status, and subject alternative names.

e.g.:

  $ fargate certificate list
  CERTIFICATE      TYPE          STATUS              SUBJECT ALTERNATIVE NAMES
  www.example.com  Amazon Issue  Pending Validation  www.example.com
`,
	Run: func(cmd *cobra.Command, args []string) {
		listCertificates()
	},
}

func init() {
	certificateCmd.AddCommand(certificateListCmd)
}

func listCertificates() {
	acm := ACM.New()
	certificates := acm.ListCertificates()

	if len(certificates) > 0 {
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 8, 1, '\t', 0)
		fmt.Fprintln(w, "CERTIFICATE\tTYPE\tSTATUS\tSUBJECT ALTERNATIVE NAMES")

		for _, c := range certificates {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				c.DomainName,
				util.Humanize(c.Type),
				util.Humanize(c.Status),
				strings.Join(c.SubjectAlternativeNames, ", "),
			)
		}

		w.Flush()
	} else {
		console.Info("No certificates found")
	}
}
