package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/starbops/harvester-mcp-server/pkg/mcp"
)

var (
	// Global flags
	kubeConfigPath string

	// Root command
	rootCmd = &cobra.Command{
		Use:   "harvester-mcp-server",
		Short: "Harvester MCP Server - MCP server for Harvester HCI",
		Long: `Harvester MCP Server is a Model Context Protocol (MCP) server for Harvester HCI.
It allows AI assistants like Claude and Cursor to interact with Harvester clusters through the MCP protocol.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServer()
		},
		// Disable the automatic help message when an error occurs
		SilenceUsage: true,
		// Disable automatic error printing since we'll handle it explicitly
		SilenceErrors: true,
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("%v", err)
	}
}

func init() {
	// Add flags
	rootCmd.PersistentFlags().StringVar(&kubeConfigPath, "kubeconfig", "", "Path to the kubeconfig file (default is $KUBECONFIG or $HOME/.kube/config)")
}

func runServer() error {
	// Create server configuration
	cfg := &mcp.Config{
		KubeConfigPath: kubeConfigPath,
	}

	// Create and start the MCP server
	server, err := mcp.NewServer(cfg)
	if err != nil {
		return fmt.Errorf("failed to create MCP server: %w", err)
	}

	// Start the server
	if err := server.ServeStdio(); err != nil {
		return fmt.Errorf("failed to start MCP server: %w", err)
	}

	return nil
}
