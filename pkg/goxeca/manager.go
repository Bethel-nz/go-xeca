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
	"github.com/shirou/gopsutil/cpu"
)

type ManagerConfig struct {
	MaxConcurrent int
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	JobQueueSize  int
}

type Manager struct {
	scheduler      *Scheduler
	executor       *Executor
	mu             sync.RWMutex
	done           chan struct{}
	maxConcurrent  int
	runningJobs    int
	runningJobsMu  sync.Mutex
	cron           *cron.Cron
	jobQueue       chan *Job
	db             *XecaDB
	config         ManagerConfig
	autoscalerDone chan struct{}
	workerPool     chan struct{}
}

func NewManager(config ManagerConfig) *Manager {
	// Set default values if empty
	if config.MaxConcurrent == 0 {
		config.MaxConcurrent = 10
	}
	if config.JobQueueSize == 0 {
		config.JobQueueSize = 100
	}

	m := &Manager{
		scheduler:      NewScheduler(),
		executor:       NewExecutor(),
		done:           make(chan struct{}),
		maxConcurrent:  config.MaxConcurrent,
		cron:           cron.New(),
		jobQueue:       make(chan *Job, config.JobQueueSize),
		db:             NewXecaDB(config.RedisAddr, config.RedisPassword, config.RedisDB),
		config:         config,
		autoscalerDone: make(chan struct{}),
		workerPool:     make(chan struct{}, config.MaxConcurrent),
	}
	go m.jobWorker()
	return m
}

func (m *Manager) AddJob(command string, schedule string, priority int, dependencies []string, maxRetries int, retryDelay time.Duration, webhook string, timeout time.Duration) (*Job, error) {
	nextRunTime, duration, err := m.scheduler.ParseSchedule(schedule)
	if err != nil {
		return nil, err
	}

	isRecurring := strings.HasPrefix(schedule, "every ") || m.scheduler.IsCronSchedule(schedule)

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
		ActiveRuns:     0,
		TotalRuns:      0,
	}

	if err := m.db.StoreJob(job); err != nil {
		return nil, fmt.Errorf("failed to store job in Redis: %w", err)
	}

	if isRecurring {
		if strings.HasPrefix(schedule, "every ") {
			go m.scheduleRecurringJob(job)
		} else {
			entryID, err := m.cron.AddFunc(schedule, func() { m.queueJob(job) })
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
		select {
		case <-m.done:
			return
		case <-time.After(time.Until(job.NextRunTime)):
			m.queueJob(job)
			job.NextRunTime = m.scheduler.GetNextRunTime(job.Schedule)
		}
	}
}

// Start starts the manager and its components
func (m *Manager) Start() {
	m.cron.Start()
	go m.runJobs()
	go m.runAutoscaler()
}

// Stop stops the manager and its components
func (m *Manager) Stop() {
	m.cron.Stop()
	close(m.done)
	close(m.autoscalerDone)
}

// runJobs runs the job scheduler
func (m *Manager) runJobs() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.done:
			return
		case <-ticker.C:
			m.checkAndQueueJobs()
		}
	}
}

// checkAndQueueJobs checks for pending jobs and queues them
func (m *Manager) checkAndQueueJobs() {
	pendingJobs, err := m.db.GetPendingJobs()
	if err != nil {
		log.Error("Failed to fetch pending jobs", "error", err)
		return
	}

	sort.Slice(pendingJobs, func(i, j int) bool {
		return pendingJobs[i].Priority > pendingJobs[j].Priority
	})

	for _, job := range pendingJobs {
		if job.Recurring {
			// Check if the job is already running
			currentStatus, err := m.db.GetJobStatus(job.ID)
			if err != nil {
				log.Error("Failed to get job status", "id", job.ID, "error", err)
				continue
			}
			if currentStatus == JobStatusRunning {
				log.Info("Skipping recurring job that is already running", "id", job.ID)
				continue
			}
		}

		if m.areDependenciesMet(job) {
			m.queueJob(job)
		}
	}
}

