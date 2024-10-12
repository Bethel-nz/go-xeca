package goxeca

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	parser cron.Parser
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		parser: cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
	}
}

func (s *Scheduler) ParseSchedule(schedule string) (time.Time, time.Duration, error) {
	if strings.HasPrefix(schedule, "in ") {
		return s.parseRelativeSchedule(schedule)
	}
	return s.parseCronSchedule(schedule)
}

func (s *Scheduler) parseRelativeSchedule(schedule string) (time.Time, time.Duration, error) {
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

func (s *Scheduler) parseCronSchedule(schedule string) (time.Time, time.Duration, error) {
	sched, err := s.parser.Parse(schedule)
	if err != nil {
		return time.Time{}, 0, fmt.Errorf("invalid cron schedule: %v", err)
	}

	now := time.Now()
	nextRunTime := sched.Next(now)
	duration := nextRunTime.Sub(now)

	return nextRunTime, duration, nil
}

func (s *Scheduler) GetNextRunTime(schedule string) time.Time {
	if strings.HasPrefix(schedule, "in ") {
		nextRunTime, _, _ := s.parseRelativeSchedule(schedule)
		return nextRunTime
	}

	sched, err := s.parser.Parse(schedule)
	if err != nil {
		return time.Now().AddDate(100, 0, 0)
	}

	return sched.Next(time.Now())
}
