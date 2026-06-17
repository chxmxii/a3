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

// pricingCatalog maps instance types to hourly on-demand prices (USD, us-east-1).
var pricingCatalog = map[string]float64{
	// T2
	"t2.nano":    0.0058,
	"t2.micro":   0.0116,
	"t2.small":   0.023,
	"t2.medium":  0.0464,
	"t2.large":   0.0928,
	"t2.xlarge":  0.1856,
	"t2.2xlarge": 0.3712,

	// T3
	"t3.nano":    0.0052,
	"t3.micro":   0.0104,
	"t3.small":   0.0208,
	"t3.medium":  0.0416,
	"t3.large":   0.0832,
	"t3.xlarge":  0.1664,
	"t3.2xlarge": 0.3328,

	// T3a
	"t3a.nano":    0.0047,
	"t3a.micro":   0.0094,
	"t3a.small":   0.0188,
	"t3a.medium":  0.0376,
	"t3a.large":   0.0752,
	"t3a.xlarge":  0.1504,
	"t3a.2xlarge": 0.3008,

	// M5
	"m5.large":    0.096,
	"m5.xlarge":   0.192,
	"m5.2xlarge":  0.384,
	"m5.4xlarge":  0.768,
	"m5.8xlarge":  1.536,
	"m5.12xlarge": 2.304,
	"m5.16xlarge": 3.072,
	"m5.24xlarge": 4.608,

	// M5a
	"m5a.large":   0.086,
	"m5a.xlarge":  0.172,
	"m5a.2xlarge": 0.344,
	"m5a.4xlarge": 0.688,

	// M6i
	"m6i.large":    0.096,
	"m6i.xlarge":   0.192,
	"m6i.2xlarge":  0.384,
	"m6i.4xlarge":  0.768,
	"m6i.8xlarge":  1.536,
	"m6i.12xlarge": 2.304,
	"m6i.16xlarge": 3.072,

	// M6g (Graviton)
	"m6g.medium":  0.0385,
	"m6g.large":   0.077,
	"m6g.xlarge":  0.154,
	"m6g.2xlarge": 0.308,
	"m6g.4xlarge": 0.616,
	"m6g.8xlarge": 1.232,

	// M7i
	"m7i.large":   0.1008,
	"m7i.xlarge":  0.2016,
	"m7i.2xlarge": 0.4032,
	"m7i.4xlarge": 0.8064,

	// C5
	"c5.large":    0.085,
	"c5.xlarge":   0.170,
	"c5.2xlarge":  0.340,
	"c5.4xlarge":  0.680,
	"c5.9xlarge":  1.530,
	"c5.18xlarge": 3.060,

	// C6i
	"c6i.large":   0.085,
	"c6i.xlarge":  0.170,
	"c6i.2xlarge": 0.340,
	"c6i.4xlarge": 0.680,
	"c6i.8xlarge": 1.360,

	// C6g (Graviton)
	"c6g.medium":  0.034,
	"c6g.large":   0.068,
	"c6g.xlarge":  0.136,
	"c6g.2xlarge": 0.272,
	"c6g.4xlarge": 0.544,

	// R5
	"r5.large":    0.126,
	"r5.xlarge":   0.252,
	"r5.2xlarge":  0.504,
	"r5.4xlarge":  1.008,
	"r5.8xlarge":  2.016,
	"r5.12xlarge": 3.024,

	// R5a
	"r5a.large":   0.113,
	"r5a.xlarge":  0.226,
	"r5a.2xlarge": 0.452,
	"r5a.4xlarge": 0.904,
	"r5a.8xlarge": 1.808,

	// R6i
	"r6i.large":   0.126,
	"r6i.xlarge":  0.252,
	"r6i.2xlarge": 0.504,
	"r6i.4xlarge": 1.008,

	// R6g (Graviton)
	"r6g.medium":  0.0504,
	"r6g.large":   0.1008,
	"r6g.xlarge":  0.2016,
	"r6g.2xlarge": 0.4032,

	// RDS
	"db.t2.micro":   0.017,
	"db.t2.small":   0.034,
	"db.t2.medium":  0.068,
	"db.t3.micro":   0.017,
	"db.t3.small":   0.034,
	"db.t3.medium":  0.068,
	"db.t3.large":   0.136,
	"db.m5.large":   0.171,
	"db.m5.xlarge":  0.342,
	"db.m5.2xlarge": 0.684,
	"db.m5.4xlarge": 1.368,
	"db.m6i.large":  0.171,
	"db.m6i.xlarge": 0.342,
	"db.r5.large":   0.240,
	"db.r5.xlarge":  0.480,
	"db.r5.2xlarge": 0.960,
	"db.r6i.large":  0.240,
	"db.r6i.xlarge": 0.480,
}

// storagePricing maps storage types to monthly per-GB prices.
var storagePricing = map[string]float64{
	"gp2":      0.10,
	"gp3":      0.08,
	"io1":      0.125,
	"io2":      0.125,
	"st1":      0.045,
	"sc1":      0.015,
	"standard": 0.05,
}

// natGatewayHourly is the hourly price for a NAT gateway.
const natGatewayHourly = 0.045

// albHourly is the hourly price for an ALB.
const albHourly = 0.0225

// nlbHourly is the hourly price for an NLB.
const nlbHourly = 0.0225

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
	"alb":              CostCategoryNetworking,
	"nlb":              CostCategoryNetworking,
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
