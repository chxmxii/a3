package oci

import (
	"context"
	"fmt"

	"github.com/chxmxii/a3/internal/assessment"
	"github.com/chxmxii/a3/internal/provider"
	"github.com/chxmxii/a3/internal/storage"
)

// AllRules returns all OCI assessment rules.
func AllRules() []assessment.Rule {
	return []assessment.Rule{
		publicBucketRule(),
		&NSGOpenIngressRule{},
		unencryptedVolumeRule(),
		dbPublicAccessRule(),
	}
}

// publicBucketRule checks for OCI Object Storage buckets with public access.
func publicBucketRule() *simpleRule {
	return &simpleRule{
		id:             "oci-public-bucket",
		standard:       "3A Security Baseline",
		controlID:      "SEC-008",
		category:       assessment.CategorySecurity,
		severity:       assessment.SeverityHigh,
		appliesTo:      []provider.ResourceType{provider.ResourceTypeObjectStorage},
		recommendation: "Set public_access_type to NoPublicAccess unless public access is explicitly required",
		describe: func(resource storage.Resource) string {
			publicAccess, _ := resource.RawMetadata["public_access_type"].(string)
			return fmt.Sprintf("Object storage bucket %s has public access type: %s", resource.Name, publicAccess)
		},
		violated: func(meta map[string]any) bool {
			// Steampipe: public_access_type field — "NoPublicAccess" is the secure value.
			publicAccess, ok := meta["public_access_type"].(string)
			return ok && publicAccess != "NoPublicAccess" && publicAccess != ""
		},
	}
}

// NSGOpenIngressRule checks for NSGs allowing all ingress traffic.
type NSGOpenIngressRule struct{}

func (r *NSGOpenIngressRule) ID() string        { return "oci-nsg-open-ingress" }
func (r *NSGOpenIngressRule) Standard() string  { return "3A Security Baseline" }
func (r *NSGOpenIngressRule) ControlID() string { return "SEC-009" }
func (r *NSGOpenIngressRule) Category() assessment.FindingCategory {
	return assessment.CategorySecurity
}
func (r *NSGOpenIngressRule) AppliesTo() []provider.ResourceType {
	return []provider.ResourceType{provider.ResourceTypeNSG}
}

func (r *NSGOpenIngressRule) Evaluate(_ context.Context, resource storage.Resource) ([]assessment.Finding, error) {
	meta := resource.RawMetadata
	var findings []assessment.Finding

	// Steampipe: rules column contains security rules.
	rules, ok := meta["rules"].([]any)
	if !ok {
		return nil, nil
	}

	for _, rule := range rules {
		ruleMap, ok := rule.(map[string]any)
		if !ok {
			continue
		}

		direction, _ := ruleMap["direction"].(string)
		if direction != "INGRESS" {
			continue
		}

		source, _ := ruleMap["source"].(string)
		if source == "0.0.0.0/0" {
			protocol, _ := ruleMap["protocol"].(string)
			if protocol == "all" || protocol == "6" { // all or TCP
				findings = append(findings, newFinding(r, assessment.SeverityHigh, resource,
					fmt.Sprintf("NSG %s allows ingress from 0.0.0.0/0 (protocol: %s)", resource.Name, protocol),
					"Restrict ingress rules to specific source CIDRs"))
			}
		}
	}

	return findings, nil
}

// unencryptedVolumeRule checks for OCI block volumes without encryption.
func unencryptedVolumeRule() *simpleRule {
	return &simpleRule{
		id:             "oci-volume-unencrypted",
		standard:       "3A Security Baseline",
		controlID:      "SEC-010",
		category:       assessment.CategorySecurity,
		severity:       assessment.SeverityLow,
		appliesTo:      []provider.ResourceType{provider.ResourceTypeBlockVolume},
		recommendation: "Consider using a customer-managed encryption key (KMS) for sensitive data",
		describe: func(resource storage.Resource) string {
			return fmt.Sprintf("Block volume %s uses Oracle-managed encryption (no customer-managed key)", resource.Name)
		},
		violated: func(meta map[string]any) bool {
			// OCI volumes are encrypted by default with Oracle-managed keys,
			// but no customer-managed key indicates lower security posture.
			kmsKeyID, _ := meta["kms_key_id"].(string)
			return kmsKeyID == ""
		},
	}
}

// dbPublicAccessRule checks for publicly accessible OCI database systems.
func dbPublicAccessRule() *simpleRule {
	return &simpleRule{
		id:             "oci-db-public",
		standard:       "3A Security Baseline",
		controlID:      "SEC-011",
		category:       assessment.CategorySecurity,
		severity:       assessment.SeverityMedium,
		appliesTo:      []provider.ResourceType{provider.ResourceTypeOCIDB},
		recommendation: "Apply NSG rules to restrict database access to authorized sources only",
		describe: func(resource storage.Resource) string {
			return fmt.Sprintf("Database system %s has no NSG protection", resource.Name)
		},
		violated: func(meta map[string]any) bool {
			// Check if the subnet is public (no prohibition on public IPs).
			// This is a heuristic — DB systems in public subnets are publicly accessible.
			hostname, ok := meta["hostname"].(string)
			if !ok || hostname == "" {
				return false
			}
			// If there's a hostname and the subnet doesn't block public IPs.
			nsgIDs, ok := meta["nsg_ids"].([]any)
			return ok && len(nsgIDs) == 0
		},
	}
}
