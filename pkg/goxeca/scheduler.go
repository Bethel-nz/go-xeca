package goxeca

import (
	"fmt"
	"time"

	"github.com/olebedev/when"
	"github.com/olebedev/when/rules/en"
)

type Scheduler struct {
	parser *when.Parser
}

func NewScheduler() *Scheduler {
	w := when.New(nil)
	w.Add(en.All...)

	return &Scheduler{
		parser: w,
	}
}

func (s *Scheduler) ParseSchedule(schedule string) (time.Time, time.Duration, error) {
	result, err := s.parser.Parse(schedule, time.Now())
	if err != nil {
		return time.Time{}, 0, err
	}
	if result == nil {
		return time.Time{}, 0, fmt.Errorf("unable to parse schedule: %s", schedule)
	}

	nextRunTime := result.Time
	duration := time.Until(nextRunTime)

	return nextRunTime, duration, nil
}