// queueJob queues a job for execution
func (m *Manager) queueJob(job *Job) {
	select {
	case m.jobQueue <- job:
		// Job successfully queued
	default:
		log.Warn("Job queue is full, skipping job", "id", job.ID)
	}
}

// jobWorker is a worker that executes jobs
func (m *Manager) jobWorker() {
	for {
		select {
		case <-m.done:
			return
		case job := <-m.jobQueue:
			m.runningJobsMu.Lock()
			if m.runningJobs < m.maxConcurrent {
				m.runningJobs++
				m.runningJobsMu.Unlock()
				go func() {
					m.executeJob(job)
					m.runningJobsMu.Lock()
					m.runningJobs--
					m.runningJobsMu.Unlock()
				}()
			} else {
				m.runningJobsMu.Unlock()
				m.queueJob(job) // Re-queue the job if we've reached max concurrency
			}
		}
	}
}

// executeJob executes a job
func (m *Manager) executeJob(job *Job) {
	m.workerPool <- struct{}{}
	go func() {
		defer func() { <-m.workerPool }()
		if !m.prepareJobExecution(job) {
			return
		}

		startTime := time.Now()
		output, err := m.runJobCommand(job)
		endTime := time.Now()

		m.handleJobCompletion(job, output, err, startTime, endTime)
	}()
}

// prepareJobExecution prepares a job for execution
func (m *Manager) prepareJobExecution(job *Job) bool {
	currentJob, err := m.db.GetJob(job.ID)
	if err != nil {
		log.Error("Failed to get current job state", "id", job.ID, "error", err)
		return false
	}
	if currentJob.Status == JobStatusCancelled || currentJob.Status == JobStatusPaused {
		log.Info("Job execution skipped due to status", "id", job.ID, "status", currentJob.Status)
		return false
	}

	if err := m.db.UpdateJobStatus(job.ID, JobStatusRunning); err != nil {
		log.Error("Failed to update job status", "id", job.ID, "error", err)
		return false
	}

	if job.Recurring {
		if err := m.db.IncrementActiveRuns(job.ID); err != nil {
			log.Error("Failed to increment active runs", "error", err)
		}
	}

	return true
}

// runJobCommand runs a job command
func (m *Manager) runJobCommand(job *Job) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), job.Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", job.Command)
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/C", job.Command)
	}

	return m.executor.Execute(cmd)
}

