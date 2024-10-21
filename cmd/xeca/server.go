package xeca

import (
	"net/http"
	"os"

	"github.com/bethel-nz/goxeca/pkg/goxeca"
	"github.com/charmbracelet/log"
)

func RunWeb(manager *goxeca.Manager) {

	// Set up API routes and server
	server := http.Server{
		Addr:    ":8080",
		Handler: Routes(manager),
	}

	// Start Go server
	log.Info("Starting web server on :8080")
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("Failed to start web server:", err.Error())
		os.Exit(1)
	}
}
