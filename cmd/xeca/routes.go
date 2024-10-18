package xeca

import (
	"net/http"

	"github.com/bethel-nz/goxeca/pkg/goxeca"
)

func Routes(manager *goxeca.Manager) http.Handler {
	mux := http.NewServeMux()
	handler := &Handler{manager}

	// handles the home route
	mux.HandleFunc("GET /{$}", handler.home)

	// returns the status and version of the backend
	mux.HandleFunc("GET /api/status", handler.status)

	// adds a new job to the manager
	mux.HandleFunc("POST /api/add-job", handler.addJob)

	// lists all jobs
	mux.HandleFunc("GET /api/jobs", handler.listJobs)

	// temporary: but it serves as a place holder for webhooks
	mux.HandleFunc("POST /api/webhook", handler.webhook)

	// returns the output of a particular job
	mux.HandleFunc("GET /api/job-output/{id}", handler.jobOutput)

	// pauses a job
	mux.HandleFunc("POST /api/jobs/{id}/pause", handler.pauseJob)

	// resumes a job
	mux.HandleFunc("POST /api/jobs/{id}/resume", handler.resumeJob)

	// stops a job
	mux.HandleFunc("POST /api/jobs/{id}/stop", handler.stopJob)

	// retries a job
	mux.HandleFunc("POST /api/jobs/{id}/retry", handler.retryJob)

	return mux
}
