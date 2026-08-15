package sizing

import "github.com/chxmxii/a3/internal/cost"

// SizingCategory groups resources by their function.
type SizingCategory string

const (
	CategoryCompute    SizingCategory = "compute"
	CategoryKubernetes SizingCategory = "kubernetes"
	CategoryDatabase   SizingCategory = "database"
	CategoryStorage    SizingCategory = "storage"
)

// InstanceSpec describes a compute instance's specifications.
type InstanceSpec struct {
	VCPUs  int
	MemGB  float64
	Family string
}

// SizingSummary aggregates sizing data across all resources.
type SizingSummary struct {
	TotalVCPUs     int
	TotalMemoryGB  float64
	TotalStorageGB float64
	ByCategory     map[SizingCategory]CategorySizing
}

// CategorySizing holds sizing totals for a category.
type CategorySizing struct {
	Count     int
	VCPUs     int
	MemoryGB  float64
	StorageGB float64
}

// GetInstanceSpec returns the spec for a given instance type, or nil if
// unknown. Specs come from the shared embedded catalog in internal/cost
// (data/aws_pricing.json); catalog rows that carry only a price and no
// hardware specs are treated as unknown here.
func GetInstanceSpec(instanceType string) *InstanceSpec {
	row, ok := cost.LookupInstance(instanceType)
	if !ok || row.Family == "" {
		return nil
	}
	return &InstanceSpec{
		VCPUs:  row.VCPUs,
		MemGB:  row.MemGB,
		Family: row.Family,
	}
}
