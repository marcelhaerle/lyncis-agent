package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/joho/godotenv"
)

type RegisterRequest struct {
	Hostname string `json:"hostname"`
	OSInfo   string `json:"os_info"`
}

type RegisterResponse struct {
	AgentID string `json:"agent_id"`
	Token   string `json:"token"`
}

type Config struct {
	AgentID string `json:"agent_id"`
	Token   string `json:"token"`
}

type Task struct {
	ID      string `json:"id"`
	Command string `json:"command"`
}

type TaskResponse struct {
	Task *Task `json:"task"`
}

type CompleteTaskRequest struct {
	Status string  `json:"status"`
	Error  *string `json:"error"`
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func startPolling(backendURL string, config *Config) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	client := &http.Client{Timeout: 5 * time.Second}
	fmt.Println("Started task polling (Heartbeat) every 10 seconds...")

	for range ticker.C {
		pollTasks(backendURL, config, client)
	}
}

func pollTasks(backendURL string, config *Config, client *http.Client) {
	reqURL := fmt.Sprintf("%s/api/v1/agent/tasks/pending", backendURL)
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.Token))

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error polling tasks: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		// Nothing pending
		return
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Failed to poll tasks, Status: %d", resp.StatusCode)
		return
	}

	var taskResp TaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&taskResp); err != nil {
		log.Printf("Error decoding task response: %v", err)
		return
	}

	if taskResp.Task == nil {
		return
	}

	fmt.Printf("Received task: %s (ID: %s)\n", taskResp.Task.Command, taskResp.Task.ID)

	// Mock execution
	fmt.Printf("Executing mock task: %s...\n", taskResp.Task.Command)
	time.Sleep(1 * time.Second)
	fmt.Println("Mock task completed.")

	// Mark task completed
	completeTask(backendURL, config, client, taskResp.Task.ID)
}

func completeTask(backendURL string, config *Config, client *http.Client, taskID string) {
	completeURL := fmt.Sprintf("%s/api/v1/agent/tasks/%s/complete", backendURL, taskID)
	reqBody := CompleteTaskRequest{
		Status: "completed",
		Error:  nil,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("Failed to marshal complete request: %v", err)
		return
	}

	req, err := http.NewRequest(http.MethodPost, completeURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error creating complete request: %v", err)
		return
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.Token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error completing task: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		log.Printf("Failed to complete task, Status: %d", resp.StatusCode)
	} else {
		fmt.Printf("Task %s successfully marked as completed.\n", taskID)
	}
}

func main() {
	fmt.Println("Starting Lyncis Agent...")

	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found or error loading it. Using existing environment variables.")
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("Failed to read hostname: %v", err)
	}
	fmt.Printf("Hostname: %s\n", hostname)

	osInfo := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)

	backendURL := os.Getenv("LYNCIS_BACKEND_URL")
	if backendURL == "" {
		panic("LYNCIS_BACKEND_URL environment variable is missing")
	}

	// Set config path
	configPath := os.Getenv("LYNCIS_CONFIG_PATH")
	if configPath == "" {
		panic("LYNCIS_CONFIG_PATH environment variable is missing")
	}

	// Check if already registered
	if _, err := os.Stat(configPath); err == nil {
		fmt.Println("Agent already registered. Configuration found at:", configPath)
		config, err := loadConfig(configPath)
		if err != nil {
			log.Fatalf("Failed to load configuration: %v", err)
		}
		// Start polling (Feature 3) here
		startPolling(backendURL, config)
		return
	}

	reqBody := RegisterRequest{
		Hostname: hostname,
		OSInfo:   osInfo,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		log.Fatalf("Failed to marshal request: %v", err)
	}

	registerURL := fmt.Sprintf("%s/api/v1/agent/register", backendURL)
	resp, err := http.Post(registerURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Failed to call backend API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyInfo, _ := io.ReadAll(resp.Body)
		log.Fatalf("Failed to register agent. Status: %d, Response: %s", resp.StatusCode, string(bodyInfo))
	}

	var regResp RegisterResponse
	if err := json.NewDecoder(resp.Body).Decode(&regResp); err != nil {
		log.Fatalf("Failed to decode response: %v", err)
	}

	fmt.Printf("Registered successfully. Agent ID: %s\n", regResp.AgentID)

	// Save the returned token locally.
	config := Config{
		AgentID: regResp.AgentID,
		Token:   regResp.Token,
	}

	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal config: %v", err)
	}

	err = os.WriteFile(configPath, configData, 0600)
	if err != nil {
		log.Fatalf("Failed to write config locally: %v", err)
	}

	fmt.Printf("Configuration saved to %s\n", configPath)

	// Start polling (Feature 3) after fresh registration
	startPolling(backendURL, &config)
}
