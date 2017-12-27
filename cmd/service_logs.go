package cmd

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	lru "github.com/hashicorp/golang-lru"
	CWL "github.com/jpignata/fargate/cloudwatchlogs"
	"github.com/jpignata/fargate/console"
	"github.com/spf13/cobra"
)

const (
	timeFormat          = "2006-01-02 15:04:05"
	timeFormatWithZone  = "2006-01-02 15:04:05 MST"
	logStreamNameFormat = "fargate/%s/%s"
	eventCacheSize      = 10000
)

type Empty struct{}

type GetServiceLogsOperation struct {
	ServiceName     string
	EndTime         time.Time
	Filter          string
	Follow          bool
	LogStreamColors map[string]int
	LogStreamNames  []string
	StartTime       time.Time
	EventCache      *lru.Cache
}

func (o *GetServiceLogsOperation) AddStartTime(rawStartTime string) {
	if rawStartTime != "" {
		o.StartTime = o.parseTime(rawStartTime)
	}
}

func (o *GetServiceLogsOperation) AddEndTime(rawEndTime string) {
	if rawEndTime != "" {
		o.EndTime = o.parseTime(rawEndTime)
	}
}

func (o *GetServiceLogsOperation) AddTasks(tasks []string) {
	for _, task := range tasks {
		logStreamName := fmt.Sprintf(logStreamNameFormat, o.ServiceName, task)
		o.LogStreamNames = append(o.LogStreamNames, logStreamName)
	}
}

func (o *GetServiceLogsOperation) Validate() {
	if o.Follow && !o.EndTime.IsZero() {
		console.ErrorExit(fmt.Errorf("--end-time cannot be specified if following"), "Invalid command line flags")
	}
}

func (o *GetServiceLogsOperation) GetStreamColor(logStreamName string) int {
	if o.LogStreamColors == nil {
		o.LogStreamColors = make(map[string]int)
	}

	if o.LogStreamColors[logStreamName] == 0 {
		o.LogStreamColors[logStreamName] = rand.Intn(256)
	}

	return o.LogStreamColors[logStreamName]
}

func (o *GetServiceLogsOperation) SeenEvent(eventId string) bool {
	if o.EventCache == nil {
		o.EventCache, _ = lru.New(eventCacheSize)
	}

	if !o.EventCache.Contains(eventId) {
		o.EventCache.Add(eventId, Empty{})
		return false
	} else {
		return true
	}
}

func (o *GetServiceLogsOperation) LogGroupName() string {
	return fmt.Sprintf(serviceLogGroupFormat, o.ServiceName)
}

func (o *GetServiceLogsOperation) parseTime(rawTime string) time.Time {
	var t time.Time

	if duration, err := time.ParseDuration(strings.ToLower(rawTime)); err == nil {
		return time.Now().Add(duration)
	}

	if t, err := time.Parse(timeFormat, rawTime); err == nil {
		return t
	}

	if t, err := time.Parse(timeFormatWithZone, rawTime); err == nil {
		return t
	}

	console.ErrorExit(fmt.Errorf("Could not parse %s", rawTime), "Invalid command line flags")

	return t
}

var (
	flagServiceLogsFilter    string
	flagServiceLogsEndTime   string
	flagServiceLogsStartTime string
	flagServiceLogsFollow    bool
	flagServiceLogsTasks     []string
)

var serviceLogsCmd = &cobra.Command{
	Use:   "logs <service name>",
	Short: "View logs from a service",
	Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		rand.Seed(time.Now().UTC().UnixNano())
	},
	Run: func(cmd *cobra.Command, args []string) {
		operation := &GetServiceLogsOperation{
			ServiceName: args[0],
			Filter:      flagServiceLogsFilter,
			Follow:      flagServiceLogsFollow,
		}

		operation.AddTasks(flagServiceLogsTasks)
		operation.AddStartTime(flagServiceLogsStartTime)
		operation.AddEndTime(flagServiceLogsEndTime)

		getServiceLogs(operation)
	},
}

func init() {
	serviceCmd.AddCommand(serviceLogsCmd)

	serviceLogsCmd.Flags().BoolVarP(&flagServiceLogsFollow, "follow", "f", false, "Poll logs and continuously print new events")
	serviceLogsCmd.Flags().StringVar(&flagServiceLogsFilter, "filter", "", "Filter pattern to apply")
	serviceLogsCmd.Flags().StringVar(&flagServiceLogsStartTime, "start", "", "Earliest time to return logs (e.g. -1h, 2018-01-01 09:36:00 EST")
	serviceLogsCmd.Flags().StringVar(&flagServiceLogsEndTime, "end", "", "Latest time to return logs (e.g. 3y, 2021-01-20 12:00:00 EST")
	serviceLogsCmd.Flags().StringSliceVarP(&flagServiceLogsTasks, "tasks", "t", []string{}, "Show logs from specific task (can be specified multiple times)")
}

func getServiceLogs(operation *GetServiceLogsOperation) {
	if operation.Follow {
		followLogs(operation)
	} else {
		getLogs(operation)
	}
}

func followLogs(operation *GetServiceLogsOperation) {
	ticker := time.NewTicker(time.Second)

	if operation.StartTime.IsZero() {
		operation.StartTime = time.Now()
	}

	for {
		getLogs(operation)

		if newStartTime := time.Now().Add(-10 * time.Second); newStartTime.After(operation.StartTime) {
			operation.StartTime = newStartTime
		}

		<-ticker.C
	}
}

func getLogs(operation *GetServiceLogsOperation) {
	cwl := CWL.New(sess)
	input := &CWL.GetLogsInput{
		LogStreamNames: operation.LogStreamNames,
		LogGroupName:   operation.LogGroupName(),
		Filter:         operation.Filter,
		StartTime:      operation.StartTime,
		EndTime:        operation.EndTime,
	}

	for _, logLine := range cwl.GetLogs(input) {
		streamColor := operation.GetStreamColor(logLine.LogStreamName)

		if !operation.SeenEvent(logLine.EventId) {
			console.LogLine(logLine.LogStreamName, logLine.Message, streamColor)
		}
	}
}
