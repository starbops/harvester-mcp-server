package kubernetes

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Define constants for supported resource types
const (
	ResourceTypePod         = "pod"
	ResourceTypePods        = "pods"
	ResourceTypeDeployment  = "deployment"
	ResourceTypeDeployments = "deployments"
	ResourceTypeService     = "service"
	ResourceTypeServices    = "services"
	ResourceTypeNamespace   = "namespace"
	ResourceTypeNamespaces  = "namespaces"
	ResourceTypeNode        = "node"
	ResourceTypeNodes       = "nodes"
	ResourceTypeCRD         = "crd"
	ResourceTypeCRDs        = "crds"
	ResourceTypeVM          = "vm"
	ResourceTypeVMs         = "vms"
	ResourceTypeVolume      = "volume"
	ResourceTypeVolumes     = "volumes"
	ResourceTypeNetwork     = "network"
	ResourceTypeNetworks    = "networks"
	ResourceTypeImage       = "image"
	ResourceTypeImages      = "images"
)

// ResourceTypeToGVR maps friendly resource type names to GroupVersionResource
var ResourceTypeToGVR = map[string]schema.GroupVersionResource{
	// Core Kubernetes resources
	ResourceTypePod:         {Group: "", Version: "v1", Resource: "pods"},
	ResourceTypePods:        {Group: "", Version: "v1", Resource: "pods"},
	ResourceTypeService:     {Group: "", Version: "v1", Resource: "services"},
	ResourceTypeServices:    {Group: "", Version: "v1", Resource: "services"},
	ResourceTypeNamespace:   {Group: "", Version: "v1", Resource: "namespaces"},
	ResourceTypeNamespaces:  {Group: "", Version: "v1", Resource: "namespaces"},
	ResourceTypeNode:        {Group: "", Version: "v1", Resource: "nodes"},
	ResourceTypeNodes:       {Group: "", Version: "v1", Resource: "nodes"},
	ResourceTypeDeployment:  {Group: "apps", Version: "v1", Resource: "deployments"},
	ResourceTypeDeployments: {Group: "apps", Version: "v1", Resource: "deployments"},
	ResourceTypeCRD:         {Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"},
	ResourceTypeCRDs:        {Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"},

	// Harvester-specific resources
	ResourceTypeVM:       {Group: "kubevirt.io", Version: "v1", Resource: "virtualmachines"},
	ResourceTypeVMs:      {Group: "kubevirt.io", Version: "v1", Resource: "virtualmachines"},
	ResourceTypeVolume:   {Group: "storage.harvesterhci.io", Version: "v1beta1", Resource: "volumes"},
	ResourceTypeVolumes:  {Group: "storage.harvesterhci.io", Version: "v1beta1", Resource: "volumes"},
	ResourceTypeNetwork:  {Group: "network.harvesterhci.io", Version: "v1beta1", Resource: "networks"},
	ResourceTypeNetworks: {Group: "network.harvesterhci.io", Version: "v1beta1", Resource: "networks"},
	ResourceTypeImage:    {Group: "harvesterhci.io", Version: "v1beta1", Resource: "virtualmachineimages"},
	ResourceTypeImages:   {Group: "harvesterhci.io", Version: "v1beta1", Resource: "virtualmachineimages"},
}

// GVRToResourceType maps GroupVersionResource to friendly resource type names
var GVRToResourceType = map[schema.GroupVersionResource]string{
	// Core Kubernetes resources
	{Group: "", Version: "v1", Resource: "pods"}:                                          ResourceTypePod,
	{Group: "", Version: "v1", Resource: "services"}:                                      ResourceTypeService,
	{Group: "", Version: "v1", Resource: "namespaces"}:                                    ResourceTypeNamespace,
	{Group: "", Version: "v1", Resource: "nodes"}:                                         ResourceTypeNode,
	{Group: "apps", Version: "v1", Resource: "deployments"}:                               ResourceTypeDeployment,
	{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"}: ResourceTypeCRD,

	// Harvester-specific resources
	{Group: "kubevirt.io", Version: "v1", Resource: "virtualmachines"}:               ResourceTypeVM,
	{Group: "storage.harvesterhci.io", Version: "v1beta1", Resource: "volumes"}:      ResourceTypeVolume,
	{Group: "network.harvesterhci.io", Version: "v1beta1", Resource: "networks"}:     ResourceTypeNetwork,
	{Group: "harvesterhci.io", Version: "v1beta1", Resource: "virtualmachineimages"}: ResourceTypeImage,
}
