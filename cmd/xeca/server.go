package xeca

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bethel-nz/goxeca/pkg/goxeca"
	"github.com/charmbracelet/log"
)

func RunWeb(manager *goxeca.Manager) {

	// Set up API routes
	setupAPIRoutes(manager)

	// Start Go server
	log.Info("Starting web server on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Failed to start web server:", err)
	}
}

func setupAPIRoutes(manager *goxeca.Manager) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to Goxeca Web Interface!")
	})

	http.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		status := map[string]string{
			"status":  "running",
			"version": "1.0.0",
		}
		json.NewEncoder(w).Encode(status)
	})

	http.HandleFunc("/api/add-job", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var jobRequest struct {
			Command      string        `json:"command"`
			Schedule     string        `json:"schedule"`
			IsRecurring  bool          `json:"isRecurring"`
			Priority     int           `json:"priority"`
			Dependencies []string      `json:"dependencies"`
			MaxRetries   int           `json:"maxRetries"`
			RetryDelay   time.Duration `json:"retryDelay"`
			Webhook      string        `json:"webhook"`
			Timeout      time.Duration `json:"timeout"`
			ChainedJobs  []string      `json:"chainedJobs"`
		}
		if err := json.NewDecoder(r.Body).Decode(&jobRequest); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		job, err := manager.AddJob(
			jobRequest.Command,
			jobRequest.Schedule,
			jobRequest.IsRecurring,
			jobRequest.Priority,
			jobRequest.Dependencies,
			jobRequest.MaxRetries,
			jobRequest.RetryDelay,
			jobRequest.Webhook,
			jobRequest.Timeout,
			jobRequest.ChainedJobs,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"id": job.ID})
	})

	http.HandleFunc("/api/jobs", func(w http.ResponseWriter, r *http.Request) {
		jobs := manager.ListJobs()
		apiJobs := make([]APIJob, len(jobs))
		for i, job := range jobs {
			apiJobs[i] = APIJob{
				ID:          job.ID,
				Command:     job.Command,
				Schedule:    job.Schedule,
				Status:      job.Status,
				NextRunTime: job.NextRunTime,
				Output:      job.Output,
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(apiJobs)
	})

	http.HandleFunc("/api/webhook", func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			JobID          string `json:"jobID"`
			Status         string `json:"status"`
			ExecutionCount int    `json:"executionCount"`
		}
		json.NewDecoder(r.Body).Decode(&payload)
		fmt.Println("Webhook received", payload)
	})
}

type APIJob struct {
	ID          string    `json:"id"`
	Command     string    `json:"command"`
	Schedule    string    `json:"schedule"`
	Status      string    `json:"status"`
	NextRunTime time.Time `json:"nextRunTime"`
	Output      string    `json:"output"`
}
