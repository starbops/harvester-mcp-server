package kubernetes

import (
	"fmt"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// VirtualMachineFormatter handles formatting for VirtualMachine resources
type VirtualMachineFormatter struct{}

func (f *VirtualMachineFormatter) FormatResource(res *unstructured.Unstructured) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Virtual Machine: %s\n", res.GetName()))
	sb.WriteString(fmt.Sprintf("Namespace: %s\n", res.GetNamespace()))

	// Get status
	status := "Unknown"
	running := getNestedBool(res.Object, "status", "ready")
	created := getNestedBool(res.Object, "status", "created")

	if running {
		status = "Running"
	} else if created {
		status = "Created"
	}
	sb.WriteString(fmt.Sprintf("Status: %s\n", status))

	// Running and created fields
	sb.WriteString(fmt.Sprintf("Ready: %t\n", running))
	sb.WriteString(fmt.Sprintf("Created: %t\n", created))

	// Detailed VM specification
	sb.WriteString("\nSpecification:\n")

	// CPU and Memory
	cpuCores := getNestedInt64(res.Object, "spec", "template", "spec", "domain", "cpu", "cores")
	memory := getNestedString(res.Object, "spec", "template", "spec", "domain", "resources", "requests", "memory")

	if cpuCores > 0 {
		sb.WriteString(fmt.Sprintf("  CPU Cores: %d\n", cpuCores))
	}

	if memory != "" {
		sb.WriteString(fmt.Sprintf("  Memory: %s\n", memory))
	}

	// Volumes
	volumes, _, _ := unstructured.NestedSlice(res.Object, "spec", "template", "spec", "volumes")
	if len(volumes) > 0 {
		sb.WriteString("\nVolumes:\n")
		for _, volumeObj := range volumes {
			volume, ok := volumeObj.(map[string]interface{})
			if !ok {
				continue
			}

			name, _, _ := unstructured.NestedString(volume, "name")
			sb.WriteString(fmt.Sprintf("  %s:\n", name))

			// Check different volume types
			if pvc, exists, _ := unstructured.NestedMap(volume, "persistentVolumeClaim"); exists && pvc != nil {
				claimName := getNestedString(volume, "persistentVolumeClaim", "claimName")
				sb.WriteString(fmt.Sprintf("    Type: PersistentVolumeClaim\n"))
				sb.WriteString(fmt.Sprintf("    Claim Name: %s\n", claimName))
			} else if container, exists, _ := unstructured.NestedMap(volume, "containerDisk"); exists && container != nil {
				image := getNestedString(volume, "containerDisk", "image")
				sb.WriteString(fmt.Sprintf("    Type: ContainerDisk\n"))
				sb.WriteString(fmt.Sprintf("    Image: %s\n", image))
			} else if cloudInit, exists, _ := unstructured.NestedMap(volume, "cloudInitNoCloud"); exists && cloudInit != nil {
				sb.WriteString(fmt.Sprintf("    Type: CloudInitNoCloud\n"))
				userData, userDataExists, _ := unstructured.NestedString(cloudInit, "userData")
				if userDataExists && userData != "" {
					sb.WriteString(fmt.Sprintf("    Has User Data: true\n"))
				}
				networkData, networkDataExists, _ := unstructured.NestedString(cloudInit, "networkData")
				if networkDataExists && networkData != "" {
					sb.WriteString(fmt.Sprintf("    Has Network Data: true\n"))
				}
			} else {
				sb.WriteString(fmt.Sprintf("    Type: Other\n"))
			}
		}
	}

	// Networks
	networks, _, _ := unstructured.NestedSlice(res.Object, "spec", "template", "spec", "networks")
	if len(networks) > 0 {
		sb.WriteString("\nNetworks:\n")
		for _, netObj := range networks {
			network, ok := netObj.(map[string]interface{})
			if !ok {
				continue
			}

			name, _, _ := unstructured.NestedString(network, "name")
			sb.WriteString(fmt.Sprintf("  %s:\n", name))

			// Check different network types
			if podNet, exists, _ := unstructured.NestedString(network, "pod"); exists && podNet != "" {
				sb.WriteString(fmt.Sprintf("    Type: Pod Network\n"))
			} else if multus, exists, _ := unstructured.NestedMap(network, "multus"); exists && multus != nil {
				networkName := getNestedString(network, "multus", "networkName")
				sb.WriteString(fmt.Sprintf("    Type: Multus\n"))
				sb.WriteString(fmt.Sprintf("    Network Name: %s\n", networkName))
			} else {
				sb.WriteString(fmt.Sprintf("    Type: Other\n"))
			}
		}
	}

	// Creation time
	creationTime := res.GetCreationTimestamp().Format(time.RFC3339)
	sb.WriteString(fmt.Sprintf("\nCreated: %s\n", creationTime))

	return sb.String()
}

