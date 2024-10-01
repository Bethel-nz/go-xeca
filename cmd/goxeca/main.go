package main

import (
	"os"

	"github.com/bethel-nz/goxeca/pkg/goxeca"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"
)

var (
	selectedTime string
	isRecurring  bool
)

func main() {
	manager := goxeca.NewManager()

	var command string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Enter a command to schedule").
				Value(&command),

			huh.NewSelect[string]().
				Title("Pick a time to schedule the job").
				Options(
					huh.NewOption("in 30 seconds", "in 30 seconds"),
					huh.NewOption("in 1 minute", "in 1 minute"),
					huh.NewOption("in 5 minutes", "in 5 minutes"),
					huh.NewOption("in 10 minutes", "in 10 minutes"),
					huh.NewOption("in 15 minutes", "in 15 minutes"),
					huh.NewOption("in 30 minutes", "in 30 minutes"),
				).
				Value(&selectedTime),

			huh.NewConfirm().Title("Is it a recurring job?").Affirmative("Yes").Negative("No").
				Value(&isRecurring),
		),
	)

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}

	job, err := manager.AddJob(command, selectedTime, isRecurring)
	if err != nil {
		log.Error("Error scheduling job:", err)
		os.Exit(1)
	}

	log.Info("Job scheduled with ID: ", job.ID)

	manager.Start()

	select {}
}
