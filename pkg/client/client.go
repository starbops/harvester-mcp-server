package client

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Config represents the configuration for the Kubernetes client.
type Config struct {
	// KubeConfigPath is the path to the kubeconfig file.
	// If empty, it defaults to the KUBECONFIG environment variable,
	// then to ~/.kube/config.
	KubeConfigPath string
}

// Client represents a Kubernetes client for interacting with Harvester clusters.
type Client struct {
	Clientset *kubernetes.Clientset
	Config    *rest.Config
}

// NewClient creates a new Kubernetes client.
func NewClient(cfg *Config) (*Client, error) {
	config, err := getKubeConfig(cfg.KubeConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get Kubernetes config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes clientset: %w", err)
	}

	return &Client{
		Clientset: clientset,
		Config:    config,
	}, nil
}

// getKubeConfig returns a Kubernetes configuration.
func getKubeConfig(kubeConfigPath string) (*rest.Config, error) {
	// Try to use in-cluster config if running in a Kubernetes cluster
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, nil
	}

	// If kubeConfigPath is specified, use it
	if kubeConfigPath != "" {
		config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to build config from specified kubeconfig file %s: %w", kubeConfigPath, err)
		}
		return config, nil
	}

	// Check KUBECONFIG environment variable
	envKubeconfig := os.Getenv("KUBECONFIG")
	if envKubeconfig != "" {
		config, err := clientcmd.BuildConfigFromFlags("", envKubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to build config from KUBECONFIG environment variable %s: %w", envKubeconfig, err)
		}
		return config, nil
	}

	// Fall back to default kubeconfig location
	kubeconfig := filepath.Join(homeDir(), ".kube", "config")
	config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build config from default kubeconfig file: %w", err)
	}

	return config, nil
}

// homeDir returns the user's home directory.
func homeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}
