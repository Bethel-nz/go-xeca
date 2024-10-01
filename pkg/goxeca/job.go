package goxeca

import (
	"time"
)

const (
	JobStatusPending   = "pending"
	JobStatusRunning   = "running"
	JobStatusCompleted = "completed"
	JobStatusFailed    = "failed"
)

type Job struct {
	ID             string
	Command        string
	Schedule       string
	NextRunTime    time.Time
	Status         string
	Output         string
	Recurring      bool
	ExecutionCount int
	Duration       time.Duration // Add this field
}
