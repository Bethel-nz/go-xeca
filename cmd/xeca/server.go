package xeca

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"time"

	"github.com/bethel-nz/goxeca/pkg/goxeca"
	"github.com/charmbracelet/log"
)


func RunWeb(manager *goxeca.Manager) {
	// Start React frontend
	go startReactFrontend()

	// Set up API routes
	setupAPIRoutes(manager)

	// Start Go server
	log.Info("Starting web server on :8080")
	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatal("Failed to start web server:", err)
		}
	}()

	// Keep the main goroutine alive
	select {}
}

func startReactFrontend() {
	cmd := exec.Command("bun", "run", "dev")
	cmd.Dir = "../../web"
	err := cmd.Start()
	if err != nil {
		log.Error("Failed to start React frontend:", err)
	} else {
		log.Info("React frontend starting...")
	}
}

func setupAPIRoutes(manager *goxeca.Manager) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to Goxeca Web Interface!")
	})

	http.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		status := map[string]string{
			"status": "running",
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
			Command string `json:"command"`
			Schedule string `json:"schedule"`
		}
		if err := json.NewDecoder(r.Body).Decode(&jobRequest); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		job, err := manager.AddJob(jobRequest.Command, jobRequest.Schedule, true, 2, []string{}, 3, time.Second, "", time.Duration(0), []string{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"id": job.ID})
	})
}
