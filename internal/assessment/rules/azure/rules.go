package azure

import (
	"context"
	"fmt"

	"github.com/chxmxii/a3/internal/assessment"
	"github.com/chxmxii/a3/internal/provider"
	"github.com/chxmxii/a3/internal/storage"
)

// AllRules returns all Azure assessment rules.
func AllRules() []assessment.Rule {
	return []assessment.Rule{
		&StorageAccountPublicRule{},
		&NSGOpenIngressRule{},
		&UnencryptedDiskRule{},
		&SQLPublicAccessRule{},
	}
}

const standardName = "3A Security Baseline"

// StorageAccountPublicRule flags storage accounts that allow public blob access.
type StorageAccountPublicRule struct{}

func (r *StorageAccountPublicRule) ID() string                          { return "azure-storage-public-blob" }
func (r *StorageAccountPublicRule) Standard() string                    { return standardName }
func (r *StorageAccountPublicRule) ControlID() string                   { return "SEC-012" }
func (r *StorageAccountPublicRule) Category() assessment.FindingCategory { return assessment.CategorySecurity }
func (r *StorageAccountPublicRule) AppliesTo() []provider.ResourceType {
	return []provider.ResourceType{provider.ResourceTypeStorageAccount}
}

func (r *StorageAccountPublicRule) Evaluate(_ context.Context, resource storage.Resource) ([]assessment.Finding, error) {
	meta := resource.RawMetadata

	// Steampipe azure_storage_account.allow_blob_public_access (bool). When true,
	// containers may be configured for anonymous public access.
	if allow, ok := toBool(meta["allow_blob_public_access"]); ok && allow {
		return []assessment.Finding{{
			Severity:       assessment.SeverityHigh,
			ResourceID:     resource.ResourceID,
			Description:    fmt.Sprintf("Storage account %s allows public blob access", resource.Name),
			Recommendation: "Set allow_blob_public_access to false unless anonymous access is explicitly required",
			StandardName:   r.Standard(),
			ControlID:      r.ControlID(),
			Category:       r.Category(),
		}}, nil
	}

	return nil, nil
}

// NSGOpenIngressRule flags network security groups allowing inbound traffic from
// the internet on dangerous ports.
type NSGOpenIngressRule struct{}

func (r *NSGOpenIngressRule) ID() string                          { return "azure-nsg-open-ingress" }
func (r *NSGOpenIngressRule) Standard() string                    { return standardName }
func (r *NSGOpenIngressRule) ControlID() string                   { return "SEC-013" }
func (r *NSGOpenIngressRule) Category() assessment.FindingCategory { return assessment.CategorySecurity }
func (r *NSGOpenIngressRule) AppliesTo() []provider.ResourceType {
	return []provider.ResourceType{provider.ResourceTypeAzureNSG}
}

func (r *NSGOpenIngressRule) Evaluate(_ context.Context, resource storage.Resource) ([]assessment.Finding, error) {
	meta := resource.RawMetadata
	var findings []assessment.Finding

	rules, ok := meta["security_rules"].([]any)
	if !ok {
		return nil, nil
	}

	for _, rule := range rules {
		ruleMap, ok := rule.(map[string]any)
		if !ok {
			continue
		}
		// Steampipe nests the rule body under "properties".
		props := ruleMap
		if p, ok := ruleMap["properties"].(map[string]any); ok {
			props = p
		}

		access, _ := props["access"].(string)
		direction, _ := props["direction"].(string)
		if access != "Allow" || direction != "Inbound" {
			continue
		}

		if !isOpenSource(props) {
			continue
		}

		port := portRange(props)
		findings = append(findings, assessment.Finding{
			Severity:       assessment.SeverityHigh,
			ResourceID:     resource.ResourceID,
			Description:    fmt.Sprintf("NSG %s allows inbound traffic from the internet (ports: %s)", resource.Name, port),
			Recommendation: "Restrict inbound security rules to specific source address prefixes",
			StandardName:   r.Standard(),
			ControlID:      r.ControlID(),
			Category:       r.Category(),
		})
	}

	return findings, nil
}

// UnencryptedDiskRule flags managed disks that rely solely on platform-managed
// keys (no customer-managed key).
type UnencryptedDiskRule struct{}

