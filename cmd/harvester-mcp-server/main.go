package main

import (
	"log"
	"os"

	"github.com/starbops/harvester-mcp-server/pkg/mcp"
)

func main() {
	log.Println("Starting Harvester MCP Server...")

	// Create a new Harvester MCP server
	server, err := mcp.NewServer()
	if err != nil {
		log.Fatalf("Failed to create MCP server: %v", err)
		os.Exit(1)
	}

	// Start the server
	if err := server.ServeStdio(); err != nil {
		log.Fatalf("Failed to start MCP server: %v", err)
		os.Exit(1)
	}
}