// handleJobCompletion handles the completion of a job
func (m *Manager) handleJobCompletion(job *Job, output string, err error, startTime, endTime time.Time) {
	duration := endTime.Sub(startTime)

	jobHistory := JobHistory{
		StartTime: startTime,
		EndTime:   endTime,
		Duration:  duration,
		Output:    output,
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if err != nil {
		m.handleJobError(job, err, duration, &jobHistory)
	} else {
		m.handleJobSuccess(job, &jobHistory)
	}

	m.updateJobState(job, &jobHistory)
}

// handleJobError handles the error of a job
func (m *Manager) handleJobError(job *Job, err error, duration time.Duration, jobHistory *JobHistory) {
	if err == context.DeadlineExceeded {
		job.Status = JobStatusFailed
		jobHistory.Status = JobStatusFailed
		errorMsg := fmt.Sprintf("Job timed out after %v", duration)
		jobHistory.Output += "\n" + errorMsg
		log.Error(errorMsg, "id", job.ID)
	} else {
		job.Status = JobStatusFailed
		jobHistory.Status = JobStatusFailed
		log.Error("Job failed", "id", job.ID, "error", err)
	}

	job.RetryCount++
	if job.RetryCount <= job.MaxRetries {
		job.Status = JobStatusPending
		job.NextRunTime = time.Now().Add(job.RetryDelay)
	}
}

// handleJobSuccess handles the success of a job
func (m *Manager) handleJobSuccess(job *Job, jobHistory *JobHistory) {
	job.Status = JobStatusCompleted
	jobHistory.Status = JobStatusCompleted
	log.Info("Job completed successfully", "id", job.ID)
}

// updateJobState updates the state of a job
func (m *Manager) updateJobState(job *Job, jobHistory *JobHistory) {
	job.ExecutionCount++
	job.History = append(job.History, *jobHistory)

	if job.Webhook != "" {
		go m.sendWebhook(job)
	}

	if job.Recurring {
		if err := m.db.DecrementActiveRuns(job.ID); err != nil {
			log.Error("Failed to decrement active runs", "error", err)
		}
		if err := m.db.IncrementTotalRuns(job.ID); err != nil {
			log.Error("Failed to increment total runs", "error", err)
		}
	}

	if job.Recurring {
		job.Status = JobStatusPending
		job.NextRunTime = m.scheduler.GetNextRunTime(job.Schedule)
		if err := m.db.StoreJob(job); err != nil {
			log.Error("Failed to update recurring job", "id", job.ID, "error", err)
		}
	} else if job.Status == JobStatusCompleted || (job.Status == JobStatusFailed && job.ExecutionCount > job.MaxRetries) {
		if err := m.db.StoreJobWithTTL(job, 24*time.Hour); err != nil {
			log.Error("Failed to update job with TTL", "id", job.ID, "error", err)
		}
	}

	if err := m.db.UpdateJobStatus(job.ID, job.Status); err != nil {
		log.Error("Failed to update job status", "id", job.ID, "error", err)
	}
}

// StopJob stops a job
func (m *Manager) StopJob(jobID string) error {
	job, err := m.db.GetJob(jobID)
	if err != nil {
		return fmt.Errorf("job with ID %s not found: %w", jobID, err)
	}

	job.Status = JobStatusCancelled
	if job.Recurring {
		m.cron.Remove(cron.EntryID(job.cronEntryID))
	}

	log.Info("Job cancelled", "id:", job.ID)
	return m.db.StoreJob(job)
}

// sendWebhook sends a webhook for a job
func (m *Manager) sendWebhook(job *Job) {
	payload := fmt.Sprintf(`{"jobID": "%s", "status": "%s", "executionCount": %d}`, job.ID, job.Status, job.ExecutionCount)
	_, err := http.Post(job.Webhook, "application/json", strings.NewReader(payload))
	if err != nil {
		log.Error("Failed to send webhook", "jobID", job.ID, "error", err)
	}
}

// GetJob gets a job by ID
func (m *Manager) GetJob(id string) (*Job, error) {
	return m.db.GetJob(id)
}

// PauseJob pauses a job
func (m *Manager) PauseJob(id string) error {
	job, err := m.db.GetJob(id)
	if err != nil {
		return fmt.Errorf("job with ID %s not found: %w", id, err)
	}
	job.Status = JobStatusPaused
	job.Paused = true
	if job.Recurring {
		if job.cronEntryID != 0 {
			m.cron.Remove(cron.EntryID(job.cronEntryID))
		}
		// For "every X" schedules, we don't need to do anything extra
	}
	return m.db.StoreJob(job)
}

// ResumeJob resumes a job
func (m *Manager) ResumeJob(id string) error {
	job, err := m.db.GetJob(id)
	if err != nil {
		return fmt.Errorf("job with ID %s not found: %w", id, err)
	}
	job.Status = JobStatusPending
	job.Paused = false
	if job.Recurring {
		if strings.HasPrefix(job.Schedule, "every ") {
			go m.scheduleRecurringJob(job)
		} else {
			entryID, err := m.cron.AddFunc(job.Schedule, func() { m.queueJob(job) })
			if err != nil {
				return fmt.Errorf("failed to resume cron job: %v", err)
			}
			job.cronEntryID = entryID
		}
	}
	return m.db.StoreJob(job)
}

// SetMaxConcurrent sets the maximum number of concurrent jobs
func (m *Manager) SetMaxConcurrent(max int) {
	m.runningJobsMu.Lock()
	defer m.runningJobsMu.Unlock()
	m.maxConcurrent = max
}

// areDependenciesMet checks if the dependencies of a job are met
func (m *Manager) areDependenciesMet(job *Job) bool {
	for _, depID := range job.Dependencies {
		dep, err := m.db.GetJob(depID)
		if err != nil {
			log.Error("Failed to fetch dependency job", "id", depID, "error", err)
			return false
		}
		if dep.Status != JobStatusCompleted {
			return false
		}
	}
	return true
}

// ListJobs lists all jobs
func (m *Manager) ListJobs() ([]*Job, error) {
	return m.db.FetchAllJobs()
}

// GetJobOutput gets the output of a job
func (m *Manager) GetJobOutput(jobID string) (string, error) {
	job, err := m.db.GetJob(jobID)
	if err != nil {
		return "", fmt.Errorf("failed to fetch job %s: %w", jobID, err)
	}

	if len(job.History) == 0 {
		return "", fmt.Errorf("no history available for job %s", jobID)
	}

	// Return the output of the most recent execution
	return string(job.History[len(job.History)-1].Output), nil
}

// CompleteRecurringJob completes a recurring job
func (m *Manager) CompleteRecurringJob(jobID string) error {
	job, err := m.db.GetJob(jobID)
	if err != nil {
		return fmt.Errorf("job with ID %s not found: %w", jobID, err)
	}

	if !job.Recurring {
		return fmt.Errorf("job with ID %s is not a recurring job", jobID)
	}

	// Remove the cron entry
	if job.cronEntryID != 0 {
		m.cron.Remove(cron.EntryID(job.cronEntryID))
	}

	// Mark the recurring job as completed
	job.Status = JobStatusCompleted
	job.Recurring = false
	if err := m.db.StoreJob(job); err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	return nil
}

// runAutoscaler runs the autoscaler best for systems with low memory or high memeory usage

func (m *Manager) runAutoscaler() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.autoscalerDone:
			return
		case <-ticker.C:
			m.adjustConcurrency()
		}
	}
}

