package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	log "github.com/sirupsen/logrus"
	"github.com/starbops/harvester-mcp-server/pkg/client"
	"github.com/starbops/harvester-mcp-server/pkg/kubernetes"
)

// Config represents the configuration for the Harvester MCP server.
type Config struct {
	// KubeConfigPath is the path to the kubeconfig file.
	KubeConfigPath string
}

// HarvesterMCPServer represents the MCP server for Harvester HCI.
type HarvesterMCPServer struct {
	mcpServer       *server.MCPServer
	k8sClient       *client.Client
	resourceHandler *kubernetes.ResourceHandler
}

// NewServer creates a new Harvester MCP server.
func NewServer(cfg *Config) (*HarvesterMCPServer, error) {
	// Create client configuration
	clientCfg := &client.Config{
		KubeConfigPath: cfg.KubeConfigPath,
	}

	// Create Kubernetes client
	k8sClient, err := client.NewClient(clientCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Create resource handler
	resourceHandler, err := kubernetes.NewResourceHandler(k8sClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource handler: %w", err)
	}

	// Create a new MCP server
	mcpServer := server.NewMCPServer(
		"Harvester MCP Server",
		"1.0.0",
	)

	harvesterServer := &HarvesterMCPServer{
		mcpServer:       mcpServer,
		k8sClient:       k8sClient,
		resourceHandler: resourceHandler,
	}

	// Register tools
	harvesterServer.registerTools()

	return harvesterServer, nil
}

// ServeStdio starts the MCP server using stdio.
func (s *HarvesterMCPServer) ServeStdio() error {
	log.Info("Starting Harvester MCP server...")
	return server.ServeStdio(s.mcpServer)
}

// registerTools registers all the tools with the MCP server.
func (s *HarvesterMCPServer) registerTools() {
	// Register Kubernetes common tools
	s.registerKubernetesPodTools()
	s.registerKubernetesDeploymentTools()
	s.registerKubernetesServiceTools()
	s.registerKubernetesNamespaceTools()
	s.registerKubernetesNodeTools()
	s.registerKubernetesCRDTools()

	// Register Harvester-specific tools
	s.registerHarvesterVirtualMachineTools()
	s.registerHarvesterImageTools()
	s.registerHarvesterVolumeTools()
	s.registerHarvesterNetworkTools()
}

// registerKubernetesPodTools registers Pod-related tools.
func (s *HarvesterMCPServer) registerKubernetesPodTools() {
	// List pods tool
	listPodsTool := mcp.NewTool(
		"list_pods",
		mcp.WithDescription("List pods in the Harvester cluster"),
		mcp.WithString("namespace",
			mcp.Description("The namespace to list pods from (optional, defaults to all namespaces)"),
		),
	)
	s.mcpServer.AddTool(listPodsTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace, _ := req.Params.Arguments["namespace"].(string)

		// Use the unified resource handler
		gvr := kubernetes.ResourceTypeToGVR[kubernetes.ResourceTypePods]
		list, err := s.resourceHandler.ListResources(ctx, gvr, namespace)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list pods: %v", err)), nil
		}

		// Format the list using the resource formatter
		formatted := s.resourceHandler.FormatResourceList(list, gvr)
		return mcp.NewToolResultText(formatted), nil
	})

	// Get pod tool
	getPodTool := mcp.NewTool(
		"get_pod",
		mcp.WithDescription("Get pod details from the Harvester cluster"),
		mcp.WithString("namespace",
			mcp.Required(),
			mcp.Description("The namespace of the pod"),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("The name of the pod"),
		),
	)
	s.mcpServer.AddTool(getPodTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace, ok := req.Params.Arguments["namespace"].(string)
		if !ok || namespace == "" {
			return mcp.NewToolResultError("Namespace is required"), nil
		}

		name, ok := req.Params.Arguments["name"].(string)
		if !ok || name == "" {
			return mcp.NewToolResultError("Pod name is required"), nil
		}

		// Use the unified resource handler
		gvr := kubernetes.ResourceTypeToGVR[kubernetes.ResourceTypePod]
		resource, err := s.resourceHandler.GetResource(ctx, gvr, namespace, name)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get pod %s in namespace %s: %v", name, namespace, err)), nil
		}

		// Format the resource using the resource formatter
		formatted := s.resourceHandler.FormatResource(resource, gvr)
		return mcp.NewToolResultText(formatted), nil
	})

	// Delete pod tool
	deletePodTool := mcp.NewTool(
		"delete_pod",
		mcp.WithDescription("Delete a pod from the Harvester cluster"),
		mcp.WithString("namespace",
			mcp.Required(),
			mcp.Description("The namespace of the pod"),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("The name of the pod to delete"),
		),
	)
	s.mcpServer.AddTool(deletePodTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace, ok := req.Params.Arguments["namespace"].(string)
		if !ok || namespace == "" {
			return mcp.NewToolResultError("Namespace is required"), nil
		}

		name, ok := req.Params.Arguments["name"].(string)
		if !ok || name == "" {
			return mcp.NewToolResultError("Pod name is required"), nil
		}

		// Use the unified resource handler
		gvr := kubernetes.ResourceTypeToGVR[kubernetes.ResourceTypePod]
		err := s.resourceHandler.DeleteResource(ctx, gvr, namespace, name)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to delete pod %s in namespace %s: %v", name, namespace, err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Pod %s in namespace %s deleted successfully", name, namespace)), nil
	})
}

