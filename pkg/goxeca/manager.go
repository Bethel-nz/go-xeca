package goxeca

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

type Manager struct {
	jobs          map[string]*Job
	scheduler     *Scheduler
	executor      *Executor
	mu            sync.RWMutex
	done          chan struct{}
	maxConcurrent int
	runningJobs   int
	runningJobsMu sync.Mutex
	templates     map[string]JobTemplate
	cron          *cron.Cron
}

func NewManager() *Manager {
	return &Manager{
		jobs:          make(map[string]*Job),
		scheduler:     NewScheduler(),
		executor:      NewExecutor(),
		done:          make(chan struct{}),
		maxConcurrent: 10,
		templates:     make(map[string]JobTemplate),
		cron:          cron.New(),
	}
}

func (m *Manager) AddJob(command string, schedule string, isRecurring bool, priority int, dependencies []string, maxRetries int, retryDelay time.Duration, webhook string, timeout time.Duration, chainedJobs []string) (*Job, error) {
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
		Priority:       priority,
		Dependencies:   dependencies,
		MaxRetries:     maxRetries,
		RetryDelay:     retryDelay,
		Webhook:        webhook,
		Timeout:        timeout,
		ChainedJobs:    chainedJobs,
	}

	m.mu.Lock()
	m.jobs[job.ID] = job
	m.mu.Unlock()

	if isRecurring {
		if strings.HasPrefix(schedule, "in ") {
			// For human-readable formats, we'll use a goroutine to schedule the job
			go func() {
				m.scheduleRecurringJob(job)
			}()
		} else {
			// For cron expressions, we'll use the cron library
			entryID, err := m.cron.AddFunc(schedule, func() { m.executeJob(job) })
			if err != nil {
				return nil, fmt.Errorf("failed to add cron job: %v", err)
			}
			job.cronEntryID = entryID
		}
	}

	log.Info("Job scheduled", "id", job.ID, "Next Run Time", job.NextRunTime.Format(time.RFC3339))
	return job, nil
}

func (m *Manager) scheduleRecurringJob(job *Job) {
	for {
		time.Sleep(time.Until(job.NextRunTime))
		m.executeJob(job)
		job.NextRunTime = m.scheduler.GetNextRunTime(job.Schedule)
	}
}

func (m *Manager) Start() {
	for {
		m.cron.Start()
		go m.runJobs()
	}
}

func (m *Manager) Stop() {
	m.cron.Stop()
	close(m.done)
}

func (m *Manager) StopJob(jobID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	job, exists := m.jobs[jobID]
	if !exists {
		return fmt.Errorf("job with ID %s not found", jobID)
	}

	switch job.Status {
	case JobStatusRunning:
		job.Status = JobStatusCancelled
	case JobStatusPending, JobStatusPaused:
		delete(m.jobs, jobID)
	default:
		return fmt.Errorf("job with ID %s is already stopped or completed", jobID)
	}

	if job.Recurring {
		m.cron.Remove(cron.EntryID(job.cronEntryID))
	}

	return nil
}

func (m *Manager) runJobs() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-m.done:
			return
		case <-ticker.C:
			m.checkAndExecuteJobs()
		}
	}
}

func (m *Manager) checkAndExecuteJobs() {
	now := time.Now()
	m.mu.RLock()
	var eligibleJobs []*Job
	for _, job := range m.jobs {
		if job.NextRunTime.Before(now) && job.Status == JobStatusPending && !job.Paused && m.areDependenciesMet(job) {
			eligibleJobs = append(eligibleJobs, job)
		}
	}
	m.mu.RUnlock()

	sort.Slice(eligibleJobs, func(i, j int) bool {
		return eligibleJobs[i].Priority > eligibleJobs[j].Priority
	})

	for _, job := range eligibleJobs {
		m.runningJobsMu.Lock()
		if m.runningJobs >= m.maxConcurrent {
			m.runningJobsMu.Unlock()
			break
		}
		m.runningJobs++
		m.runningJobsMu.Unlock()

		go func(j *Job) {
			m.executeJob(j)
			m.runningJobsMu.Lock()
			m.runningJobs--
			m.runningJobsMu.Unlock()
		}(job)
	}
}

