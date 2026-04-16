package main

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/joho/godotenv"

	"github.com/marcelhaerle/lyncis-agent/agent"
	"github.com/marcelhaerle/lyncis-agent/api"
	"github.com/marcelhaerle/lyncis-agent/config"
)

func main() {
	fmt.Println("Starting Lyncis Agent...")

	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found or error loading it. Using existing environment variables.")
	}

	backendURL := os.Getenv("LYNCIS_BACKEND_URL")
	if backendURL == "" {
		log.Fatal("LYNCIS_BACKEND_URL environment variable is missing")
	}

	configPath := os.Getenv("LYNCIS_CONFIG_PATH")
	if configPath == "" {
		configPath = "/etc/lyncis/config.json"
		fmt.Printf("LYNCIS_CONFIG_PATH is missing, defaulting to %s\n", configPath)
	}

	client := api.NewClient(backendURL)

	// Check if already registered
	if _, err := os.Stat(configPath); err == nil {
		fmt.Println("Agent already registered. Configuration found at:", configPath)
		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			log.Fatalf("Failed to load configuration: %v", err)
		}

		ag := agent.NewAgent(cfg, client)
		ag.StartPolling()
		return
	}

	// Register agent
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("Failed to read hostname: %v", err)
	}
	fmt.Printf("Hostname: %s\n", hostname)

	osInfo := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)

	req := api.RegisterRequest{
		Hostname: hostname,
		OSInfo:   osInfo,
	}

	regResp, err := client.Register(req)
	if err != nil {
		log.Fatalf("Failed to register agent: %v", err)
	}

	fmt.Printf("Registered successfully. Agent ID: %s\n", regResp.AgentID)

	// Save configuration
	cfg := &config.Config{
		AgentID: regResp.AgentID,
		Token:   regResp.Token,
	}

	if err := config.SaveConfig(configPath, cfg); err != nil {
		log.Fatalf("Failed to write config locally: %v", err)
	}

	fmt.Printf("Configuration saved to %s\n", configPath)

	ag := agent.NewAgent(cfg, client)
	ag.StartPolling()
}