// registerKubernetesDeploymentTools registers Deployment-related tools.
func (s *HarvesterMCPServer) registerKubernetesDeploymentTools() {
	// List deployments tool
	listDeploymentsTool := mcp.NewTool(
		"list_deployments",
		mcp.WithDescription("List deployments in the Harvester cluster"),
		mcp.WithString("namespace",
			mcp.Description("The namespace to list deployments from (optional, defaults to all namespaces)"),
		),
	)
	s.mcpServer.AddTool(listDeploymentsTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace, _ := req.Params.Arguments["namespace"].(string)

		// Use the unified resource handler
		gvr := kubernetes.ResourceTypeToGVR[kubernetes.ResourceTypeDeployments]
		list, err := s.resourceHandler.ListResources(ctx, gvr, namespace)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list deployments: %v", err)), nil
		}

		// Format the list using the resource formatter
		formatted := s.resourceHandler.FormatResourceList(list, gvr)
		return mcp.NewToolResultText(formatted), nil
	})

	// Get deployment tool
	getDeploymentTool := mcp.NewTool(
		"get_deployment",
		mcp.WithDescription("Get deployment details from the Harvester cluster"),
		mcp.WithString("namespace",
			mcp.Required(),
			mcp.Description("The namespace of the deployment"),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("The name of the deployment"),
		),
	)
	s.mcpServer.AddTool(getDeploymentTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace, ok := req.Params.Arguments["namespace"].(string)
		if !ok || namespace == "" {
			return mcp.NewToolResultError("Namespace is required"), nil
		}

		name, ok := req.Params.Arguments["name"].(string)
		if !ok || name == "" {
			return mcp.NewToolResultError("Deployment name is required"), nil
		}

		// Use the unified resource handler
		gvr := kubernetes.ResourceTypeToGVR[kubernetes.ResourceTypeDeployment]
		resource, err := s.resourceHandler.GetResource(ctx, gvr, namespace, name)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get deployment %s in namespace %s: %v", name, namespace, err)), nil
		}

		// Format the resource using the resource formatter
		formatted := s.resourceHandler.FormatResource(resource, gvr)
		return mcp.NewToolResultText(formatted), nil
	})
}

// registerKubernetesServiceTools registers Service-related tools.
func (s *HarvesterMCPServer) registerKubernetesServiceTools() {
	// List services tool
	listServicesTool := mcp.NewTool(
		"list_services",
		mcp.WithDescription("List services in the Harvester cluster"),
		mcp.WithString("namespace",
			mcp.Description("The namespace to list services from (optional, defaults to all namespaces)"),
		),
	)
	s.mcpServer.AddTool(listServicesTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace, _ := req.Params.Arguments["namespace"].(string)

		// Use the unified resource handler
		gvr := kubernetes.ResourceTypeToGVR[kubernetes.ResourceTypeServices]
		list, err := s.resourceHandler.ListResources(ctx, gvr, namespace)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list services: %v", err)), nil
		}

		// Format the list using the resource formatter
		formatted := s.resourceHandler.FormatResourceList(list, gvr)
		return mcp.NewToolResultText(formatted), nil
	})

	// Get service tool
	getServiceTool := mcp.NewTool(
		"get_service",
		mcp.WithDescription("Get service details from the Harvester cluster"),
		mcp.WithString("namespace",
			mcp.Required(),
			mcp.Description("The namespace of the service"),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("The name of the service"),
		),
	)
	s.mcpServer.AddTool(getServiceTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace, ok := req.Params.Arguments["namespace"].(string)
		if !ok || namespace == "" {
			return mcp.NewToolResultError("Namespace is required"), nil
		}

		name, ok := req.Params.Arguments["name"].(string)
		if !ok || name == "" {
			return mcp.NewToolResultError("Service name is required"), nil
		}

		// Use the unified resource handler
		gvr := kubernetes.ResourceTypeToGVR[kubernetes.ResourceTypeService]
		resource, err := s.resourceHandler.GetResource(ctx, gvr, namespace, name)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get service %s in namespace %s: %v", name, namespace, err)), nil
		}

		// Format the resource using the resource formatter
		formatted := s.resourceHandler.FormatResource(resource, gvr)
		return mcp.NewToolResultText(formatted), nil
	})
}

