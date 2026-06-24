package sizing

import (
	"log"

	"github.com/chxmxii/a3/internal/storage"
)

// Analyzer computes sizing information from discovered resources.
type Analyzer struct {
	store *storage.Store
}

// NewAnalyzer creates a new sizing analyzer.
func NewAnalyzer(store *storage.Store) *Analyzer {
	return &Analyzer{store: store}
}

// Analyze processes all resources in an assessment and stores sizing data.
func (a *Analyzer) Analyze(assessmentID string) (*SizingSummary, error) {
	resources, err := a.store.GetResourcesByAssessment(assessmentID)
	if err != nil {
		return nil, err
	}

	summary := &SizingSummary{
		ByCategory: make(map[SizingCategory]CategorySizing),
	}

	for _, res := range resources {
		entry := a.analyzeResource(res)
		if entry == nil {
			continue
		}

		entry.AssessmentID = assessmentID
		if err := a.store.InsertSizing(entry); err != nil {
			log.Printf("[sizing] failed to store sizing for %s: %v", res.ResourceID, err)
			continue
		}

		cat := SizingCategory(entry.Category)
		cs := summary.ByCategory[cat]
		cs.Count++

		if vcpus, ok := entry.Data["vcpus"].(int); ok {
			cs.VCPUs += vcpus
			summary.TotalVCPUs += vcpus
		}
		if mem, ok := entry.Data["memory_gb"].(float64); ok {
			cs.MemoryGB += mem
			summary.TotalMemoryGB += mem
		}
		if stor, ok := entry.Data["storage_gb"].(float64); ok {
			cs.StorageGB += stor
			summary.TotalStorageGB += stor
		}

		summary.ByCategory[cat] = cs
	}

	return summary, nil
}

func (a *Analyzer) analyzeResource(res storage.Resource) *storage.SizingEntry {
	switch res.ResourceType {
	case "ec2_instance":
		return a.analyzeEC2(res)
	case "rds_instance":
		return a.analyzeRDS(res)
	case "eks_cluster", "eks_node_group", "aks_cluster":
		return a.analyzeKubernetes(res)
	case "ebs_volume":
		return a.analyzeEBS(res)
	case "s3_bucket", "object_storage", "storage_account":
		return a.analyzeStorage(res)
	case "virtual_machine":
		return a.analyzeAzureVM(res)
	case "managed_disk":
		return a.analyzeManagedDisk(res)
	case "compute_instance":
		return a.analyzeOCICompute(res)
	case "oci_database":
		return a.analyzeOCIDB(res)
	case "block_volume":
		return a.analyzeBlockVolume(res)
	case "lambda_function":
		return a.analyzeLambda(res)
	default:
		return nil
	}
}

func (a *Analyzer) analyzeEC2(res storage.Resource) *storage.SizingEntry {
	instanceType := getStr(res.RawMetadata, "instance_type")
	if instanceType == "" {
		instanceType = getStr(res.RawMetadata, "instanceType")
	}

	data := map[string]any{
		"instance_type": instanceType,
		"state":         getStr(res.RawMetadata, "instance_state"),
	}

	if spec := GetInstanceSpec(instanceType); spec != nil {
		data["vcpus"] = spec.VCPUs
		data["memory_gb"] = spec.MemGB
		data["family"] = spec.Family
	}

	return &storage.SizingEntry{
		Category:   string(CategoryCompute),
		ResourceID: res.ResourceID,
		Data:       data,
	}
}

func (a *Analyzer) analyzeRDS(res storage.Resource) *storage.SizingEntry {
	instanceClass := getStr(res.RawMetadata, "db_instance_class")
	if instanceClass == "" {
		instanceClass = getStr(res.RawMetadata, "instanceClass")
	}

	data := map[string]any{
		"instance_class": instanceClass,
		"engine":         getStr(res.RawMetadata, "engine"),
		"multi_az":       res.RawMetadata["multi_az"],
	}

	if allocatedStorage, ok := res.RawMetadata["allocated_storage"].(float64); ok {
		data["storage_gb"] = allocatedStorage
	}

	if spec := GetInstanceSpec(instanceClass); spec != nil {
		data["vcpus"] = spec.VCPUs
		data["memory_gb"] = spec.MemGB
	}

	return &storage.SizingEntry{
		Category:   string(CategoryDatabase),
		ResourceID: res.ResourceID,
		Data:       data,
	}
}

