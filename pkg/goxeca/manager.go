package goxeca

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
)

type Manager struct {
	jobs      map[string]*Job
	scheduler *Scheduler
	executor  *Executor
	queue     *Queue
	mu        sync.Mutex
}

func NewManager() *Manager {
	return &Manager{
		jobs:      make(map[string]*Job),
		scheduler: NewScheduler(),
		executor:  NewExecutor(),
		queue:     NewQueue(),
	}
}

func (m *Manager) AddJob(command string, schedule string, isRecurring bool) (*Job, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	nextRunTime, duration, err := m.scheduler.ParseSchedule(schedule)
	if err != nil {
		return nil, err
	}

	job := &Job{
		ID:             uuid.New().String(),
		Command:        command,
		Schedule:       schedule,
		NextRunTime:    nextRunTime,
		Status:         JobStatusPending,
		Recurring:      isRecurring,
		ExecutionCount: 0,
		Duration:       duration,
	}

	m.jobs[job.ID] = job
	logMessage := fmt.Sprintf("Job %s scheduled to run at %s", job.ID, job.NextRunTime.Format(time.RFC3339))
	log.Info(logMessage)
	return job, nil
}

func (m *Manager) Start() {
	go m.runJobs()
}

func (m *Manager) runJobs() {
	ticker := time.NewTicker(10 * time.Millisecond) // Reduced from 100ms to 10ms
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		m.mu.Lock()
		for id, job := range m.jobs {
			if job.NextRunTime.Before(now) && job.Status == JobStatusPending {
				m.queue.Enqueue(job)
				job.Status = JobStatusQueued
				m.mu.Lock()
				if !job.Recurring {
					delete(m.jobs, id)
				}
				m.mu.Unlock()
			}
		}
		m.mu.Unlock()

		for m.queue.Length() > 0 {
			job := m.queue.Dequeue()
			go m.executeJob(job)
		}
	}
}

func (m *Manager) executeJob(job *Job) {
	job.Status = JobStatusRunning
	startTime := time.Now()

	output, err := m.executor.Execute(job)
	job.Output = output

	duration := time.Since(startTime)
	job.ExecutionCount++

	if err != nil {
		job.Status = JobStatusFailed
		logMessage := fmt.Sprintf("Job %s (execution %d) failed after %v: %v", job.ID, job.ExecutionCount, duration, err)
		log.Error(logMessage)
		logMessage = fmt.Sprintf("Job output:\n%s\n", output)
		log.Error(logMessage)
		if !job.Recurring {
			logMessage = "Terminating program due to non-recurring job failure."
			log.Error(logMessage)
			os.Exit(1)
		}
	} else {
		job.Status = JobStatusCompleted
		logMessage := fmt.Sprintf("Job %s (execution %d) completed successfully in %v", job.ID, job.ExecutionCount, duration)
		log.Info(logMessage)
		logMessage = fmt.Sprintf("Job output:\n%s\n", output)
		log.Info(logMessage)
	}

	if job.Recurring {
		job.NextRunTime = time.Now().Add(job.Duration)
		job.Status = JobStatusPending
	} else {
		delete(m.jobs, job.ID)
		logMessage := "Job completed. Exiting program."
		log.Info(logMessage)
		os.Exit(0)
	}
}

func (m *Manager) GetJob(id string) (*Job, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	job, exists := m.jobs[id]
	if !exists {
		return nil, fmt.Errorf("job with ID %s not found", id)
	}
	return job, nil
}
