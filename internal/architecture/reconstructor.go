package architecture

import (
	"log"

	"github.com/chxmxii/a3/internal/storage"
)

// Reconstructor analyzes discovered resources and infers relationships between them.
type Reconstructor struct {
	store *storage.Store
	rules []RelationshipRule
}

// NewReconstructor creates a new architecture reconstructor with provider-specific rules.
func NewReconstructor(store *storage.Store, providerType string) *Reconstructor {
	var rules []RelationshipRule
	switch providerType {
	case "aws":
		rules = awsRelationshipRules()
	case "oci":
		rules = ociRelationshipRules()
	case "azure":
		rules = azureRelationshipRules()
	}
	return &Reconstructor{
		store: store,
		rules: rules,
	}
}

// Reconstruct loads all resources for the assessment and applies relationship rules.
func (r *Reconstructor) Reconstruct(assessmentID string) error {
	resources, err := r.store.GetResourcesByAssessment(assessmentID)
	if err != nil {
		return err
	}

	// Build a lookup map by resource ID for resolving references.
	byID := make(map[string]*storage.Resource, len(resources))
	for i := range resources {
		byID[resources[i].ResourceID] = &resources[i]
	}

	// Also build lookup by internal identifiers (vpc_id, subnet_id, etc.)
	byInternalID := make(map[string]*storage.Resource)
	for i := range resources {
		res := &resources[i]
		// Index by common ID fields in metadata.
		for _, key := range []string{"vpc_id", "vpcId", "subnet_id", "subnetId",
			"instance_id", "instanceId", "cluster_name", "group_id", "groupId",
			"route_table_id", "routeTableId", "internet_gateway_id", "nat_gateway_id",
			"transit_gateway_id", "id", "vcn_id", "name"} {
			if val := getStr(res.RawMetadata, key); val != "" {
				// Only store if not already taken (first match wins).
				mapKey := key + ":" + val
				if _, exists := byInternalID[mapKey]; !exists {
					byInternalID[mapKey] = res
				}
			}
		}
		// Index by the resource_id directly for lookup by ARN/OCID.
		byInternalID["resource_id:"+res.ResourceID] = res
	}

	// Apply each rule against the resources.
	for _, rule := range r.rules {
		rels := rule.Apply(resources, byID, byInternalID)
		for _, rel := range rels {
			rel.AssessmentID = assessmentID
			if err := r.store.InsertRelationship(&rel); err != nil {
				log.Printf("[architecture] failed to insert relationship %s -> %s: %v",
					rel.SourceID, rel.TargetID, err)
			}
		}
	}

	return nil
}

// getStr extracts a string from a map[string]any safely.
func getStr(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	default:
		return ""
	}
}
