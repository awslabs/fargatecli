package cmd

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/jpignata/fargate/elbv2"
	"github.com/spf13/cobra"
	"golang.org/x/time/rate"
)

type lbListOperation struct {
	elbv2  elbv2.Client
	output Output
}

func (o lbListOperation) find() (elbv2.LoadBalancers, error) {
	var wg sync.WaitGroup

	o.output.Debug("Describing Load Balancers [API=elbv2 Action=DescribeLoadBalancers]")
	loadBalancers, err := o.elbv2.DescribeLoadBalancers()

	if err != nil {
		return elbv2.LoadBalancers{}, err
	}

	errs := make(chan error)
	done := make(chan bool)
	limiter := rate.NewLimiter(describeRequestLimitRate, 1)

	for i := 0; i < len(loadBalancers); i++ {
		wg.Add(1)

		go func(index int) {
			defer wg.Done()

			if err := limiter.Wait(context.Background()); err == nil {
				o.output.Debug("Describing Listeners [API=elbv2 Action=DescribeListeners LoadBalancerArn=%s]", loadBalancers[index].ARN)
				listeners, err := o.elbv2.DescribeListeners(loadBalancers[index].ARN)

				if err != nil {
					errs <- err
				}

				loadBalancers[index].Listeners = listeners
			}
		}(i)
	}

	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case err := <-errs:
		return elbv2.LoadBalancers{}, err
	case <-done:
		return loadBalancers, nil
	}
}

func (o lbListOperation) execute() {
	loadBalancers, err := o.find()

	if err != nil {
		o.output.Fatal(err, "Could not list load balancers")
		return
	}

	if len(loadBalancers) == 0 {
		o.output.Info("No load balancers found")
		return
	}

	sort.Slice(loadBalancers, func(i, j int) bool {
		return loadBalancers[i].Name < loadBalancers[j].Name
	})

	rows := [][]string{
		[]string{"NAME", "TYPE", "STATUS", "DNS NAME", "PORTS"},
	}

	for _, loadBalancer := range loadBalancers {
		rows = append(rows,
			[]string{
				loadBalancer.Name,
				Titleize(loadBalancer.Type),
				Titleize(loadBalancer.State),
				loadBalancer.DNSName,
				fmt.Sprintf("%s", loadBalancer.Listeners),
			},
		)
	}

	o.output.Table("", rows)
}

var lbListCmd = &cobra.Command{
	Use:   "list",
	Short: "List load balancers",
	Run: func(cmd *cobra.Command, args []string) {
		lbListOperation{
			elbv2:  elbv2.New(sess),
			output: output,
		}.execute()
	},
}

func init() {
	lbCmd.AddCommand(lbListCmd)
}
