package aws

import (
	"fmt"

	"github.com/chxmxii/a3/internal/assessment"
	"github.com/chxmxii/a3/internal/provider"
	"github.com/chxmxii/a3/internal/storage"
)

// rdsNoMultiAZRule checks for RDS instances without Multi-AZ.
func rdsNoMultiAZRule() *simpleRule {
	return &simpleRule{
		id:             "aws-rds-no-multi-az",
		standard:       "AWS Well-Architected",
		controlID:      "REL-002",
		category:       assessment.CategoryReliability,
		severity:       assessment.SeverityMedium,
		appliesTo:      []provider.ResourceType{provider.ResourceTypeRDS},
		recommendation: "Enable Multi-AZ for production databases to ensure automatic failover",
		describe: func(resource storage.Resource) string {
			return fmt.Sprintf("RDS instance %s is not Multi-AZ", resource.Name)
		},
		violated: func(meta map[string]any) bool {
			multiAZ, ok := meta["multi_az"].(bool)
			return ok && !multiAZ
		},
	}
}

// rdsNoEncryptionRule checks for unencrypted RDS instances.
func rdsNoEncryptionRule() *simpleRule {
	return &simpleRule{
		id:             "aws-rds-no-encryption",
		standard:       "CIS AWS Foundations",
		controlID:      "SEC-011",
		category:       assessment.CategorySecurity,
		severity:       assessment.SeverityHigh,
		appliesTo:      []provider.ResourceType{provider.ResourceTypeRDS},
		recommendation: "Enable encryption at rest. Requires creating a new encrypted instance and migrating data.",
		describe: func(resource storage.Resource) string {
			return fmt.Sprintf("RDS instance %s storage is not encrypted", resource.Name)
		},
		violated: func(meta map[string]any) bool {
			encrypted, ok := meta["storage_encrypted"].(bool)
			return ok && !encrypted
		},
	}
}

// rdsNoBackupRule checks for RDS instances with no automated backups.
func rdsNoBackupRule() *simpleRule {
	return &simpleRule{
		id:             "aws-rds-no-backup",
		standard:       "AWS Well-Architected",
		controlID:      "REL-003",
		category:       assessment.CategoryReliability,
		severity:       assessment.SeverityHigh,
		appliesTo:      []provider.ResourceType{provider.ResourceTypeRDS},
		recommendation: "Enable automated backups with at least 7 days retention",
		describe: func(resource storage.Resource) string {
			return fmt.Sprintf("RDS instance %s has automated backups disabled (retention = 0)", resource.Name)
		},
		violated: func(meta map[string]any) bool {
			retention, ok := meta["backup_retention_period"].(float64)
			return ok && retention == 0
		},
	}
}

// rdsAutoMinorUpgradeRule checks if auto minor version upgrade is disabled.
func rdsAutoMinorUpgradeRule() *simpleRule {
	return &simpleRule{
		id:             "aws-rds-no-auto-upgrade",
		standard:       "AWS Well-Architected",
		controlID:      "OPS-002",
		category:       assessment.CategoryOperationalExcellence,
		severity:       assessment.SeverityLow,
		appliesTo:      []provider.ResourceType{provider.ResourceTypeRDS},
		recommendation: "Enable auto minor version upgrades to receive security patches automatically",
		describe: func(resource storage.Resource) string {
			return fmt.Sprintf("RDS instance %s has auto minor version upgrade disabled", resource.Name)
		},
		violated: func(meta map[string]any) bool {
			autoUpgrade, ok := meta["auto_minor_version_upgrade"].(bool)
			return ok && !autoUpgrade
		},
	}
}
