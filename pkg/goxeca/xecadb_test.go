package goxeca

import (
	"testing"
	"time"

	"github.com/bethel-nz/goxeca/internal/assert"
)

func TestXecaDB(t *testing.T) {
	t.Run("StoreAndGetJob", func(t *testing.T) {
		db := NewXecaDB("localhost:6379", "", 3)
		defer db.Close()

		job := &Job{
			ID:      "test-job",
			Command: "echo 'test'",
			Status:  JobStatusPending,
		}

		err := db.StoreJob(job)
		assert.NilError(t, err)

		retrievedJob, err := db.GetJob(job.ID)
		assert.NilError(t, err)
		assert.Equal(t, retrievedJob.ID, job.ID)
	})

	t.Run("UpdateJobStatus", func(t *testing.T) {
		db := NewXecaDB("localhost:6379", "", 3)
		defer db.Close()

		job := &Job{
			ID:      "test-job-status",
			Command: "echo 'test'",
			Status:  JobStatusPending,
		}

		err := db.StoreJob(job)
		assert.NilError(t, err)

		err = db.UpdateJobStatus(job.ID, JobStatusRunning)
		assert.NilError(t, err)

		updatedJob, err := db.GetJob(job.ID)
		assert.NilError(t, err)
		assert.Equal(t, updatedJob.Status, JobStatusRunning)
	})

	t.Run("GetPendingJobs", func(t *testing.T) {
		db := NewXecaDB("localhost:6379", "", 3)
		defer db.Close()

		// Clear all existing jobs
		keys, _ := db.client.Keys(db.ctx, "job:*").Result()
		for _, key := range keys {
			db.client.Del(db.ctx, key)
		}

		// Add a pending job
		pendingJob := &Job{
			ID:          "pending-job",
			Command:     "echo 'pending'",
			Status:      JobStatusPending,
			NextRunTime: time.Now().Add(-time.Minute),
			Paused:      false,
		}
		err := db.StoreJob(pendingJob)
		assert.NilError(t, err)

		// Add a non-pending job
		nonPendingJob := &Job{
			ID:      "non-pending-job",
			Command: "echo 'not pending'",
			Status:  JobStatusRunning,
		}
		err = db.StoreJob(nonPendingJob)
		assert.NilError(t, err)

		// Add a paused job
		pausedJob := &Job{
			ID:      "paused-job",
			Command: "echo 'paused'",
			Status:  JobStatusPending,
			Paused:  true,
		}
		err = db.StoreJob(pausedJob)
		assert.NilError(t, err)

		pendingJobs, err := db.GetPendingJobs()
		assert.NilError(t, err)
		assert.Equal(t, len(pendingJobs), 1)
		assert.Equal(t, pendingJobs[0].ID, pendingJob.ID)
	})
}
