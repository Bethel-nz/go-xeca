package xeca

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bethel-nz/goxeca/pkg/goxeca"
)

type Handler struct {
	manager *goxeca.Manager
}

func (h *Handler) home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to Goxeca Web Interface!")
}

func (h *Handler) status(w http.ResponseWriter, r *http.Request) {
	status := map[string]string{
		"name":        "Goxeca",
		"description": "Goxeca is a job scheduler and executor that allows you to schedule and execute Background jobs.",
		"status":      "running",
		"version":     "1.0.0",
	}
	json.NewEncoder(w).Encode(status)
}

func (h *Handler) addJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var jobRequest APIJobRequest

	if err := json.NewDecoder(r.Body).Decode(&jobRequest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	job, err := h.manager.AddJob(
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
}

func (h *Handler) listJobs(w http.ResponseWriter, r *http.Request) {
	jobs := h.manager.ListJobs()
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
}

func (h *Handler) webhook(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		JobID          string `json:"jobID"`
		Status         string `json:"status"`
		ExecutionCount int    `json:"executionCount"`
	}
	json.NewDecoder(r.Body).Decode(&payload)
	fmt.Println("Webhook received", payload)
}

func (h *Handler) jobOutput(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("id")
	output, err := h.manager.GetJobOutput(jobID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(output))
}
