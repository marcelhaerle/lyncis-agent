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
		// We would load config and start polling (Feature 3) here
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
}
