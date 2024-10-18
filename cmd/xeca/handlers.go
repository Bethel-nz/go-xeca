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
		jobRequest.Priority,
		jobRequest.Dependencies,
		jobRequest.MaxRetries,
		jobRequest.RetryDelay,
		jobRequest.Webhook,
		jobRequest.Timeout,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"message": "Job added successfully",
		"job": map[string]interface{}{
			"id":           job.ID,
			"command":      job.Command,
			"schedule":     job.Schedule,
			"status":       job.Status,
			"nextRunTime":  job.NextRunTime,
			"recurring":    job.Recurring,
			"priority":     job.Priority,
			"dependencies": job.Dependencies,
			"maxRetries":   job.MaxRetries,
			"retryDelay":   job.RetryDelay,
			"webhook":      job.Webhook,
			"timeout":      job.Timeout,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) listJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := h.manager.ListJobs()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, "Webhook received", payload)
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

func (h *Handler) pauseJob(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("id")
	err := h.manager.PauseJob(jobID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	job, _ := h.manager.GetJob(jobID)
	response := map[string]interface{}{
		"message": "Job paused successfully",
		"job":     job,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) resumeJob(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("id")
	err := h.manager.ResumeJob(jobID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	job, _ := h.manager.GetJob(jobID)
	response := map[string]interface{}{
		"message": "Job resumed successfully",
		"job":     job,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) stopJob(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("id")
	err := h.manager.StopJob(jobID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	job, _ := h.manager.GetJob(jobID)
	response := map[string]interface{}{
		"message": "Job stopped successfully",
		"job":     job,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) retryJob(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("id")
	err := h.manager.RetryJob(jobID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Job retried successfully"))
}