func (m *Manager) adjustConcurrency() {
	cpuUsage, _ := cpu.Percent(time.Second, false)
	avgCPUUsage := cpuUsage[0]

	// Use exponential backoff for adjustments
	if avgCPUUsage < 50 {
		m.SetMaxConcurrent(m.maxConcurrent * 2)
	} else if avgCPUUsage > 80 {
		m.SetMaxConcurrent(m.maxConcurrent / 2)
	}
}

func (m *Manager) pauseLeastPriorityJob() {
	jobs, err := m.ListJobs()
	if err != nil {
		log.Error("Failed to list jobs", "error", err)
		return
	}

	var leastPriorityJob *Job
	for _, job := range jobs {
		if job.Status == JobStatusRunning && (leastPriorityJob == nil || job.Priority < leastPriorityJob.Priority) {
			leastPriorityJob = job
		}
	}

	if leastPriorityJob != nil {
		err := m.PauseJob(leastPriorityJob.ID)
		if err != nil {
			log.Error("Failed to pause job", "id", leastPriorityJob.ID, "error", err)
		}
	}
}

func (m *Manager) resumePausedJobs() {
	jobs, err := m.ListJobs()
	if err != nil {
		log.Error("Failed to list jobs", "error", err)
		return
	}

	for _, job := range jobs {
		if job.Status == JobStatusPaused {
			err := m.ResumeJob(job.ID)
			if err != nil {
				log.Error("Failed to resume job", "id", job.ID, "error", err)
			}
		}
	}
}

// RetryJob retries a job
func (m *Manager) RetryJob(jobID string) error {
	job, err := m.db.GetJob(jobID)
	if err != nil {
		return fmt.Errorf("job with ID %s not found: %w", jobID, err)
	}

	if job.Status != JobStatusFailed && job.Status != JobStatusCancelled {
		return fmt.Errorf("job with ID %s is not in a failed or cancelled state", jobID)
	}

	job.Status = JobStatusPending
	job.RetryCount = 0
	job.NextRunTime = time.Now()

	if err := m.db.StoreJob(job); err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	m.queueJob(job)
	return nil
}