func (f *VirtualMachineFormatter) FormatResourceList(list *unstructured.UnstructuredList) string {
	if len(list.Items) == 0 {
		return "No virtual machines found in the specified namespace(s)."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d virtual machine(s):\n\n", len(list.Items)))

	// Group VMs by namespace
	vmsByNamespace := make(map[string][]unstructured.Unstructured)
	for _, item := range list.Items {
		namespace := item.GetNamespace()
		vmsByNamespace[namespace] = append(vmsByNamespace[namespace], item)
	}

	// Print VMs grouped by namespace
	for namespace, vms := range vmsByNamespace {
		sb.WriteString(fmt.Sprintf("Namespace: %s (%d VMs)\n", namespace, len(vms)))

		for _, vm := range vms {
			// Get status
			status := "Unknown"
			running := getNestedBool(vm.Object, "status", "ready")
			created := getNestedBool(vm.Object, "status", "created")

			if running {
				status = "Running"
			} else if created {
				status = "Created"
			}

			// Get spec details
			cpuCores := getNestedInt64(vm.Object, "spec", "template", "spec", "domain", "cpu", "cores")
			memory := getNestedString(vm.Object, "spec", "template", "spec", "domain", "resources", "requests", "memory")

			// Basic VM info
			sb.WriteString(fmt.Sprintf("  • %s\n", vm.GetName()))
			sb.WriteString(fmt.Sprintf("    Status: %s\n", status))

			if cpuCores > 0 {
				sb.WriteString(fmt.Sprintf("    CPU Cores: %d\n", cpuCores))
			}

			if memory != "" {
				sb.WriteString(fmt.Sprintf("    Memory: %s\n", memory))
			}

			// Volumes
			volumes, _, _ := unstructured.NestedSlice(vm.Object, "spec", "template", "spec", "volumes")
			if len(volumes) > 0 {
				sb.WriteString("    Volumes:\n")
				for _, volumeObj := range volumes {
					volume, ok := volumeObj.(map[string]interface{})
					if !ok {
						continue
					}

					name, _, _ := unstructured.NestedString(volume, "name")

					// Check different volume types
					if pvc, exists, _ := unstructured.NestedMap(volume, "persistentVolumeClaim"); exists && pvc != nil {
						claimName := getNestedString(volume, "persistentVolumeClaim", "claimName")
						sb.WriteString(fmt.Sprintf("      %s: PVC %s\n", name, claimName))
					} else if container, exists, _ := unstructured.NestedMap(volume, "containerDisk"); exists && container != nil {
						image := getNestedString(volume, "containerDisk", "image")
						sb.WriteString(fmt.Sprintf("      %s: ContainerDisk %s\n", name, image))
					} else if cloudInit, exists, _ := unstructured.NestedMap(volume, "cloudInitNoCloud"); exists && cloudInit != nil {
						sb.WriteString(fmt.Sprintf("      %s: CloudInit\n", name))
					} else {
						sb.WriteString(fmt.Sprintf("      %s\n", name))
					}
				}
			}

			// Networks
			networks, _, _ := unstructured.NestedSlice(vm.Object, "spec", "template", "spec", "networks")
			if len(networks) > 0 {
				sb.WriteString("    Networks:\n")
				for _, netObj := range networks {
					network, ok := netObj.(map[string]interface{})
					if !ok {
						continue
					}

					name, _, _ := unstructured.NestedString(network, "name")

					// Check different network types
					if podNet, exists, _ := unstructured.NestedString(network, "pod"); exists && podNet != "" {
						sb.WriteString(fmt.Sprintf("      %s: Pod Network\n", name))
					} else if multus, exists, _ := unstructured.NestedMap(network, "multus"); exists && multus != nil {
						networkName := getNestedString(network, "multus", "networkName")
						sb.WriteString(fmt.Sprintf("      %s: Multus %s\n", name, networkName))
					} else {
						sb.WriteString(fmt.Sprintf("      %s\n", name))
					}
				}
			}

			// Creation time
			creationTime := vm.GetCreationTimestamp().Format(time.RFC3339)
			sb.WriteString(fmt.Sprintf("    Created: %s\n", creationTime))

			sb.WriteString("\n")
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

// VolumeFormatter handles formatting for Volume resources
type VolumeFormatter struct{}

func (f *VolumeFormatter) FormatResource(res *unstructured.Unstructured) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Volume: %s\n", res.GetName()))
	sb.WriteString(fmt.Sprintf("Namespace: %s\n", res.GetNamespace()))

	// Get size and status
	size := getNestedString(res.Object, "spec", "size")
	status := getNestedString(res.Object, "status", "state")

	if status != "" {
		sb.WriteString(fmt.Sprintf("Status: %s\n", status))
	}
	if size != "" {
		sb.WriteString(fmt.Sprintf("Size: %s\n", size))
	}

	// Storage Class
	storageClass := getNestedString(res.Object, "spec", "storageClassName")
	if storageClass != "" {
		sb.WriteString(fmt.Sprintf("Storage Class: %s\n", storageClass))
	}

	// Additional details
	accessModes, _, _ := unstructured.NestedStringSlice(res.Object, "spec", "accessModes")
	if len(accessModes) > 0 {
		sb.WriteString("\nAccess Modes:\n")
		for _, mode := range accessModes {
			sb.WriteString(fmt.Sprintf("  %s\n", mode))
		}
	}

	// Creation time
	creationTime := res.GetCreationTimestamp().Format(time.RFC3339)
	sb.WriteString(fmt.Sprintf("\nCreated: %s\n", creationTime))

	return sb.String()
}

func (f *VolumeFormatter) FormatResourceList(list *unstructured.UnstructuredList) string {
	if len(list.Items) == 0 {
		return "No volumes found in the specified namespace(s)."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d volume(s):\n\n", len(list.Items)))

	// Group volumes by namespace
	volumesByNamespace := make(map[string][]unstructured.Unstructured)
	for _, item := range list.Items {
		namespace := item.GetNamespace()
		volumesByNamespace[namespace] = append(volumesByNamespace[namespace], item)
	}

	// Print volumes grouped by namespace
	for namespace, volumes := range volumesByNamespace {
		sb.WriteString(fmt.Sprintf("Namespace: %s (%d volumes)\n", namespace, len(volumes)))

		for _, volume := range volumes {
			// Get size and status
			size := getNestedString(volume.Object, "spec", "size")
			status := getNestedString(volume.Object, "status", "state")

			// Basic volume info
			sb.WriteString(fmt.Sprintf("  • %s\n", volume.GetName()))
			if status != "" {
				sb.WriteString(fmt.Sprintf("    Status: %s\n", status))
			}
			if size != "" {
				sb.WriteString(fmt.Sprintf("    Size: %s\n", size))
			}

			// Creation time
			creationTime := volume.GetCreationTimestamp().Format(time.RFC3339)
			sb.WriteString(fmt.Sprintf("    Created: %s\n", creationTime))

			sb.WriteString("\n")
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

// NetworkFormatter handles formatting for Network resources
type NetworkFormatter struct{}

func (f *NetworkFormatter) FormatResource(res *unstructured.Unstructured) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Network: %s\n", res.GetName()))
	sb.WriteString(fmt.Sprintf("Namespace: %s\n", res.GetNamespace()))

	// Get network type and config
	networkType := getNestedString(res.Object, "spec", "type")
	if networkType != "" {
		sb.WriteString(fmt.Sprintf("Type: %s\n", networkType))
	}

	// Config details
	config, configFound, _ := unstructured.NestedMap(res.Object, "spec", "config")
	if configFound && len(config) > 0 {
		sb.WriteString("\nConfiguration:\n")
		for key, value := range config {
			sb.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
		}
	}

	// Creation time
	creationTime := res.GetCreationTimestamp().Format(time.RFC3339)
	sb.WriteString(fmt.Sprintf("\nCreated: %s\n", creationTime))

	return sb.String()
}

func (f *NetworkFormatter) FormatResourceList(list *unstructured.UnstructuredList) string {
	if len(list.Items) == 0 {
		return "No networks found in the specified namespace(s)."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d network(s):\n\n", len(list.Items)))

	// Group networks by namespace
	networksByNamespace := make(map[string][]unstructured.Unstructured)
	for _, item := range list.Items {
		namespace := item.GetNamespace()
		networksByNamespace[namespace] = append(networksByNamespace[namespace], item)
	}

	// Print networks grouped by namespace
	for namespace, networks := range networksByNamespace {
		sb.WriteString(fmt.Sprintf("Namespace: %s (%d networks)\n", namespace, len(networks)))

		for _, network := range networks {
			// Get type and config
			networkType := getNestedString(network.Object, "spec", "type")

			// Basic network info
			sb.WriteString(fmt.Sprintf("  • %s\n", network.GetName()))
			if networkType != "" {
				sb.WriteString(fmt.Sprintf("    Type: %s\n", networkType))
			}

			// VLAN ID if present
			vlanId := getNestedInt64(network.Object, "spec", "config", "vlan")
			if vlanId > 0 {
				sb.WriteString(fmt.Sprintf("    VLAN ID: %d\n", vlanId))
			}

			// Creation time
			creationTime := network.GetCreationTimestamp().Format(time.RFC3339)
			sb.WriteString(fmt.Sprintf("    Created: %s\n", creationTime))

			sb.WriteString("\n")
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

// VMImageFormatter handles formatting for VirtualMachineImage resources
type VMImageFormatter struct{}

func (f *VMImageFormatter) FormatResource(res *unstructured.Unstructured) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("VM Image: %s\n", res.GetName()))
	sb.WriteString(fmt.Sprintf("Namespace: %s\n", res.GetNamespace()))

	// Get image details
	displayName := getNestedString(res.Object, "spec", "displayName")
	url := getNestedString(res.Object, "spec", "url")
	description := getNestedString(res.Object, "spec", "description")

	if displayName != "" {
		sb.WriteString(fmt.Sprintf("Display Name: %s\n", displayName))
	}
	if url != "" {
		sb.WriteString(fmt.Sprintf("URL: %s\n", url))
	}
	if description != "" {
		sb.WriteString(fmt.Sprintf("Description: %s\n", description))
	}

	// Status details
	status, statusFound, _ := unstructured.NestedMap(res.Object, "status")
	if statusFound && len(status) > 0 {
		sb.WriteString("\nStatus:\n")

		state := getNestedString(res.Object, "status", "state")
		if state != "" {
			sb.WriteString(fmt.Sprintf("  State: %s\n", state))
		}

		progress := getNestedString(res.Object, "status", "progress")
		if progress != "" {
			sb.WriteString(fmt.Sprintf("  Progress: %s\n", progress))
		}

		size := getNestedString(res.Object, "status", "size")
		if size != "" {
			sb.WriteString(fmt.Sprintf("  Size: %s\n", size))
		}
	}

	// Creation time
	creationTime := res.GetCreationTimestamp().Format(time.RFC3339)
	sb.WriteString(fmt.Sprintf("\nCreated: %s\n", creationTime))

	return sb.String()
}

func (f *VMImageFormatter) FormatResourceList(list *unstructured.UnstructuredList) string {
	if len(list.Items) == 0 {
		return "No VM images found in the specified namespace(s)."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d VM image(s):\n\n", len(list.Items)))

	// Group images by namespace
	imagesByNamespace := make(map[string][]unstructured.Unstructured)
	for _, item := range list.Items {
		namespace := item.GetNamespace()
		imagesByNamespace[namespace] = append(imagesByNamespace[namespace], item)
	}

	// Print images grouped by namespace
	for namespace, images := range imagesByNamespace {
		sb.WriteString(fmt.Sprintf("Namespace: %s (%d images)\n", namespace, len(images)))

		for _, image := range images {
			// Get basic info
			url := getNestedString(image.Object, "spec", "displayName")
			if url == "" {
				url = getNestedString(image.Object, "spec", "url")
			}

			size := getNestedString(image.Object, "status", "size")
			progress := getNestedString(image.Object, "status", "progress")

			// Basic image info
			sb.WriteString(fmt.Sprintf("  • %s\n", image.GetName()))
			if url != "" {
				sb.WriteString(fmt.Sprintf("    Source: %s\n", url))
			}
			if size != "" {
				sb.WriteString(fmt.Sprintf("    Size: %s\n", size))
			}
			if progress != "" {
				sb.WriteString(fmt.Sprintf("    Progress: %s\n", progress))
			}

			// Creation time
			creationTime := image.GetCreationTimestamp().Format(time.RFC3339)
			sb.WriteString(fmt.Sprintf("    Created: %s\n", creationTime))

			sb.WriteString("\n")
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

// CRDFormatter handles formatting for CustomResourceDefinition resources
type CRDFormatter struct{}

func (f *CRDFormatter) FormatResource(res *unstructured.Unstructured) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Custom Resource Definition: %s\n", res.GetName()))

	// Get CRD details
	group := getNestedString(res.Object, "spec", "group")
	kind := getNestedString(res.Object, "spec", "names", "kind")
	plural := getNestedString(res.Object, "spec", "names", "plural")
	scope := getNestedString(res.Object, "spec", "scope")

	sb.WriteString(fmt.Sprintf("Group: %s\n", group))
	sb.WriteString(fmt.Sprintf("Kind: %s\n", kind))
	sb.WriteString(fmt.Sprintf("Plural: %s\n", plural))
	sb.WriteString(fmt.Sprintf("Scope: %s\n", scope))

	// Short names
	shortNames, _, _ := unstructured.NestedStringSlice(res.Object, "spec", "names", "shortNames")
	if len(shortNames) > 0 {
		sb.WriteString("Short Names: ")
		for i, name := range shortNames {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(name)
		}
		sb.WriteString("\n")
	}

	// Versions
	versions, _, _ := unstructured.NestedSlice(res.Object, "spec", "versions")
	if len(versions) > 0 {
		sb.WriteString("\nVersions:\n")
		for _, versionObj := range versions {
			version, ok := versionObj.(map[string]interface{})
			if !ok {
				continue
			}

			name, _, _ := unstructured.NestedString(version, "name")
			served, _, _ := unstructured.NestedBool(version, "served")
			storage, _, _ := unstructured.NestedBool(version, "storage")

			sb.WriteString(fmt.Sprintf("  %s:\n", name))
			sb.WriteString(fmt.Sprintf("    Served: %t\n", served))
			sb.WriteString(fmt.Sprintf("    Storage: %t\n", storage))

			// Schema details if present
			schema, schemaFound, _ := unstructured.NestedMap(version, "schema", "openAPIV3Schema")
			if schemaFound && len(schema) > 0 {
				sb.WriteString("    Schema: Available\n")
			}
		}
	}

	// Creation time
	creationTime := res.GetCreationTimestamp().Format(time.RFC3339)
	sb.WriteString(fmt.Sprintf("\nCreated: %s\n", creationTime))

	return sb.String()
}

func (f *CRDFormatter) FormatResourceList(list *unstructured.UnstructuredList) string {
	if len(list.Items) == 0 {
		return "No custom resource definitions found."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d custom resource definition(s):\n\n", len(list.Items)))

	// Group CRDs by group
	crdsByGroup := make(map[string][]unstructured.Unstructured)
	for _, item := range list.Items {
		group := getNestedString(item.Object, "spec", "group")
		if group == "" {
			group = "core"
		}
		crdsByGroup[group] = append(crdsByGroup[group], item)
	}

	// Print CRDs grouped by group
	for group, crds := range crdsByGroup {
		sb.WriteString(fmt.Sprintf("Group: %s (%d CRDs)\n", group, len(crds)))

		for _, crd := range crds {
			// Get basic info
			kind := getNestedString(crd.Object, "spec", "names", "kind")
			plural := getNestedString(crd.Object, "spec", "names", "plural")

			// Basic CRD info
			sb.WriteString(fmt.Sprintf("  • %s\n", crd.GetName()))
			sb.WriteString(fmt.Sprintf("    Kind: %s\n", kind))
			sb.WriteString(fmt.Sprintf("    Plural: %s\n", plural))

			// Versions
			versions, _, _ := unstructured.NestedSlice(crd.Object, "spec", "versions")
			if len(versions) > 0 {
				sb.WriteString("    Versions:\n")
				for _, versionObj := range versions {
					version, ok := versionObj.(map[string]interface{})
					if !ok {
						continue
					}

					name, _, _ := unstructured.NestedString(version, "name")
					served, _, _ := unstructured.NestedBool(version, "served")
					storage, _, _ := unstructured.NestedBool(version, "storage")

					sb.WriteString(fmt.Sprintf("      %s (served: %t, storage: %t)\n", name, served, storage))
				}
			}

			// Creation time
			creationTime := crd.GetCreationTimestamp().Format(time.RFC3339)
			sb.WriteString(fmt.Sprintf("    Created: %s\n", creationTime))

			sb.WriteString("\n")
		}

		sb.WriteString("\n")
	}

	return sb.String()
}
