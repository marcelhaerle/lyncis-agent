package agent

import (
	"fmt"
	"log"
	"time"

	"github.com/marcelhaerle/lyncis-agent/api"
	"github.com/marcelhaerle/lyncis-agent/config"
)

type Agent struct {
	Config *config.Config
	Client *api.Client
}

func NewAgent(cfg *config.Config, client *api.Client) *Agent {
	return &Agent{
		Config: cfg,
		Client: client,
	}
}

func (a *Agent) StartPolling() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	fmt.Println("Started task polling (Heartbeat) every 10 seconds...")

	for range ticker.C {
		a.pollOnce()
	}
}

func (a *Agent) pollOnce() {
	task, err := a.Client.PollPendingTask(a.Config.Token)
	if err != nil {
		log.Printf("Error polling tasks: %v", err)
		return
	}

	if task == nil {
		// Nothing pending
		return
	}

	fmt.Printf("Received task: %s (ID: %s)\n", task.Command, task.ID)

	// Mock execution
	fmt.Printf("Executing mock task: %s...\n", task.Command)
	time.Sleep(1 * time.Second)
	fmt.Println("Mock task completed.")

	// Mark task completed
	if err := a.Client.CompleteTask(a.Config.Token, task.ID); err != nil {
		log.Printf("Error completing task: %v", err)
	} else {
		fmt.Printf("Task %s successfully marked as completed.\n", task.ID)
	}
}
