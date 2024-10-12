package main_test

import (
	"testing"
	"time"

	"github.com/bethel-nz/goxeca/pkg/goxeca"
	"github.com/charmbracelet/log"
)

func TestXeca(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	manager := goxeca.NewManager()
	manager.Start()
	defer manager.Stop()
	t.Run("Non-recurring job", func(t *testing.T) {
		job, err := manager.AddJob("echo 'Hello world'", "in 1 second", false, 1, []string{}, 0, time.Duration(0), "", time.Duration(0), []string{})
		if err != nil {
			t.Fatalf("Failed to add job: %v", err)
		}
		if job.ID == "" {
			t.Error("Expected non-empty job ID")
		}

		time.Sleep(2 * time.Second)

		updatedJob, err := manager.GetJob(job.ID)
		if err == nil {
			t.Errorf("Expected job to be removed, but it still exists")
		}

		if updatedJob != nil && updatedJob.Status != "completed" {
			t.Errorf("Expected job status to be 'completed', got '%s'", updatedJob.Status)
		}
	})
	t.Run("Recurring job", func(t *testing.T) {
		job, err := manager.AddJob("echo 'Recurring'", "in 1 second", true, 2, []string{}, 3, time.Second, "", time.Duration(0), []string{})
		if err != nil {
			t.Fatalf("Failed to add recurring job: %v", err)
		}

		time.Sleep(3500 * time.Millisecond)

		updatedJob, err := manager.GetJob(job.ID)
		if err != nil {
			t.Errorf("Failed to get recurring job: %v", err)
		} else if updatedJob.ExecutionCount < 2 {
			t.Errorf("Expected at least 2 executions of recurring job, got %d", updatedJob.ExecutionCount)
		}

		err = manager.PauseJob(job.ID)
		if err != nil {
			t.Errorf("Failed to pause job: %v", err)
		}

		time.Sleep(2 * time.Second)

		updatedJob, _ = manager.GetJob(job.ID)
		if updatedJob.ExecutionCount > 3 {
			t.Errorf("Job executed while paused")
		}

		err = manager.ResumeJob(job.ID)
		if err != nil {
			t.Errorf("Failed to resume job: %v", err)
		}
	})
	t.Run("Job with dependencies", func(t *testing.T) {
		job1, err := manager.AddJob("echo 'Job 1'", "in 1 second", false, 1, []string{}, 0, time.Second, "", time.Duration(0), []string{})
		if err != nil {
			t.Fatalf("Failed to add job1: %v", err)
		}

		job2, err := manager.AddJob("echo 'Job 2'", "in 1 second", false, 1, []string{job1.ID}, 0, time.Second, "", time.Duration(0), []string{})
		if err != nil {
			t.Fatalf("Failed to add job2: %v", err)
		}

		// Wait for both jobs to complete
		time.Sleep(3 * time.Second)

		// Check job1 status
		updatedJob1, err := manager.GetJob(job1.ID)
		if err != nil {
			if err.Error() != "job with ID "+job1.ID+" not found" {
				t.Errorf("Unexpected error for job1: %v", err)
			}
		} else if updatedJob1.Status != "completed" {
			t.Errorf("Expected job1 status to be completed, got %s", updatedJob1.Status)
		}

		// Check job2 status
		updatedJob2, err := manager.GetJob(job2.ID)
		if err != nil {
			if err.Error() != "job with ID "+job2.ID+" not found" {
				t.Errorf("Unexpected error for job2: %v", err)
			}
		} else if updatedJob2.Status != "completed" {
			t.Errorf("Expected job2 status to be completed, got %s", updatedJob2.Status)
		}

		// If both jobs are not found, it means they've been removed after completion
		if updatedJob1 == nil && updatedJob2 == nil {
			t.Log("Both jobs completed and were removed as expected")
		}
	})

	log.Info("All tests completed")
}
