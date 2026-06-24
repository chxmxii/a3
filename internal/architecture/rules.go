package architecture

import (
	"github.com/chxmxii/a3/internal/storage"
)

// RelationshipRule defines how to infer relationships between resources.
type RelationshipRule interface {
	Apply(resources []storage.Resource, byID map[string]*storage.Resource, byInternalID map[string]*storage.Resource) []storage.Relationship
}

// metadataLinkRule links a source resource to a target based on a metadata field.
type metadataLinkRule struct {
	sourceType       string
	targetType       string
	metadataKey      string
	relationshipType string
	lookupPrefix     string // prefix for byInternalID lookup
}

func (r *metadataLinkRule) Apply(resources []storage.Resource, byID map[string]*storage.Resource, byInternalID map[string]*storage.Resource) []storage.Relationship {
	var rels []storage.Relationship
	for _, res := range resources {
		if res.ResourceType != r.sourceType {
			continue
		}
		targetRef := getStr(res.RawMetadata, r.metadataKey)
		if targetRef == "" {
			continue
		}

		// Try to find the target resource.
		status := "resolved"
		reason := ""
		targetID := targetRef

		// Look up by internal ID first.
		if r.lookupPrefix != "" {
			key := r.lookupPrefix + ":" + targetRef
			if target, ok := byInternalID[key]; ok {
				targetID = target.ResourceID
			} else {
				status = "unresolved"
				reason = "target not found in assessment"
			}
		} else {
			if _, ok := byID[targetRef]; !ok {
				status = "unresolved"
				reason = "target not found in assessment"
			}
		}

		rels = append(rels, storage.Relationship{
			SourceID:         res.ResourceID,
			TargetID:         targetID,
			RelationshipType: r.relationshipType,
			Status:           status,
			UnresolvedReason: reason,
			TargetRegion:     res.Region,
		})
	}
	return rels
}

func awsRelationshipRules() []RelationshipRule {
	return []RelationshipRule{
		// Subnet → VPC
		&metadataLinkRule{
			sourceType:       "subnet",
			targetType:       "vpc",
			metadataKey:      "vpc_id",
			relationshipType: "belongs_to",
			lookupPrefix:     "vpc_id",
		},
		// Route Table → VPC
		&metadataLinkRule{
			sourceType:       "route_table",
			targetType:       "vpc",
			metadataKey:      "vpc_id",
			relationshipType: "belongs_to",
			lookupPrefix:     "vpc_id",
		},
		// Security Group → VPC
		&metadataLinkRule{
			sourceType:       "security_group",
			targetType:       "vpc",
			metadataKey:      "vpc_id",
			relationshipType: "belongs_to",
			lookupPrefix:     "vpc_id",
		},
		// EC2 → Subnet
		&metadataLinkRule{
			sourceType:       "ec2_instance",
			targetType:       "subnet",
			metadataKey:      "subnet_id",
			relationshipType: "deployed_in",
			lookupPrefix:     "subnet_id",
		},
		// EC2 → VPC
		&metadataLinkRule{
			sourceType:       "ec2_instance",
			targetType:       "vpc",
			metadataKey:      "vpc_id",
			relationshipType: "belongs_to",
			lookupPrefix:     "vpc_id",
		},
		// NAT Gateway → VPC
		&metadataLinkRule{
			sourceType:       "nat_gateway",
			targetType:       "vpc",
			metadataKey:      "vpc_id",
			relationshipType: "belongs_to",
			lookupPrefix:     "vpc_id",
		},
		// NAT Gateway → Subnet
		&metadataLinkRule{
			sourceType:       "nat_gateway",
			targetType:       "subnet",
			metadataKey:      "subnet_id",
			relationshipType: "deployed_in",
			lookupPrefix:     "subnet_id",
		},
		// Internet Gateway → VPC
		&metadataLinkRule{
			sourceType:       "internet_gateway",
			targetType:       "vpc",
			metadataKey:      "vpc_id",
			relationshipType: "attached_to",
			lookupPrefix:     "vpc_id",
		},
		// ALB → VPC
		&metadataLinkRule{
			sourceType:       "alb",
			targetType:       "vpc",
			metadataKey:      "vpc_id",
			relationshipType: "belongs_to",
			lookupPrefix:     "vpc_id",
		},
		// NLB → VPC
		&metadataLinkRule{
			sourceType:       "nlb",
			targetType:       "vpc",
			metadataKey:      "vpc_id",
			relationshipType: "belongs_to",
			lookupPrefix:     "vpc_id",
		},
		// EKS Node Group → Cluster (by cluster_name)
		&metadataLinkRule{
			sourceType:       "eks_node_group",
			targetType:       "eks_cluster",
			metadataKey:      "cluster_name",
			relationshipType: "belongs_to",
			lookupPrefix:     "cluster_name",
		},
		// Lambda → VPC (via vpc_id in vpc_config)
		&metadataLinkRule{
			sourceType:       "lambda_function",
			targetType:       "vpc",
			metadataKey:      "vpc_id",
			relationshipType: "deployed_in",
			lookupPrefix:     "vpc_id",
		},
		// RDS → VPC
		&metadataLinkRule{
			sourceType:       "rds_instance",
			targetType:       "vpc",
			metadataKey:      "vpc_id",
			relationshipType: "deployed_in",
			lookupPrefix:     "vpc_id",
		},
	}
}

