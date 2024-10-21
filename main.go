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

	config := goxeca.ManagerConfig{
		MaxConcurrent: 20,
		RedisAddr:     "localhost:6379", // <- ideally, you should store this in a .env file and read it
		RedisPassword: "",               // <- same for this
		RedisDB:       1,                // <- and this too
		JobQueueSize:  100,
	}
	manager := goxeca.NewManager(config)
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
