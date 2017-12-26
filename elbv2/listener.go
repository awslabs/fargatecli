package elbv2

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	awselbv2 "github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/jpignata/fargate/console"
)

type CreateListenerInput struct {
	Port                  int64
	Protocol              string
	CertificateArns       []string
	LoadBalancerArn       string
	DefaultTargetGroupArn string
}

type Rule struct {
	Type           string
	Value          string
	TargetGroupArn string
}

func (r *Rule) String() string {
	if r.Value != "" {
		return fmt.Sprintf("%s=%s", r.Type, r.Value)
	} else {
		return fmt.Sprintf("%s", r.Type)
	}
}

type Listener struct {
	Arn      string
	Port     int64
	Protocol string
}

func (l *Listener) String() string {
	return fmt.Sprintf("%s:%d", l.Protocol, l.Port)
}

type DirRange []int64

func (a DirRange) Len() int           { return len(a) }
func (a DirRange) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a DirRange) Less(i, j int) bool { return a[i] < a[j] }

func (input *CreateListenerInput) SetCertificateArns(arns []string) {
	input.CertificateArns = arns
}

func (elbv2 *ELBV2) CreateListener(i *CreateListenerInput) string {
	console.Debug("Creating ELB listener [%s:%s]", i.Protocol, i.Port)

	action := &awselbv2.Action{
		TargetGroupArn: aws.String(i.DefaultTargetGroupArn),
		Type:           aws.String(awselbv2.ActionTypeEnumForward),
	}

	input := &awselbv2.CreateListenerInput{
		Port:            aws.Int64(i.Port),
		Protocol:        aws.String(i.Protocol),
		LoadBalancerArn: aws.String(i.LoadBalancerArn),
		DefaultActions:  []*awselbv2.Action{action},
	}

	if len(i.CertificateArns) > 0 {
		certificates := []*awselbv2.Certificate{}

		for _, certificateArn := range i.CertificateArns {
			certificates = append(certificates,
				&awselbv2.Certificate{
					CertificateArn: aws.String(certificateArn),
				},
			)
		}

		input.SetCertificates(certificates)
	}

	resp, err := elbv2.svc.CreateListener(input)

	if err != nil || len(resp.Listeners) != 1 {
		console.ErrorExit(err, "Could not create ELB listener")
	}

	return aws.StringValue(resp.Listeners[0].ListenerArn)
}

func (elbv2 *ELBV2) ModifyLoadBalancerDefaultAction(lbArn, targetGroupArn string) {
	console.Debug("Setting ELB listener default action")
	listeners := elbv2.GetListeners(lbArn)
	action := &awselbv2.Action{
		TargetGroupArn: aws.String(targetGroupArn),
		Type:           aws.String(awselbv2.ActionTypeEnumForward),
	}

	for _, listener := range listeners {
		elbv2.svc.ModifyListener(
			&awselbv2.ModifyListenerInput{
				ListenerArn:    aws.String(listener.Arn),
				DefaultActions: []*awselbv2.Action{action},
			},
		)
	}
}

func (elbv2 *ELBV2) AddRule(lbArn, targetGroupArn string, rule Rule) {
	console.Debug("Adding ELB listener rule [%s=%s]", rule.Type, rule.Value)
	var ruleType string

	if rule.Type == "HOST" {
		ruleType = "host-header"
	} else {
		ruleType = "path-pattern"
	}

	ruleCondition := &awselbv2.RuleCondition{
		Field:  aws.String(ruleType),
		Values: aws.StringSlice([]string{rule.Value}),
	}

	listeners := elbv2.GetListeners(lbArn)

	for _, listener := range listeners {
		highestPriority := elbv2.GetHighestPriorityFromListener(listener.Arn)
		priority := highestPriority + 10
		action := &awselbv2.Action{
			TargetGroupArn: aws.String(targetGroupArn),
			Type:           aws.String(awselbv2.ActionTypeEnumForward),
		}

		elbv2.svc.CreateRule(
			&awselbv2.CreateRuleInput{
				Priority:    aws.Int64(priority),
				ListenerArn: aws.String(listener.Arn),
				Actions:     []*awselbv2.Action{action},
				Conditions:  []*awselbv2.RuleCondition{ruleCondition},
			},
		)
	}
}

func (elbv2 *ELBV2) DescribeRules(listenerArn string) []Rule {
	var rules []Rule

	resp, err := elbv2.svc.DescribeRules(
		&awselbv2.DescribeRulesInput{
			ListenerArn: aws.String(listenerArn),
		},
	)

	if err != nil {
		console.ErrorExit(err, "Could not describe ELB rules")
	}

	for _, r := range resp.Rules {
		for _, c := range r.Conditions {
			var field string

			switch aws.StringValue(c.Field) {
			case "host-header":
				field = "HOST"
			case "path-pattern":
				field = "PATH"
			}

			for _, v := range c.Values {
				rule := Rule{
					TargetGroupArn: aws.StringValue(r.Actions[0].TargetGroupArn),
					Type:           field,
					Value:          aws.StringValue(v),
				}

				rules = append(rules, rule)
			}
		}

		if aws.BoolValue(r.IsDefault) == true {
			rule := Rule{
				TargetGroupArn: aws.StringValue(r.Actions[0].TargetGroupArn),
				Type:           "DEFAULT",
			}

			rules = append(rules, rule)
		}
	}

	return rules
}

func (elbv2 *ELBV2) GetHighestPriorityFromListener(listenerArn string) int64 {
	priorities := DirRange{}

	resp, err := elbv2.svc.DescribeRules(
		&awselbv2.DescribeRulesInput{
			ListenerArn: aws.String(listenerArn),
		},
	)

	if err != nil {
		console.ErrorExit(err, "Could not retrieve ELB listener rules")
	}

	for _, rule := range resp.Rules {
		priority, _ := strconv.ParseInt(*rule.Priority, 10, 64)
		priorities = append(priorities, priority)
	}

	sort.Sort(priorities)

	return priorities[len(priorities)-1]
}

func (elbv2 *ELBV2) GetListeners(lbArn string) []Listener {
	var listeners []Listener

	input := &awselbv2.DescribeListenersInput{
		LoadBalancerArn: aws.String(lbArn),
	}

	err := elbv2.svc.DescribeListenersPages(
		input,
		func(resp *awselbv2.DescribeListenersOutput, lastPage bool) bool {
			for _, listener := range resp.Listeners {
				listeners = append(
					listeners,
					Listener{
						Arn:      aws.StringValue(listener.ListenerArn),
						Port:     aws.Int64Value(listener.Port),
						Protocol: aws.StringValue(listener.Protocol),
					},
				)
			}

			return true
		},
	)

	if err != nil {
		console.ErrorExit(err, "Could not retrieve ELB listeners")
	}

	return listeners
}
