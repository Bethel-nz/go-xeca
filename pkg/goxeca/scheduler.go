package goxeca

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Scheduler struct{}

func NewScheduler() *Scheduler {
	return &Scheduler{}
}

func (s *Scheduler) ParseSchedule(schedule string) (time.Time, time.Duration, error) {
	parts := strings.Fields(schedule)
	if len(parts) != 3 || parts[0] != "in" {
		return time.Time{}, 0, fmt.Errorf("invalid schedule format: expected 'in X seconds/minutes/hours/days'")
	}

	amount, err := strconv.Atoi(parts[1])
	if err != nil {
		return time.Time{}, 0, fmt.Errorf("invalid time amount: %v", err)
	}

	var duration time.Duration
	switch parts[2] {
	case "seconds", "second":
		duration = time.Duration(amount) * time.Second
	case "minutes", "minute":
		duration = time.Duration(amount) * time.Minute
	case "hours", "hour":
		duration = time.Duration(amount) * time.Hour
	case "days", "day":
		duration = time.Duration(amount) * 24 * time.Hour
	default:
		return time.Time{}, 0, fmt.Errorf("unsupported time unit: %s", parts[2])
	}

	nextRunTime := time.Now().Add(duration)
	return nextRunTime, duration, nil
}
