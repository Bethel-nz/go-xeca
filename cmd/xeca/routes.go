package xeca

import (
	"net/http"

	"github.com/bethel-nz/goxeca/pkg/goxeca"
)

func Routes(manager *goxeca.Manager) http.Handler {
	mux := http.NewServeMux()
	handler := &Handler{manager}

	mux.HandleFunc("/{$}", handler.home)

	mux.HandleFunc("/api/status", handler.status)

	mux.HandleFunc("/api/add-job", handler.addJob)

	mux.HandleFunc("/api/jobs", handler.listJobs)

	mux.HandleFunc("/api/webhook", handler.webhook)

	mux.HandleFunc("/api/job-output/{id}", handler.jobOutput)

	return mux
}
