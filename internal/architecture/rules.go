package architecture

import (
	"github.com/chxmxii/a3/internal/metautil"
	"github.com/chxmxii/a3/internal/storage"
)

// RelationshipRule defines how to infer relationships between resources.
type RelationshipRule interface {
	Apply(resources []storage.Resource, byID map[string]*storage.Resource, byInternalID map[string]*storage.Resource) []storage.Relationship
}

// metadataLinkRule links a source resource to a target based on a metadata field.
type metadataLinkRule struct {
	sourceType       string
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
		targetRef := metautil.GetString(res.RawMetadata, r.metadataKey)
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
			metadataKey:      "vpc_id",
			relationshipType: "belongs_to",
			lookupPrefix:     "vpc_id",
		},
		// Route Table → VPC
		&metadataLinkRule{
			sourceType:       "route_table",
			metadataKey:      "vpc_id",
			relationshipType: "belongs_to",
			lookupPrefix:     "vpc_id",
		},
		// Security Group → VPC
		&metadataLinkRule{
			sourceType:       "security_group",
			metadataKey:      "vpc_id",
			relationshipType: "belongs_to",
			lookupPrefix:     "vpc_id",
		},
		// EC2 → Subnet
		&metadataLinkRule{
			sourceType:       "ec2_instance",
			metadataKey:      "subnet_id",
			relationshipType: "deployed_in",
			lookupPrefix:     "subnet_id",
		},
		// EC2 → VPC
		&metadataLinkRule{
			sourceType:       "ec2_instance",
			metadataKey:      "vpc_id",
			relationshipType: "belongs_to",
			lookupPrefix:     "vpc_id",
		},
		// NAT Gateway → VPC
		&metadataLinkRule{
			sourceType:       "nat_gateway",
			metadataKey:      "vpc_id",
			relationshipType: "belongs_to",
			lookupPrefix:     "vpc_id",
		},
		// NAT Gateway → Subnet
		&metadataLinkRule{
			sourceType:       "nat_gateway",
			metadataKey:      "subnet_id",
			relationshipType: "deployed_in",
			lookupPrefix:     "subnet_id",
		},
		// Internet Gateway → VPC
		&metadataLinkRule{
			sourceType:       "internet_gateway",
			metadataKey:      "vpc_id",
			relationshipType: "attached_to",
			lookupPrefix:     "vpc_id",
		},
		// ALB → VPC
		&metadataLinkRule{
			sourceType:       "alb",
			metadataKey:      "vpc_id",
			relationshipType: "belongs_to",
			lookupPrefix:     "vpc_id",
		},
		// NLB → VPC
		&metadataLinkRule{
			sourceType:       "nlb",
			metadataKey:      "vpc_id",
			relationshipType: "belongs_to",
			lookupPrefix:     "vpc_id",
		},
		// EKS Node Group → Cluster (by cluster_name)
		&metadataLinkRule{
			sourceType:       "eks_node_group",
			metadataKey:      "cluster_name",
			relationshipType: "belongs_to",
			lookupPrefix:     "cluster_name",
		},
		// Lambda → VPC (via vpc_id in vpc_config)
		&metadataLinkRule{
			sourceType:       "lambda_function",
			metadataKey:      "vpc_id",
			relationshipType: "deployed_in",
			lookupPrefix:     "vpc_id",
		},
		// RDS → VPC
		&metadataLinkRule{
			sourceType:       "rds_instance",
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
			metadataKey:      "vcn_id",
			relationshipType: "belongs_to",
			lookupPrefix:     "vcn_id",
		},
		// Route Table → VCN
		&metadataLinkRule{
			sourceType:       "oci_route_table",
			metadataKey:      "vcn_id",
			relationshipType: "belongs_to",
			lookupPrefix:     "vcn_id",
		},
		// Security List → VCN
		&metadataLinkRule{
			sourceType:       "security_list",
			metadataKey:      "vcn_id",
			relationshipType: "belongs_to",
			lookupPrefix:     "vcn_id",
		},
		// NSG → VCN
		&metadataLinkRule{
			sourceType:       "nsg",
			metadataKey:      "vcn_id",
			relationshipType: "belongs_to",
			lookupPrefix:     "vcn_id",
		},
		// Internet Gateway → VCN
		&metadataLinkRule{
			sourceType:       "oci_internet_gateway",
			metadataKey:      "vcn_id",
			relationshipType: "attached_to",
			lookupPrefix:     "vcn_id",
		},
		// NAT Gateway → VCN
		&metadataLinkRule{
			sourceType:       "oci_nat_gateway",
			metadataKey:      "vcn_id",
			relationshipType: "attached_to",
			lookupPrefix:     "vcn_id",
		},
		// Service Gateway → VCN
		&metadataLinkRule{
			sourceType:       "service_gateway",
			metadataKey:      "vcn_id",
			relationshipType: "attached_to",
			lookupPrefix:     "vcn_id",
		},
		// OKE Cluster → VCN
		&metadataLinkRule{
			sourceType:       "oke_cluster",
			metadataKey:      "vcn_id",
			relationshipType: "deployed_in",
			lookupPrefix:     "vcn_id",
		},
		// Subnet → Route Table
		&metadataLinkRule{
			sourceType:       "oci_subnet",
			metadataKey:      "route_table_id",
			relationshipType: "uses",
			lookupPrefix:     "route_table_id",
		},
	}
}
