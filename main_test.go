package main_test

import (
	"testing"
	"time"

	"github.com/bethel-nz/goxeca/internal/assert"
	"github.com/bethel-nz/goxeca/pkg/goxeca"
	"github.com/charmbracelet/log"
)

func TestXeca(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	config := goxeca.ManagerConfig{
		MaxConcurrent: 20,
		RedisAddr:     "localhost:6379",
		RedisPassword: "",
		RedisDB:       0,
		JobQueueSize:  200,
	}
	manager := goxeca.NewManager(config)
	manager.Start()
	defer manager.Stop()

	t.Run("Non-recurring job", func(t *testing.T) {
		job, err := manager.AddJob("echo 'Hello world'", "in 1 second", 1, []string{}, 0, time.Duration(0), "", time.Second*5)
		assert.NilError(t, err)
		assert.NotNil(t, job)

		time.Sleep(3 * time.Second)

		updatedJob, err := manager.GetJob(job.ID)
		if err != nil {
			assert.StringContains(t, err.Error(), "not found")
		} else {
			assert.Equal(t, updatedJob.Status, goxeca.JobStatusCompleted)
		}
	})

	t.Run("Recurring job", func(t *testing.T) {
		job, err := manager.AddJob("echo 'Recurring'", "*/1 * * * *", 2, []string{}, 3, time.Second, "", time.Second*5)
		assert.NilError(t, err)

		time.Sleep(65 * time.Second) // Wait for at least one execution

		updatedJob, err := manager.GetJob(job.ID)
		assert.NilError(t, err)
		t.Logf("Execution count after initial run: %d", updatedJob.ExecutionCount)
		assert.True(t, updatedJob.ExecutionCount >= 1)
		if updatedJob.ExecutionCount < 1 {
			t.Error("Expected at least one execution")
		}

		err = manager.PauseJob(job.ID)
		assert.NilError(t, err)

		pausedJob, err := manager.GetJob(job.ID)
		assert.NilError(t, err)
		assert.Equal(t, pausedJob.Status, goxeca.JobStatusPaused)

		time.Sleep(65 * time.Second)

		pausedJobAfterWait, err := manager.GetJob(job.ID)
		assert.NilError(t, err)
		t.Logf("Execution count after pause: %d", pausedJobAfterWait.ExecutionCount)
		assert.Equal(t, pausedJobAfterWait.ExecutionCount, updatedJob.ExecutionCount)
		if pausedJobAfterWait.ExecutionCount != updatedJob.ExecutionCount {
			t.Error("Job should not execute while paused")
		}

		err = manager.ResumeJob(job.ID)
		assert.NilError(t, err)

		time.Sleep(65 * time.Second)

		resumedJob, err := manager.GetJob(job.ID)
		assert.NilError(t, err)
		t.Logf("Execution count after resume: %d", resumedJob.ExecutionCount)
		assert.True(t, resumedJob.ExecutionCount > pausedJobAfterWait.ExecutionCount)
		if resumedJob.ExecutionCount <= pausedJobAfterWait.ExecutionCount {
			t.Error("Job should execute after being resumed")
		}
	})

	t.Run("Job with dependencies", func(t *testing.T) {
		job1, err := manager.AddJob("echo 'Job 1'", "in 1 second", 1, []string{}, 0, time.Second, "", time.Second*5)
		assert.NilError(t, err)

		job2, err := manager.AddJob("echo 'Job 2'", "in 1 second", 1, []string{job1.ID}, 0, time.Second, "", time.Second*5)
		assert.NilError(t, err)

		time.Sleep(5 * time.Second)

		updatedJob1, err := manager.GetJob(job1.ID)
		if err != nil {
			assert.StringContains(t, err.Error(), "not found")
		} else {
			assert.Equal(t, updatedJob1.Status, goxeca.JobStatusCompleted)
		}

		updatedJob2, err := manager.GetJob(job2.ID)
		if err != nil {
			assert.StringContains(t, err.Error(), "not found")
		} else {
			assert.Equal(t, updatedJob2.Status, goxeca.JobStatusCompleted)
		}

		if updatedJob1 == nil && updatedJob2 == nil {
			t.Log("Both jobs completed and were removed as expected")
		}
	})

	log.Info("All tests completed")
}
