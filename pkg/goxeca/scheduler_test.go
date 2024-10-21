package goxeca

import (
	"testing"
	"time"

	"github.com/bethel-nz/goxeca/internal/assert"
)

func TestScheduler(t *testing.T) {
	t.Run("ParseSchedule", func(t *testing.T) {
		scheduler := NewScheduler()

		tests := []struct {
			schedule string
			wantErr  bool
		}{
			{"in 5 minutes", false},
			{"every 1 hour", false},
			{"* * * * *", false},
			{"invalid schedule", true},
		}

		for _, tt := range tests {
			t.Run(tt.schedule, func(t *testing.T) {
				_, _, err := scheduler.ParseSchedule(tt.schedule)
				if tt.wantErr {
					assert.NotNil(t, err)
				} else {
					assert.NilError(t, err)
				}
			})
		}
	})

	t.Run("GetNextRunTime", func(t *testing.T) {
		scheduler := NewScheduler()

		tests := []struct {
			schedule string
		}{
			{"every 1 hour"},
			{"* * * * *"},
		}

		for _, tt := range tests {
			t.Run(tt.schedule, func(t *testing.T) {
				nextRunTime := scheduler.GetNextRunTime(tt.schedule)
				assert.False(t, nextRunTime.IsZero())
				assert.True(t, nextRunTime.After(time.Now()))
			})
		}
	})
}
