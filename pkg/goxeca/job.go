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

type JobHistory struct {
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
	Output    string
	Status    string
}

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
	ChainedJobs    []string
	Template       bool
	cronEntryID    cron.EntryID
}

type JobTemplate struct {
	Command      string
	Schedule     string
	Recurring    bool
	Priority     int
	Dependencies []string
	MaxRetries   int
	RetryDelay   time.Duration
	Webhook      string
	Timeout      time.Duration
	ChainedJobs  []string
}
