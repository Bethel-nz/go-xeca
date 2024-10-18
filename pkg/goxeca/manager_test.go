package goxeca

import (
	"testing"
	"time"

	"github.com/bethel-nz/goxeca/internal/assert"
)

func TestManager(t *testing.T) {
	config := ManagerConfig{
		MaxConcurrent: 5,
		RedisAddr:     "localhost:6379",
		RedisPassword: "",
		RedisDB:       3,
		JobQueueSize:  100,
	}

	t.Run("AddJob", func(t *testing.T) {
		manager := NewManager(config)

		job, err := manager.AddJob("echo 'test'", "in 5 minutes", 1, nil, 3, time.Second, "", time.Minute)
		assert.NilError(t, err)
		assert.NotNil(t, job)
		assert.True(t, job.ID != "")
		assert.Equal(t, job.Status, JobStatusPending)
		assert.False(t, job.Recurring)

		// Test adding a recurring job
		recurringJob, err := manager.AddJob("echo 'recurring'", "every 1 hour", 1, nil, 3, time.Second, "", time.Minute)
		assert.NilError(t, err)
		assert.NotNil(t, recurringJob)
		assert.True(t, recurringJob.Recurring)
	})

	t.Run("ExecuteJob", func(t *testing.T) {
		manager := NewManager(config)

		job := &Job{
			ID:      "test-job",
			Command: "echo 'test execution'",
			Status:  JobStatusPending,
			Timeout: time.Second * 5,
		}

		// Store the job before executing it
		err := manager.db.StoreJob(job)
		assert.NilError(t, err)

		manager.executeJob(job)

		// Wait a bit for the job to complete
		time.Sleep(time.Second)

		updatedJob, err := manager.GetJob(job.ID)
		assert.NilError(t, err)
		assert.Equal(t, updatedJob.Status, JobStatusCompleted)
		assert.True(t, len(updatedJob.History) > 0)
	})

	t.Run("StopJob", func(t *testing.T) {
		manager := NewManager(config)

		job, err := manager.AddJob("sleep 10", "in 1 minute", 1, nil, 0, 0, "", time.Minute)
		assert.NilError(t, err)

		err = manager.StopJob(job.ID)
		assert.NilError(t, err)

		stoppedJob, err := manager.GetJob(job.ID)
		assert.NilError(t, err)
		assert.Equal(t, stoppedJob.Status, JobStatusCancelled)
	})
}