func (m *Manager) executeJob(job *Job) {
	startTime := time.Now()
	job.Status = JobStatusRunning

	ctx, cancel := context.WithTimeout(context.Background(), job.Timeout)
	defer cancel()

	// Add a check for cancelled status
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				m.mu.RLock()
				if job.Status == JobStatusCancelled {
					cancel()
					m.mu.RUnlock()
					return
				}
				m.mu.RUnlock()
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	cmd := exec.CommandContext(ctx, "sh", "-c", job.Command)
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/C", job.Command)
	}

	output, err := m.executor.Execute(cmd)

	endTime := time.Now()
	duration := endTime.Sub(startTime)

	jobHistory := JobHistory{
		StartTime: startTime,
		EndTime:   endTime,
		Duration:  duration,
		Output:    output,
	}

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			job.Status = JobStatusFailed
			jobHistory.Status = JobStatusFailed
			log.Error("Job timed out", "id", job.ID, "duration", duration)
		} else {
			job.Status = JobStatusFailed
			jobHistory.Status = JobStatusFailed
			log.Error("Job failed", "id", job.ID, "error", err, "duration", duration)
		}
	} else {
		job.Status = JobStatusCompleted
		jobHistory.Status = JobStatusCompleted
		log.Info("Job completed", "id", job.ID, "duration", duration)
	}

	job.History = append(job.History, jobHistory)
	job.ExecutionCount++

	if job.Webhook != "" {
		go m.sendWebhook(job)
	}

	if job.Status == JobStatusCompleted && len(job.ChainedJobs) > 0 {
		for _, chainedJobID := range job.ChainedJobs {
			if chainedJob, exists := m.jobs[chainedJobID]; exists {
				chainedJob.NextRunTime = time.Now()
			}
		}
	}

	m.mu.Lock()
	if !job.Recurring || (job.Status == JobStatusFailed && job.ExecutionCount > job.MaxRetries) {
		delete(m.jobs, job.ID)
	} else if job.Recurring {
		job.Status = JobStatusPending
		job.NextRunTime = m.scheduler.GetNextRunTime(job.Schedule)
	}
	m.mu.Unlock()
}

func (m *Manager) sendWebhook(job *Job) {
	payload := fmt.Sprintf(`{"jobID": "%s", "status": "%s", "executionCount": %d}`, job.ID, job.Status, job.ExecutionCount)
	_, err := http.Post(job.Webhook, "application/json", strings.NewReader(payload))
	if err != nil {
		log.Error("Failed to send webhook", "jobID", job.ID, "error", err)
	}
}

func (m *Manager) CreateJobTemplate(name string, template JobTemplate) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.templates[name]; exists {
		return fmt.Errorf("template with name %s already exists", name)
	}
	m.templates[name] = template
	return nil
}

func (m *Manager) AddJobFromTemplate(templateName string, customizations map[string]interface{}) (*Job, error) {
	m.mu.RLock()
	template, exists := m.templates[templateName]
	m.mu.RUnlock()
	if !exists {
		return nil, fmt.Errorf("template with name %s not found", templateName)
	}

	// Apply customizations
	for key, value := range customizations {
		switch key {
		case "command":
			template.Command = value.(string)
		case "schedule":
			template.Schedule = value.(string)
			// Add more customization options as needed
		}
	}

	return m.AddJob(
		template.Command,
		template.Schedule,
		template.Recurring,
		template.Priority,
		template.Dependencies,
		template.MaxRetries,
		template.RetryDelay,
		template.Webhook,
		template.Timeout,
		template.ChainedJobs,
	)
}

func (m *Manager) GetJob(id string) (*Job, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	job, exists := m.jobs[id]
	if !exists {
		return nil, fmt.Errorf("job with ID %s not found", id)
	}
	return job, nil
}

func (m *Manager) PauseJob(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	job, exists := m.jobs[id]
	if !exists {
		return fmt.Errorf("job with ID %s not found", id)
	}
	job.Status = JobStatusPaused
	return nil
}

func (m *Manager) ResumeJob(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	job, exists := m.jobs[id]
	if !exists {
		return fmt.Errorf("job with ID %s not found", id)
	}
	job.Status = JobStatusPending
	return nil
}

func (m *Manager) SetMaxConcurrent(max int) {
	m.runningJobsMu.Lock()
	defer m.runningJobsMu.Unlock()
	m.maxConcurrent = max
}

func (m *Manager) areDependenciesMet(job *Job) bool {
	for _, depID := range job.Dependencies {
		m.mu.RLock()
		dep, exists := m.jobs[depID]
		m.mu.RUnlock()
		if !exists {
			// If the dependency job doesn't exist, we assume it has completed
			continue
		}
		if dep.Status != JobStatusCompleted {
			return false
		}
	}
	return true
}

func (m *Manager) ListJobs() []*Job {
	m.mu.RLock()
	defer m.mu.RUnlock()

	jobs := make([]*Job, 0, len(m.jobs))
	for _, job := range m.jobs {
		jobs = append(jobs, job)
	}
	return jobs
}
