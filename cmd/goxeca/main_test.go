package main_test

import (
	"sync"
	"testing"

	"github.com/bethel-nz/goxeca/pkg/goxeca"
	"github.com/charmbracelet/log"
)

func TestXeca(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	manager := goxeca.NewManager()

	t.Run("Non-recurring job", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)

		job, err := manager.AddJob("echo 'Hello, World!'", "in 2 seconds", false)
		if err != nil {
			t.Fatalf("Failed to add job: %v", err)
		}
		if job.ID == "" {
			t.Error("Expected non-empty job ID")
		}

		go func() {
			defer wg.Done()
			manager.Start()
			updatedJob, err := manager.GetJob(job.ID)
			if err == nil {
				t.Log(updatedJob)
			} else {
				t.Log(err)
			}
		}()

		// Wait for the job to complete or timeout

	})

	log.Info("All tests completed")

}

// package main_test

// import (
// 	"sync"
// 	"testing"

// 	"github.com/bethel-nz/goxeca/pkg/goxeca"
// 	"github.com/charmbracelet/log"
// )

// func TestXeca(t *testing.T) {
// 	log.SetLevel(log.DebugLevel)
// 	manager := goxeca.NewManager()

// 	t.Run("Non-recurring job", func(t *testing.T) {
// 		var wg sync.WaitGroup
// 		wg.Add(1)

// 		job, err := manager.AddJob("echo 'Hello, World!'", "in 2 seconds", false)
// 		if err != nil {
// 			t.Fatalf("Failed to add job: %v", err)
// 		}
// 		if job.ID == "" {
// 			t.Error("Expected non-empty job ID")
// 		}

// 		go func() {
// 			defer wg.Done()
// 			// Wait for a short time to allow the job to be scheduled
// 			manager.Start()

// 			if job.Status != "completed" {
// 				t.Errorf("Expected job status to be 'completed', got '%s'", job.Status)
// 			}

// 		}()
// 		wg.Wait()
// 	})
// 	// Test 2: Recurring job
// 	// t.Run("Recurring job", func(t *testing.T) {
// 	// 	wg.Add(1)
// 	// 	jobExecutions := make(chan bool, 3)

// 	// 	job2, err := manager.AddJob("date", "in 1 second", true)
// 	// 	if err != nil {
// 	// 		t.Fatalf("Failed to add job: %v", err)
// 	// 	}
// 	// 	if job2.ID == "" {
// 	// 		t.Error("Expected non-empty job ID")
// 	// 	}

// 	// 	go func() {
// 	// 		defer wg.Done()
// 	// 		executionCount := 0
// 	// 		for {
// 	// 			select {
// 	// 			case <-jobExecutions:
// 	// 				executionCount++
// 	// 				if executionCount == 3 {
// 	// 					log.Debug("Recurring job executed 3 times")
// 	// 					return
// 	// 				}
// 	// 			case <-time.After(5 * time.Second):
// 	// 				t.Error("Recurring job didn't execute 3 times in time")
// 	// 				return
// 	// 			}
// 	// 		}
// 	// 	}()

// 	// 	// Simulate job executions (replace this with actual job execution signals)
// 	// 	for i := 0; i < 3; i++ {
// 	// 		time.AfterFunc(time.Duration(i+1)*time.Second, func() { jobExecutions <- true })
// 	// 	}
// 	// })

// 	// Test 3: Multiple jobs
// 	// t.Run("Multiple jobs", func(t *testing.T) {
// 	// 	wg.Add(1)
// 	// 	jobCompletions := make(chan string, 3)

// 	// 	jobs := []struct {
// 	// 		command string
// 	// 		delay   string
// 	// 	}{
// 	// 		{"echo 'Job 1'", "in 1 second"},
// 	// 		{"echo 'Job 2'", "in 2 seconds"},
// 	// 		{"echo 'Job 3'", "in 3 seconds"},
// 	// 	}

// 	// 	for _, job := range jobs {
// 	// 		_, err := manager.AddJob(job.command, job.delay, false)
// 	// 		if err != nil {
// 	// 			t.Fatalf("Failed to add job: %v", err)
// 	// 		}
// 	// 	}

// 	// 	go func() {
// 	// 		defer wg.Done()
// 	// 		completedJobs := make(map[string]bool)
// 	// 		for {
// 	// 			select {
// 	// 			case job := <-jobCompletions:
// 	// 				completedJobs[job] = true
// 	// 				if len(completedJobs) == 3 {
// 	// 					log.Debug("All multiple jobs completed")
// 	// 					return
// 	// 				}
// 	// 			case <-time.After(5 * time.Second):
// 	// 				t.Error("Not all multiple jobs completed in time")
// 	// 				return
// 	// 			}
// 	// 		}
// 	// 	}()

// 	// 	for i, job := range jobs {
// 	// 		time.AfterFunc(time.Duration(i+1)*time.Second, func() { jobCompletions <- job.command })
// 	// 	}
// 	// })

// 	// Wait for all tests to complete

// 	log.Info("All tests completed")
// }
