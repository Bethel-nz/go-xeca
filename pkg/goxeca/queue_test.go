package goxeca

import (
	"testing"
)

func TestQueue_Enqueue(t *testing.T) {
	q := NewQueue()

	// Create a sample job
	job1 := &Job{ID: "job1"}
	job2 := &Job{ID: "job2"}

	t.Run("Test Queue", func(t *testing.T) {

			q.Enqueue(job1)

			q.Enqueue(job2)

			if q.Length() != 2 {
				t.Errorf("Expected queue length 2, but got %d", q.Length())
			}


	})

}

// func TestQueue_Dequeue(t *testing.T) {
// 	q := goxeca.NewQueue()

// 	// Create a sample job
// 	job1 := &goxeca.Job{ID: "job1"}
// 	job2 := &goxeca.Job{ID: "job2"}

// 	// Enqueue the jobs
// 	q.Enqueue(job1)
// 	q.Enqueue(job2)

// 	// Dequeue the first job and check if it's the correct job
// 	dequeuedJob := q.Dequeue()
// 	if dequeuedJob.ID != "job1" {
// 		t.Errorf("Expected job1 to be dequeued, but got %s", dequeuedJob.ID)
// 	}

// 	// Dequeue the second job and check if it's the correct job
// 	dequeuedJob = q.Dequeue()
// 	if dequeuedJob.ID != "job2" {
// 		t.Errorf("Expected job2 to be dequeued, but got %s", dequeuedJob.ID)
// 	}

// 	// Ensure queue is empty after dequeuing both jobs
// 	if q.Length() != 0 {
// 		t.Errorf("Expected queue length 0, but got %d", q.Length())
// 	}
// }

// func TestQueue_Length(t *testing.T) {
// 	q := goxeca.NewQueue()

// 	// Initially, the queue should be empty
// 	if q.Length() != 0 {
// 		t.Errorf("Expected queue length 0, but got %d", q.Length())
// 	}

// 	// Enqueue a job and check length
// 	job1 := &goxeca.Job{ID: "job1"}
// 	q.Enqueue(job1)
// 	if q.Length() != 1 {
// 		t.Errorf("Expected queue length 1, but got %d", q.Length())
// 	}

// 	// Dequeue the job and check length again
// 	q.Dequeue()
// 	if q.Length() != 0 {
// 		t.Errorf("Expected queue length 0, but got %d", q.Length())
// 	}
// }
