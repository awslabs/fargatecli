package cmd

import (
	"fmt"
	"math/rand"
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
)

var (
	filter         string
	endTime        time.Time
	startTime      time.Time
	limit          int64
	follow         bool
	startTimeRaw   string
	endTimeRaw     string
	streamColors   map[string]int
	logStreamNames []string
	tasks          []string
)

var eventCache, _ = lru.New(10000)

var serviceLogsCmd = &cobra.Command{
	Use:   "logs <service name>",
	Short: "View logs from a service", Args: cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		streamColors = make(map[string]int)
		rand.Seed(time.Now().UTC().UnixNano())

		if startTimeRaw != "" {
			startTime = parseTime(startTimeRaw)
		}

		if endTimeRaw != "" {
			endTime = parseTime(endTimeRaw)
		}

		if follow && !endTime.IsZero() {
			console.ErrorExit(fmt.Errorf("--end-time cannot be specified if following"), "Invalid command line flags")
		}

		for _, task := range tasks {
			logStreamName := fmt.Sprintf(logStreamNameFormat, args[0], task)
			logStreamNames = append(logStreamNames, logStreamName)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		getServiceLogs(args[0])
	},
}

func init() {
	serviceCmd.AddCommand(serviceLogsCmd)

	serviceLogsCmd.Flags().BoolVarP(&follow, "follow", "f", false, "Poll logs and continuously print new events")
	serviceLogsCmd.Flags().StringVar(&filter, "filter", "", "Filter pattern to apply")
	serviceLogsCmd.Flags().StringVar(&startTimeRaw, "start", "", "Earliest time to return logs (e.g. -1h, 2018-01-01 09:36:00 EST")
	serviceLogsCmd.Flags().StringVar(&endTimeRaw, "end", "", "Latest time to return logs (e.g. 3y, 2021-01-20 12:00:00 EST")
	serviceLogsCmd.Flags().StringSliceVarP(&tasks, "tasks", "t", []string{}, "Show logs from specific task (can be specified multiple times)")
}

func parseTime(timeRaw string) time.Time {
	var t time.Time

	if duration, err := time.ParseDuration(timeRaw); err == nil {
		return time.Now().Add(duration)
	}

	if t, err := time.Parse(timeFormat, timeRaw); err == nil {
		return t
	}

	if t, err := time.Parse(timeFormatWithZone, timeRaw); err == nil {
		return t
	}

	console.ErrorExit(fmt.Errorf("Could not parse %s", timeRaw), "Invalid command line flags")

	return t
}

func getServiceLogs(serviceName string) {
	logGroupName := fmt.Sprintf(logGroupFormat, serviceName)

	if follow {
		followLogs(logGroupName)
	} else {
		getLogs(logGroupName)
	}
}

func followLogs(logGroupName string) {
	ticker := time.NewTicker(time.Second)
	followStartTime := time.Now()

	if startTime.IsZero() {
		startTime = time.Now()
	}

	endTime = time.Time{}

	for {
		getLogs(logGroupName)

		if newStartTime := time.Now().Add(-10 * time.Second); newStartTime.After(followStartTime) {
			startTime = newStartTime
		}

		<-ticker.C
	}
}

func getLogs(logGroupName string) {
	var empty struct{}

	cwl := CWL.New(sess)
	input := &CWL.GetLogsInput{
		LogStreamNames: logStreamNames,
		LogGroupName:   logGroupName,
		Filter:         filter,
		StartTime:      startTime,
		EndTime:        endTime,
	}

	for _, logLine := range cwl.GetLogs(input) {
		if streamColors[logLine.LogStreamName] == 0 {
			streamColors[logLine.LogStreamName] = rand.Intn(256)
		}

		if !eventCache.Contains(logLine.EventId) {
			console.LogLine(logLine.LogStreamName, logLine.Message, streamColors[logLine.LogStreamName])
			eventCache.Add(logLine.EventId, empty)
		}
	}
}
