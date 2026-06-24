package provider

import (
	"context"
	"time"
)

// ResourceType identifies a specific kind of cloud resource.
type ResourceType string

// Provider is the top-level interface each cloud provider implements.
type Provider interface {
	Name() string
	Authenticate(ctx context.Context) error
	Discoverer() Discoverer
	MetricsClient() MetricsClient
	PricingClient() PricingClient
}

// Discoverer handles resource enumeration for a provider.
type Discoverer interface {
	DiscoverResources(ctx context.Context, regions []string, results chan<- DiscoveredResource) error
	SupportedResourceTypes() []ResourceType
}

// DiscoveredResource represents a single resource found during discovery.
type DiscoveredResource struct {
	ProviderType string
	ResourceType ResourceType
	ResourceID   string
	Region       string
	Name         string
	Tags         map[string]string
	RawMetadata  map[string]any
}

// MetricsClient retrieves utilization metrics for resources.
type MetricsClient interface {
	GetCPUUtilization(ctx context.Context, resourceID string, region string) (float64, error)
	GetMemoryUtilization(ctx context.Context, resourceID string, region string) (float64, error)
	GetNetworkTraffic(ctx context.Context, resourceID string, region string) (int64, int64, error)
}

// PricingClient retrieves cost information for resources.
type PricingClient interface {
	GetOnDemandPrice(ctx context.Context, req PricingRequest) (PricingResponse, error)
}

// PricingRequest describes what to price.
type PricingRequest struct {
	ResourceType ResourceType
	Region       string
	InstanceType string
	Attributes   map[string]string
}

// PricingResponse contains pricing data.
type PricingResponse struct {
	HourlyPrice float64
	Currency    string
	Confidence  PricingConfidence
	LastUpdated time.Time
}

// PricingConfidence indicates how reliable a price estimate is.
type PricingConfidence string

const (
	PricingConfidenceHigh   PricingConfidence = "high"
	PricingConfidenceMedium PricingConfidence = "medium"
	PricingConfidenceLow    PricingConfidence = "low"
)

// Common AWS resource types
const (
	ResourceTypeEC2Instance   ResourceType = "ec2_instance"
	ResourceTypeEKSCluster    ResourceType = "eks_cluster"
	ResourceTypeECSCluster    ResourceType = "ecs_cluster"
	ResourceTypeLambda        ResourceType = "lambda_function"
	ResourceTypeRDS           ResourceType = "rds_instance"
	ResourceTypeS3Bucket      ResourceType = "s3_bucket"
	ResourceTypeVPC           ResourceType = "vpc"
	ResourceTypeSubnet        ResourceType = "subnet"
	ResourceTypeSecurityGroup ResourceType = "security_group"
	ResourceTypeRouteTable    ResourceType = "route_table"
	ResourceTypeIGW           ResourceType = "internet_gateway"
	ResourceTypeNATGW         ResourceType = "nat_gateway"
	ResourceTypeTGW           ResourceType = "transit_gateway"
	ResourceTypeALB           ResourceType = "alb"
	ResourceTypeNLB           ResourceType = "nlb"
	ResourceTypeIAMUser       ResourceType = "iam_user"
	ResourceTypeIAMRole       ResourceType = "iam_role"
	ResourceTypeIAMPolicy     ResourceType = "iam_policy"
	ResourceTypeRoute53Zone   ResourceType = "route53_zone"
	ResourceTypeKMSKey        ResourceType = "kms_key"
	ResourceTypeSecret        ResourceType = "secret"
	ResourceTypeOrganization  ResourceType = "organization"
	ResourceTypeAccount       ResourceType = "account"
	ResourceTypeEBSVolume     ResourceType = "ebs_volume"
	ResourceTypeTargetGroup   ResourceType = "target_group"
	ResourceTypeEKSNodeGroup  ResourceType = "eks_node_group"
	ResourceTypeEFS          ResourceType = "efs_file_system"
	ResourceTypeASG          ResourceType = "autoscaling_group"
)

// Common OCI resource types
const (
	ResourceTypeCompartment     ResourceType = "compartment"
	ResourceTypeVCN             ResourceType = "vcn"
	ResourceTypeOCISubnet       ResourceType = "oci_subnet"
	ResourceTypeOCIRouteTable   ResourceType = "oci_route_table"
	ResourceTypeSecurityList    ResourceType = "security_list"
	ResourceTypeNSG             ResourceType = "nsg"
	ResourceTypeDRG             ResourceType = "drg"
	ResourceTypeOCIIGW          ResourceType = "oci_internet_gateway"
	ResourceTypeOCINATGW        ResourceType = "oci_nat_gateway"
	ResourceTypeServiceGateway  ResourceType = "service_gateway"
	ResourceTypeComputeInstance ResourceType = "compute_instance"
	ResourceTypeOKECluster      ResourceType = "oke_cluster"
	ResourceTypeOCILB           ResourceType = "oci_load_balancer"
	ResourceTypeOCIDB           ResourceType = "oci_database"
	ResourceTypeObjectStorage   ResourceType = "object_storage"
	ResourceTypeOCIVault        ResourceType = "oci_vault"
	ResourceTypeBlockVolume     ResourceType = "block_volume"
	ResourceTypeOKENodePool     ResourceType = "oke_node_pool"
	ResourceTypeOCIUser         ResourceType = "oci_user"
	ResourceTypeOCIGroup        ResourceType = "oci_group"
	ResourceTypeOCIPolicy       ResourceType = "oci_policy"
	ResourceTypeOCILogGroup     ResourceType = "oci_log_group"
	ResourceTypeOCIAlarm        ResourceType = "oci_alarm"
)

// Common Azure resource types
const (
	ResourceTypeResourceGroup      ResourceType = "resource_group"
	ResourceTypeVNet               ResourceType = "vnet"
	ResourceTypeAzureSubnet        ResourceType = "azure_subnet"
	ResourceTypeAzureNSG           ResourceType = "azure_nsg"
	ResourceTypeAzureRouteTable    ResourceType = "azure_route_table"
	ResourceTypeAzureNATGW         ResourceType = "azure_nat_gateway"
	ResourceTypeVirtualMachine     ResourceType = "virtual_machine"
	ResourceTypeAKSCluster         ResourceType = "aks_cluster"
	ResourceTypeManagedDisk        ResourceType = "managed_disk"
	ResourceTypeStorageAccount     ResourceType = "storage_account"
	ResourceTypeSQLServer          ResourceType = "sql_server"
	ResourceTypeSQLDatabase        ResourceType = "sql_database"
	ResourceTypeAzureLB            ResourceType = "azure_load_balancer"
	ResourceTypeAppGateway         ResourceType = "application_gateway"
	ResourceTypePublicIP           ResourceType = "public_ip"
	ResourceTypeKeyVault           ResourceType = "key_vault"
	ResourceTypeAppService         ResourceType = "app_service"
	ResourceTypeCosmosDB           ResourceType = "cosmosdb_account"
)