// registerKubernetesNamespaceTools registers Namespace-related tools.
func (s *HarvesterMCPServer) registerKubernetesNamespaceTools() {
	// List namespaces tool
	listNamespacesTool := mcp.NewTool(
		"list_namespaces",
		mcp.WithDescription("List namespaces in the Harvester cluster"),
	)
	s.mcpServer.AddTool(listNamespacesTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Use the unified resource handler
		gvr := kubernetes.ResourceTypeToGVR[kubernetes.ResourceTypeNamespaces]
		list, err := s.resourceHandler.ListResources(ctx, gvr, "")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list namespaces: %v", err)), nil
		}

		// Format the list using the resource formatter
		formatted := s.resourceHandler.FormatResourceList(list, gvr)
		return mcp.NewToolResultText(formatted), nil
	})

	// Get namespace tool
	getNamespaceTool := mcp.NewTool(
		"get_namespace",
		mcp.WithDescription("Get namespace details from the Harvester cluster"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("The name of the namespace"),
		),
	)
	s.mcpServer.AddTool(getNamespaceTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, ok := req.Params.Arguments["name"].(string)
		if !ok || name == "" {
			return mcp.NewToolResultError("Namespace name is required"), nil
		}

		// Use the unified resource handler
		gvr := kubernetes.ResourceTypeToGVR[kubernetes.ResourceTypeNamespace]
		resource, err := s.resourceHandler.GetResource(ctx, gvr, "", name)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get namespace %s: %v", name, err)), nil
		}

		// Format the resource using the resource formatter
		formatted := s.resourceHandler.FormatResource(resource, gvr)
		return mcp.NewToolResultText(formatted), nil
	})
}

// registerKubernetesNodeTools registers Node-related tools.
func (s *HarvesterMCPServer) registerKubernetesNodeTools() {
	// List nodes tool
	listNodesTool := mcp.NewTool(
		"list_nodes",
		mcp.WithDescription("List nodes in the Harvester cluster"),
	)
	s.mcpServer.AddTool(listNodesTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Use the unified resource handler
		gvr := kubernetes.ResourceTypeToGVR[kubernetes.ResourceTypeNodes]
		list, err := s.resourceHandler.ListResources(ctx, gvr, "")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list nodes: %v", err)), nil
		}

		// Format the list using the resource formatter
		formatted := s.resourceHandler.FormatResourceList(list, gvr)
		return mcp.NewToolResultText(formatted), nil
	})

	// Get node tool
	getNodeTool := mcp.NewTool(
		"get_node",
		mcp.WithDescription("Get node details from the Harvester cluster"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("The name of the node"),
		),
	)
	s.mcpServer.AddTool(getNodeTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, ok := req.Params.Arguments["name"].(string)
		if !ok || name == "" {
			return mcp.NewToolResultError("Node name is required"), nil
		}

		// Use the unified resource handler
		gvr := kubernetes.ResourceTypeToGVR[kubernetes.ResourceTypeNode]
		resource, err := s.resourceHandler.GetResource(ctx, gvr, "", name)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get node %s: %v", name, err)), nil
		}

		// Format the resource using the resource formatter
		formatted := s.resourceHandler.FormatResource(resource, gvr)
		return mcp.NewToolResultText(formatted), nil
	})
}

// registerKubernetesCRDTools registers CRD-related tools.
func (s *HarvesterMCPServer) registerKubernetesCRDTools() {
	// List CRDs tool
	listCRDsTool := mcp.NewTool(
		"list_crds",
		mcp.WithDescription("List Custom Resource Definitions in the Harvester cluster"),
	)
	s.mcpServer.AddTool(listCRDsTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Use the unified resource handler
		gvr := kubernetes.ResourceTypeToGVR[kubernetes.ResourceTypeCRDs]
		list, err := s.resourceHandler.ListResources(ctx, gvr, "")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list CRDs: %v", err)), nil
		}

		// Format the list using the resource formatter
		formatted := s.resourceHandler.FormatResourceList(list, gvr)
		return mcp.NewToolResultText(formatted), nil
	})
}

