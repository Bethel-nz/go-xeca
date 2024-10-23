package goxeca

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron *cron.Cron
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		cron: cron.New(),
	}
}

func (s *Scheduler) ParseSchedule(schedule string) (time.Time, time.Duration, error) {
	if strings.HasPrefix(schedule, "every ") || strings.HasPrefix(schedule, "in ") {
		return s.parseHRSchedule(schedule)

	} else {
		return s.parseCronSchedule(schedule)
	}
}

func (s *Scheduler) parseHRSchedule(schedule string) (time.Time, time.Duration, error) { //Human Readable Schedule " HR Schedule "
	parts := strings.Fields(schedule)
	if len(parts) != 3 {
		return time.Time{}, 0, fmt.Errorf("invalid schedule format: %s", schedule)
	}

	quantity, err := strconv.Atoi(parts[1])
	if err != nil {
		return time.Time{}, 0, fmt.Errorf("invalid quantity in schedule: %s", schedule)
	}

	var duration time.Duration
	switch parts[2] {
	case "second", "seconds":
		duration = time.Duration(quantity) * time.Second
	case "minute", "minutes":
		duration = time.Duration(quantity) * time.Minute
	case "hour", "hours":
		duration = time.Duration(quantity) * time.Hour
	case "day", "days":
		duration = time.Duration(quantity) * 24 * time.Hour
	default:
		return time.Time{}, 0, fmt.Errorf("invalid time unit in schedule: %s", schedule)
	}

	nextRunTime := time.Now().Add(duration)
	return nextRunTime, duration, nil
}

func (s *Scheduler) parseCronSchedule(schedule string) (time.Time, time.Duration, error) {
	cronSchedule, err := cron.ParseStandard(schedule)
	if err != nil {
		return time.Time{}, 0, fmt.Errorf("invalid cron schedule: %v", err)
	}

	nextRunTime := cronSchedule.Next(time.Now())
	duration := time.Until(nextRunTime)
	return nextRunTime, duration, nil
}

func (s *Scheduler) GetNextRunTime(schedule string) time.Time {
	nextRunTime, _, _ := s.ParseSchedule(schedule)
	return nextRunTime
}

func (s *Scheduler) IsCronSchedule(schedule string) bool {
	_, err := cron.ParseStandard(schedule)
	return err == nil
}