func (r *UnencryptedDiskRule) ID() string                          { return "azure-disk-no-cmk" }
func (r *UnencryptedDiskRule) Standard() string                    { return standardName }
func (r *UnencryptedDiskRule) ControlID() string                   { return "SEC-014" }
func (r *UnencryptedDiskRule) Category() assessment.FindingCategory { return assessment.CategorySecurity }
func (r *UnencryptedDiskRule) AppliesTo() []provider.ResourceType {
	return []provider.ResourceType{provider.ResourceTypeManagedDisk}
}

func (r *UnencryptedDiskRule) Evaluate(_ context.Context, resource storage.Resource) ([]assessment.Finding, error) {
	meta := resource.RawMetadata

	// Steampipe azure_compute_disk.encryption_type. Platform-managed-only keys
	// are the default; customer-managed keys ("...CustomerKey") are stronger.
	encType, _ := meta["encryption_type"].(string)
	if encType == "EncryptionAtRestWithPlatformKey" || encType == "" {
		return []assessment.Finding{{
			Severity:       assessment.SeverityLow,
			ResourceID:     resource.ResourceID,
			Description:    fmt.Sprintf("Managed disk %s uses platform-managed encryption only (no customer-managed key)", resource.Name),
			Recommendation: "Consider customer-managed keys (Key Vault) for disks holding sensitive data",
			StandardName:   r.Standard(),
			ControlID:      r.ControlID(),
			Category:       r.Category(),
		}}, nil
	}

	return nil, nil
}

// SQLPublicAccessRule flags SQL servers reachable from public networks.
type SQLPublicAccessRule struct{}

func (r *SQLPublicAccessRule) ID() string                          { return "azure-sql-public-access" }
func (r *SQLPublicAccessRule) Standard() string                    { return standardName }
func (r *SQLPublicAccessRule) ControlID() string                   { return "SEC-015" }
func (r *SQLPublicAccessRule) Category() assessment.FindingCategory { return assessment.CategorySecurity }
func (r *SQLPublicAccessRule) AppliesTo() []provider.ResourceType {
	return []provider.ResourceType{provider.ResourceTypeSQLServer}
}

func (r *SQLPublicAccessRule) Evaluate(_ context.Context, resource storage.Resource) ([]assessment.Finding, error) {
	meta := resource.RawMetadata

	// Steampipe azure_sql_server.public_network_access ("Enabled"/"Disabled").
	if access, _ := meta["public_network_access"].(string); access == "Enabled" {
		return []assessment.Finding{{
			Severity:       assessment.SeverityMedium,
			ResourceID:     resource.ResourceID,
			Description:    fmt.Sprintf("SQL server %s has public network access enabled", resource.Name),
			Recommendation: "Disable public network access and use private endpoints or firewall rules",
			StandardName:   r.Standard(),
			ControlID:      r.ControlID(),
			Category:       r.Category(),
		}}, nil
	}

	return nil, nil
}

// isOpenSource reports whether an NSG rule's source matches the internet.
func isOpenSource(props map[string]any) bool {
	candidates := []string{}
	if s, ok := props["source_address_prefix"].(string); ok {
		candidates = append(candidates, s)
	}
	if list, ok := props["source_address_prefixes"].([]any); ok {
		for _, v := range list {
			if s, ok := v.(string); ok {
				candidates = append(candidates, s)
			}
		}
	}
	for _, s := range candidates {
		switch s {
		case "*", "0.0.0.0/0", "Internet", "::/0":
			return true
		}
	}
	return false
}

// portRange returns a human-readable destination port for an NSG rule.
func portRange(props map[string]any) string {
	if p, ok := props["destination_port_range"].(string); ok && p != "" {
		return p
	}
	if list, ok := props["destination_port_ranges"].([]any); ok && len(list) > 0 {
		parts := make([]string, 0, len(list))
		for _, v := range list {
			if s, ok := v.(string); ok {
				parts = append(parts, s)
			}
		}
		if len(parts) > 0 {
			return fmt.Sprintf("%v", parts)
		}
	}
	return "any"
}

// toBool coerces common representations of a boolean from steampipe metadata.
func toBool(v any) (bool, bool) {
	switch val := v.(type) {
	case bool:
		return val, true
	case *bool:
		if val != nil {
			return *val, true
		}
	}
	return false, false
}
