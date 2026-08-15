package cost

// CostCategory groups resource costs by function.
type CostCategory string

const (
	CostCategoryCompute    CostCategory = "Compute"
	CostCategoryDatabase   CostCategory = "Database"
	CostCategoryStorage    CostCategory = "Storage"
	CostCategoryNetworking CostCategory = "Networking"
	CostCategoryKubernetes CostCategory = "Kubernetes"
	CostCategoryServerless CostCategory = "Serverless"
	CostCategoryOther      CostCategory = "Other"
)

// HoursPerMonth is the standard AWS billing hours per month.
const HoursPerMonth = 730.0

// Instance, storage, and flat-rate service prices live in the embedded
// catalog data/aws_pricing.json (see pricing.go).

// resourceCategory maps resource types to cost categories.
var resourceCategory = map[string]CostCategory{
	"ec2_instance":      CostCategoryCompute,
	"compute_instance":  CostCategoryCompute,
	"rds_instance":      CostCategoryDatabase,
	"oci_database":      CostCategoryDatabase,
	"s3_bucket":         CostCategoryStorage,
	"object_storage":    CostCategoryStorage,
	"ebs_volume":        CostCategoryStorage,
	"block_volume":      CostCategoryStorage,
	"nat_gateway":       CostCategoryNetworking,
	"alb":               CostCategoryNetworking,
	"nlb":               CostCategoryNetworking,
	"oci_load_balancer": CostCategoryNetworking,
	"eks_cluster":       CostCategoryKubernetes,
	"oke_cluster":       CostCategoryKubernetes,
	"lambda_function":   CostCategoryServerless,
}

// GetCategory returns the cost category for a resource type.
func GetCategory(resourceType string) CostCategory {
	if cat, ok := resourceCategory[resourceType]; ok {
		return cat
	}
	return CostCategoryOther
}