// registerHarvesterVirtualMachineTools registers Harvester VM-related tools.
func (s *HarvesterMCPServer) registerHarvesterVirtualMachineTools() {
	// List VMs tool
	listVMsTool := mcp.NewTool(
		"list_vms",
		mcp.WithDescription("List Virtual Machines in the Harvester cluster"),
		mcp.WithString("namespace",
			mcp.Description("The namespace to list VMs from (optional, defaults to all namespaces)"),
		),
	)
	s.mcpServer.AddTool(listVMsTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace, _ := req.Params.Arguments["namespace"].(string)

		// Use the unified resource handler
		gvr := kubernetes.ResourceTypeToGVR[kubernetes.ResourceTypeVMs]
		list, err := s.resourceHandler.ListResources(ctx, gvr, namespace)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list VMs: %v", err)), nil
		}

		// Format the list using the resource formatter
		formatted := s.resourceHandler.FormatResourceList(list, gvr)
		return mcp.NewToolResultText(formatted), nil
	})

	// Get VM tool
	getVMTool := mcp.NewTool(
		"get_vm",
		mcp.WithDescription("Get Virtual Machine details from the Harvester cluster"),
		mcp.WithString("namespace",
			mcp.Required(),
			mcp.Description("The namespace of the VM"),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("The name of the VM"),
		),
	)
	s.mcpServer.AddTool(getVMTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace, ok := req.Params.Arguments["namespace"].(string)
		if !ok || namespace == "" {
			return mcp.NewToolResultError("Namespace is required"), nil
		}

		name, ok := req.Params.Arguments["name"].(string)
		if !ok || name == "" {
			return mcp.NewToolResultError("VM name is required"), nil
		}

		// Use the unified resource handler
		gvr := kubernetes.ResourceTypeToGVR[kubernetes.ResourceTypeVM]
		resource, err := s.resourceHandler.GetResource(ctx, gvr, namespace, name)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get VM %s in namespace %s: %v", name, namespace, err)), nil
		}

		// Format the resource using the resource formatter
		formatted := s.resourceHandler.FormatResource(resource, gvr)
		return mcp.NewToolResultText(formatted), nil
	})
}

// registerHarvesterImageTools registers Harvester Image-related tools.
func (s *HarvesterMCPServer) registerHarvesterImageTools() {
	// List images tool
	listImagesTool := mcp.NewTool(
		"list_images",
		mcp.WithDescription("List Images in the Harvester cluster"),
		mcp.WithString("namespace",
			mcp.Description("The namespace to list images from (optional, defaults to all namespaces)"),
		),
	)
	s.mcpServer.AddTool(listImagesTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace, _ := req.Params.Arguments["namespace"].(string)

		// Use the unified resource handler
		gvr := kubernetes.ResourceTypeToGVR[kubernetes.ResourceTypeImages]
		list, err := s.resourceHandler.ListResources(ctx, gvr, namespace)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list images: %v", err)), nil
		}

		// Format the list using the resource formatter
		formatted := s.resourceHandler.FormatResourceList(list, gvr)
		return mcp.NewToolResultText(formatted), nil
	})
}

// registerHarvesterVolumeTools registers Harvester Volume-related tools.
func (s *HarvesterMCPServer) registerHarvesterVolumeTools() {
	// List volumes tool
	listVolumesTool := mcp.NewTool(
		"list_volumes",
		mcp.WithDescription("List Volumes in the Harvester cluster"),
		mcp.WithString("namespace",
			mcp.Description("The namespace to list volumes from (optional, defaults to all namespaces)"),
		),
	)
	s.mcpServer.AddTool(listVolumesTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace, _ := req.Params.Arguments["namespace"].(string)

		// Use the unified resource handler
		gvr := kubernetes.ResourceTypeToGVR[kubernetes.ResourceTypeVolumes]
		list, err := s.resourceHandler.ListResources(ctx, gvr, namespace)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list volumes: %v", err)), nil
		}

		// Format the list using the resource formatter
		formatted := s.resourceHandler.FormatResourceList(list, gvr)
		return mcp.NewToolResultText(formatted), nil
	})
}

// registerHarvesterNetworkTools registers Harvester Network-related tools.
func (s *HarvesterMCPServer) registerHarvesterNetworkTools() {
	// List networks tool
	listNetworksTool := mcp.NewTool(
		"list_networks",
		mcp.WithDescription("List Networks in the Harvester cluster"),
		mcp.WithString("namespace",
			mcp.Description("The namespace to list networks from (optional, defaults to all namespaces)"),
		),
	)
	s.mcpServer.AddTool(listNetworksTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace, _ := req.Params.Arguments["namespace"].(string)

		// Use the unified resource handler
		gvr := kubernetes.ResourceTypeToGVR[kubernetes.ResourceTypeNetworks]
		list, err := s.resourceHandler.ListResources(ctx, gvr, namespace)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list networks: %v", err)), nil
		}

		// Format the list using the resource formatter
		formatted := s.resourceHandler.FormatResourceList(list, gvr)
		return mcp.NewToolResultText(formatted), nil
	})
}
