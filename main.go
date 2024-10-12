package main

import (
	"flag"
	"fmt"

	"github.com/bethel-nz/goxeca/cmd/xeca"
	"github.com/bethel-nz/goxeca/pkg/goxeca"
	"github.com/charmbracelet/log"
)

func main() {
	fmt.Println("Starting GoXeca...")
	mode := flag.String("mode", "cli", "Mode to run the application (cli or web)")
	flag.Parse()

	manager := goxeca.NewManager()
	manager.Start()
	defer manager.Stop()

	switch *mode {
	case "cli":
		xeca.RunCLI(manager)
	case "web":
		xeca.RunWeb(manager)
	default:
		log.Fatal("Invalid mode. Use 'cli' or 'web'")
	}

}
