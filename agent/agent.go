package agent

import (
	"log"
	"time"

	"github.com/marcelhaerle/lyncis-agent/api"
	"github.com/marcelhaerle/lyncis-agent/config"
)

type Agent struct {
	Config           *config.Config
	Client           *api.Client
	consecutiveErrs int
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

	log.Println("Started task polling (Heartbeat) every 10 seconds...")

	for range ticker.C {
		a.pollOnce()
	}
}

func (a *Agent) pollOnce() {
	task, err := a.Client.PollPendingTask(a.Config.Token)
	if err != nil {
		a.consecutiveErrs++
		// Log only on the first failure or every 6th failure (approx once per minute)
		if a.consecutiveErrs == 1 || a.consecutiveErrs%6 == 0 {
			log.Printf("Error polling tasks (consecutive failures: %d)", a.consecutiveErrs)
		}
		return
	}

	a.consecutiveErrs = 0

	if task == nil {
		// Nothing pending
		return
	}

	log.Printf("Received task: %s (ID: %s)", task.Command, task.ID)

	var status string
	var errStr *string

	if task.Command == "run_lynis" {
		log.Printf("Executing task: %s...", task.Command)

		payload, runErr := RunLynis()

		if runErr != nil {
			log.Printf("Error executing Lynis")
			status = "failed"
			eMsg := "internal execution error"
			errStr = &eMsg
		} else {
			// Send to backend
			if sendErr := a.Client.SendScan(a.Config.Token, payload); sendErr != nil {
				log.Printf("Error sending scan payload")
				status = "failed"
				eMsg := "internal transmission error"
				errStr = &eMsg
			} else {
				log.Println("Scan payload sent successfully.")
				status = "completed"
			}
		}
	} else {
		log.Printf("Unknown task command: %s", task.Command)
		status = "failed"
		eMsg := "unknown command"
		errStr = &eMsg
	}

	// Mark task completed/failed
	if err := a.Client.CompleteTask(a.Config.Token, task.ID, status, errStr); err != nil {
		log.Printf("Error completing task")
	} else {
		log.Printf("Task %s successfully marked as %s.", task.ID, status)
	}
}