func (a *Analyzer) analyzeKubernetes(res storage.Resource) *storage.SizingEntry {
	data := map[string]any{
		"name":               getStr(res.RawMetadata, "name"),
		"kubernetes_version": getStr(res.RawMetadata, "version"),
	}

	return &storage.SizingEntry{
		Category:   string(CategoryKubernetes),
		ResourceID: res.ResourceID,
		Data:       data,
	}
}

func (a *Analyzer) analyzeEBS(res storage.Resource) *storage.SizingEntry {
	data := map[string]any{
		"volume_type": getStr(res.RawMetadata, "volume_type"),
	}

	if size, ok := res.RawMetadata["size"].(float64); ok {
		data["storage_gb"] = size
	}

	return &storage.SizingEntry{
		Category:   string(CategoryStorage),
		ResourceID: res.ResourceID,
		Data:       data,
	}
}

func (a *Analyzer) analyzeStorage(res storage.Resource) *storage.SizingEntry {
	return &storage.SizingEntry{
		Category:   string(CategoryStorage),
		ResourceID: res.ResourceID,
		Data: map[string]any{
			"name": res.Name,
		},
	}
}

func (a *Analyzer) analyzeOCICompute(res storage.Resource) *storage.SizingEntry {
	shape := getStr(res.RawMetadata, "shape")
	data := map[string]any{
		"shape": shape,
		"state": getStr(res.RawMetadata, "lifecycle_state"),
	}

	// OCI shape naming: VM.Standard.E4.Flex, VM.Standard2.1, etc.
	// Extract OCPU count from shape config if available.
	if shapeConfig, ok := res.RawMetadata["shape_config"].(map[string]any); ok {
		if ocpus, ok := shapeConfig["ocpus"].(float64); ok {
			data["vcpus"] = int(ocpus * 2) // OCI OCPUs ≈ 2 vCPUs
			data["memory_gb"] = ocpus * 16  // Default memory ratio
		}
		if mem, ok := shapeConfig["memory_in_gbs"].(float64); ok {
			data["memory_gb"] = mem
		}
	}

	return &storage.SizingEntry{
		Category:   string(CategoryCompute),
		ResourceID: res.ResourceID,
		Data:       data,
	}
}

func (a *Analyzer) analyzeOCIDB(res storage.Resource) *storage.SizingEntry {
	data := map[string]any{
		"shape":   getStr(res.RawMetadata, "shape"),
		"edition": getStr(res.RawMetadata, "database_edition"),
	}

	if nodeCount, ok := res.RawMetadata["node_count"].(float64); ok {
		data["node_count"] = int(nodeCount)
	}

	return &storage.SizingEntry{
		Category:   string(CategoryDatabase),
		ResourceID: res.ResourceID,
		Data:       data,
	}
}

func (a *Analyzer) analyzeBlockVolume(res storage.Resource) *storage.SizingEntry {
	data := map[string]any{}

	if size, ok := res.RawMetadata["size_in_gbs"].(float64); ok {
		data["storage_gb"] = size
	}

	return &storage.SizingEntry{
		Category:   string(CategoryStorage),
		ResourceID: res.ResourceID,
		Data:       data,
	}
}

func (a *Analyzer) analyzeLambda(res storage.Resource) *storage.SizingEntry {
	data := map[string]any{
		"runtime": getStr(res.RawMetadata, "runtime"),
	}

	if memSize, ok := res.RawMetadata["memory_size"].(float64); ok {
		data["memory_gb"] = memSize / 1024.0
	}

	return &storage.SizingEntry{
		Category:   string(CategoryCompute),
		ResourceID: res.ResourceID,
		Data:       data,
	}
}

func (a *Analyzer) analyzeAzureVM(res storage.Resource) *storage.SizingEntry {
	data := map[string]any{
		"size":  getStr(res.RawMetadata, "size"),
		"state": getStr(res.RawMetadata, "power_state"),
	}

	return &storage.SizingEntry{
		Category:   string(CategoryCompute),
		ResourceID: res.ResourceID,
		Data:       data,
	}
}

func (a *Analyzer) analyzeManagedDisk(res storage.Resource) *storage.SizingEntry {
	data := map[string]any{
		"sku": getStr(res.RawMetadata, "sku_name"),
	}

	if size, ok := res.RawMetadata["disk_size_gb"].(float64); ok {
		data["storage_gb"] = size
	}

	return &storage.SizingEntry{
		Category:   string(CategoryStorage),
		ResourceID: res.ResourceID,
		Data:       data,
	}
}

func getStr(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}
