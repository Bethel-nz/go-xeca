package goxeca

import (
	"time"

	"github.com/robfig/cron/v3"
)

const (
	JobStatusPending   = "pending"
	JobStatusRunning   = "running"
	JobStatusCompleted = "completed"
	JobStatusFailed    = "failed"
	JobStatusPaused    = "paused"
	JobStatusCancelled = "cancelled"
)

// Defines a Go struct to represent the history of a job, including start and end times, duration, output, and status.
type JobHistory struct {
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
	Output    string
	Status    string
}

// Defines a Go struct representing a job with scheduling parameters, execution history, and dependencies.
type Job struct {
	ID             string
	Command        string
	Schedule       string
	NextRunTime    time.Time
	Status         string
	Output         string
	Recurring      bool
	ExecutionCount int
	Duration       time.Duration
	Priority       int
	Dependencies   []string
	RetryCount     int
	MaxRetries     int
	RetryDelay     time.Duration
	Paused         bool
	History        []JobHistory
	Webhook        string
	Timeout        time.Duration
	cronEntryID    cron.EntryID
	ActiveRuns     int
	TotalRuns      int
}
