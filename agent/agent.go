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

	var status string
	var errStr *string

	if task.Command == "run_lynis" {
		fmt.Printf("Executing task: %s...\n", task.Command)

		payload, runErr := RunLynis()

		if runErr != nil {
			log.Printf("Error executing Lynis: %v", runErr)
			status = "failed"
			eMsg := runErr.Error()
			errStr = &eMsg
		} else {
			// Send to backend
			if sendErr := a.Client.SendScan(a.Config.Token, payload); sendErr != nil {
				log.Printf("Error sending scan payload: %v", sendErr)
				status = "failed"
				eMsg := sendErr.Error()
				errStr = &eMsg
			} else {
				fmt.Println("Scan payload sent successfully.")
				status = "completed"
			}
		}
	} else {
		log.Printf("Unknown task command: %s", task.Command)
		status = "failed"
		eMsg := "unknown command: " + task.Command
		errStr = &eMsg
	}

	// Mark task completed/failed
	if err := a.Client.CompleteTask(a.Config.Token, task.ID, status, errStr); err != nil {
		log.Printf("Error completing task: %v", err)
	} else {
		fmt.Printf("Task %s successfully marked as %s.\n", task.ID, status)
	}
}
