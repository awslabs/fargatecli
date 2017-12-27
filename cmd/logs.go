package cmd

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	lru "github.com/hashicorp/golang-lru"
	CWL "github.com/jpignata/fargate/cloudwatchlogs"
	"github.com/jpignata/fargate/console"
)

const (
	timeFormat          = "2006-01-02 15:04:05"
	timeFormatWithZone  = "2006-01-02 15:04:05 MST"
	logStreamNameFormat = "fargate/%s/%s"
	eventCacheSize      = 10000
)

type Empty struct{}

type GetLogsOperation struct {
	LogGroupName    string
	Namespace       string
	EndTime         time.Time
	Filter          string
	Follow          bool
	LogStreamColors map[string]int
	LogStreamNames  []string
	StartTime       time.Time
	EventCache      *lru.Cache
}

func (o *GetLogsOperation) AddStartTime(rawStartTime string) {
	if rawStartTime != "" {
		o.StartTime = o.parseTime(rawStartTime)
	}
}

func (o *GetLogsOperation) AddEndTime(rawEndTime string) {
	if rawEndTime != "" {
		o.EndTime = o.parseTime(rawEndTime)
	}
}

func (o *GetLogsOperation) AddTasks(tasks []string) {
	for _, task := range tasks {
		logStreamName := fmt.Sprintf(logStreamNameFormat, o.Namespace, task)
		o.LogStreamNames = append(o.LogStreamNames, logStreamName)
	}
}

func (o *GetLogsOperation) Validate() {
	if o.Follow && !o.EndTime.IsZero() {
		console.ErrorExit(fmt.Errorf("--end-time cannot be specified if following"), "Invalid command line flags")
	}
}

func (o *GetLogsOperation) GetStreamColor(logStreamName string) int {
	if o.LogStreamColors == nil {
		o.LogStreamColors = make(map[string]int)
	}

	if o.LogStreamColors[logStreamName] == 0 {
		o.LogStreamColors[logStreamName] = rand.Intn(256)
	}

	return o.LogStreamColors[logStreamName]
}

func (o *GetLogsOperation) SeenEvent(eventId string) bool {
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

func (o *GetLogsOperation) parseTime(rawTime string) time.Time {
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

func GetLogs(operation *GetLogsOperation) {
	rand.Seed(time.Now().UTC().UnixNano())

	if operation.Follow {
		followLogs(operation)
	} else {
		getLogs(operation)
	}
}

func followLogs(operation *GetLogsOperation) {
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

func getLogs(operation *GetLogsOperation) {
	cwl := CWL.New(sess)
	input := &CWL.GetLogsInput{
		LogStreamNames: operation.LogStreamNames,
		LogGroupName:   operation.LogGroupName,
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
