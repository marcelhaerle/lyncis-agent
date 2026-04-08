package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	fmt.Println("Starting Lyncis Agent...")

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("Failed to read hostname: %v", err)
	}
	fmt.Printf("Hostname: %s\n", hostname)
}
