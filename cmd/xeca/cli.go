package xeca

import (
	"fmt"
	"os"

	"github.com/bethel-nz/goxeca/pkg/goxeca"
	"github.com/charmbracelet/log"
)

func RunCLI(manager *goxeca.Manager) {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]

	switch command {
	case "add":
		addJob(manager)
	case "list":
		listJobs(manager)
	case "stop":
		stopJob(manager)
	case "pause":
		pauseJob(manager)
	case "resume":
		resumeJob(manager)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Usage: xeca <command> [arguments]")
	fmt.Println("Commands:")
	fmt.Println("  add    - Add a new job")
	fmt.Println("  list   - List all jobs")
	fmt.Println("  stop   - Stop a job")
	fmt.Println("  pause  - Pause a job")
	fmt.Println("  resume - Resume a paused job")
}

func addJob(manager *goxeca.Manager) {
	if len(os.Args) < 4 {
		fmt.Println("Usage: xeca add <command> <schedule>")
		return
	}

	command := os.Args[2]
	schedule := os.Args[3]

	job, err := manager.AddJob(command, schedule, false, 1, nil, 0, 0, "", 0, nil)
	if err != nil {
		log.Error("Error adding job:", err)
		return
	}

	fmt.Printf("Job added with ID: %s\n", job.ID)
}

func listJobs(manager *goxeca.Manager) {
	jobs := manager.ListJobs()
	for _, job := range jobs {
		fmt.Printf("ID: %s, Command: %s, Status: %s\n", job.ID, job.Command, job.Status)
	}
}

func stopJob(manager *goxeca.Manager) {
	if len(os.Args) < 3 {
		fmt.Println("Usage: xeca stop <job_id>")
		return
	}

	jobID := os.Args[2]
	err := manager.StopJob(jobID)
	if err != nil {
		log.Error("Error stopping job:", err)
		return
	}

	fmt.Printf("Job %s stopped\n", jobID)
}

func pauseJob(manager *goxeca.Manager) {
	if len(os.Args) < 3 {
		fmt.Println("Usage: xeca pause <job_id>")
		return
	}

	jobID := os.Args[2]
	err := manager.PauseJob(jobID)
	if err != nil {
		log.Error("Error pausing job:", err)
		return
	}

	fmt.Printf("Job %s paused\n", jobID)
}

func resumeJob(manager *goxeca.Manager) {
	if len(os.Args) < 3 {
		fmt.Println("Usage: xeca resume <job_id>")
		return
	}

	jobID := os.Args[2]
	err := manager.ResumeJob(jobID)
	if err != nil {
		log.Error("Error resuming job:", err)
		return
	}

	fmt.Printf("Job %s resumed\n", jobID)
}