func ociRelationshipRules() []RelationshipRule {
	return []RelationshipRule{
		// Subnet → VCN
		&metadataLinkRule{
			sourceType:       "oci_subnet",
			targetType:       "vcn",
			metadataKey:      "vcn_id",
			relationshipType: "belongs_to",
			lookupPrefix:     "vcn_id",
		},
		// Route Table → VCN
		&metadataLinkRule{
			sourceType:       "oci_route_table",
			targetType:       "vcn",
			metadataKey:      "vcn_id",
			relationshipType: "belongs_to",
			lookupPrefix:     "vcn_id",
		},
		// Security List → VCN
		&metadataLinkRule{
			sourceType:       "security_list",
			targetType:       "vcn",
			metadataKey:      "vcn_id",
			relationshipType: "belongs_to",
			lookupPrefix:     "vcn_id",
		},
		// NSG → VCN
		&metadataLinkRule{
			sourceType:       "nsg",
			targetType:       "vcn",
			metadataKey:      "vcn_id",
			relationshipType: "belongs_to",
			lookupPrefix:     "vcn_id",
		},
		// Internet Gateway → VCN
		&metadataLinkRule{
			sourceType:       "oci_internet_gateway",
			targetType:       "vcn",
			metadataKey:      "vcn_id",
			relationshipType: "attached_to",
			lookupPrefix:     "vcn_id",
		},
		// NAT Gateway → VCN
		&metadataLinkRule{
			sourceType:       "oci_nat_gateway",
			targetType:       "vcn",
			metadataKey:      "vcn_id",
			relationshipType: "attached_to",
			lookupPrefix:     "vcn_id",
		},
		// Service Gateway → VCN
		&metadataLinkRule{
			sourceType:       "service_gateway",
			targetType:       "vcn",
			metadataKey:      "vcn_id",
			relationshipType: "attached_to",
			lookupPrefix:     "vcn_id",
		},
		// OKE Cluster → VCN
		&metadataLinkRule{
			sourceType:       "oke_cluster",
			targetType:       "vcn",
			metadataKey:      "vcn_id",
			relationshipType: "deployed_in",
			lookupPrefix:     "vcn_id",
		},
		// Subnet → Route Table
		&metadataLinkRule{
			sourceType:       "oci_subnet",
			targetType:       "oci_route_table",
			metadataKey:      "route_table_id",
			relationshipType: "uses",
			lookupPrefix:     "route_table_id",
		},
	}
}

func azureRelationshipRules() []RelationshipRule {
	return []RelationshipRule{
		// Subnet → Virtual Network (Azure subnets carry the parent vnet name)
		&metadataLinkRule{
			sourceType:       "azure_subnet",
			targetType:       "vnet",
			metadataKey:      "virtual_network_name",
			relationshipType: "belongs_to",
			lookupPrefix:     "name",
		},
		// Subnet → NSG
		&metadataLinkRule{
			sourceType:       "azure_subnet",
			targetType:       "azure_nsg",
			metadataKey:      "network_security_group_id",
			relationshipType: "uses",
			lookupPrefix:     "resource_id",
		},
		// Subnet → Route Table
		&metadataLinkRule{
			sourceType:       "azure_subnet",
			targetType:       "azure_route_table",
			metadataKey:      "route_table_id",
			relationshipType: "uses",
			lookupPrefix:     "resource_id",
		},
		// Virtual Network → Resource Group
		&metadataLinkRule{
			sourceType:       "vnet",
			targetType:       "resource_group",
			metadataKey:      "resource_group",
			relationshipType: "belongs_to",
			lookupPrefix:     "name",
		},
		// Virtual Machine → Resource Group
		&metadataLinkRule{
			sourceType:       "virtual_machine",
			targetType:       "resource_group",
			metadataKey:      "resource_group",
			relationshipType: "belongs_to",
			lookupPrefix:     "name",
		},
		// Storage Account → Resource Group
		&metadataLinkRule{
			sourceType:       "storage_account",
			targetType:       "resource_group",
			metadataKey:      "resource_group",
			relationshipType: "belongs_to",
			lookupPrefix:     "name",
		},
		// AKS Cluster → Resource Group
		&metadataLinkRule{
			sourceType:       "aks_cluster",
			targetType:       "resource_group",
			metadataKey:      "resource_group",
			relationshipType: "belongs_to",
			lookupPrefix:     "name",
		},
	}
}
