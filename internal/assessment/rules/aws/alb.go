package aws

import (
	"fmt"

	"github.com/chxmxii/a3/internal/assessment"
	"github.com/chxmxii/a3/internal/metautil"
	"github.com/chxmxii/a3/internal/provider"
	"github.com/chxmxii/a3/internal/storage"
)

// albNoHTTPSRule checks for internet-facing ALBs without a WAF associated.
func albNoHTTPSRule() *simpleRule {
	return &simpleRule{
		id:             "aws-alb-no-https",
		standard:       "AWS Well-Architected",
		controlID:      "SEC-015",
		category:       assessment.CategorySecurity,
		severity:       assessment.SeverityMedium,
		appliesTo:      []provider.ResourceType{provider.ResourceTypeALB},
		recommendation: "Associate a WAF Web ACL to protect against common web attacks",
		describe: func(resource storage.Resource) string {
			return fmt.Sprintf("Internet-facing ALB %s has no WAF (Web ACL) associated", resource.Name)
		},
		violated: func(meta map[string]any) bool {
			// If internet-facing, flag if there is no WAF association.
			return metautil.GetString(meta, "scheme") == "internet-facing" &&
				metautil.GetString(meta, "web_acl_arn") == ""
		},
	}
}

// albDeletionProtectionRule checks if deletion protection is enabled.
func albDeletionProtectionRule() *simpleRule {
	return &simpleRule{
		id:             "aws-alb-no-deletion-protection",
		standard:       "AWS Well-Architected",
		controlID:      "REL-004",
		category:       assessment.CategoryReliability,
		severity:       assessment.SeverityLow,
		appliesTo:      []provider.ResourceType{provider.ResourceTypeALB, provider.ResourceTypeNLB},
		recommendation: "Enable deletion protection to prevent accidental deletion of production load balancers",
		describe: func(resource storage.Resource) string {
			return fmt.Sprintf("Load balancer %s does not have deletion protection enabled", resource.Name)
		},
		violated: func(meta map[string]any) bool {
			deletionProtection, ok := meta["deletion_protection_enabled"].(bool)
			return ok && !deletionProtection
		},
	}
}

// albAccessLogsRule checks if access logging is enabled.
func albAccessLogsRule() *simpleRule {
	return &simpleRule{
		id:             "aws-alb-no-access-logs",
		standard:       "CIS AWS Foundations",
		controlID:      "SEC-016",
		category:       assessment.CategorySecurity,
		severity:       assessment.SeverityLow,
		appliesTo:      []provider.ResourceType{provider.ResourceTypeALB, provider.ResourceTypeNLB},
		recommendation: "Enable access logs to S3 for audit and troubleshooting",
		describe: func(resource storage.Resource) string {
			return fmt.Sprintf("Load balancer %s does not have access logging enabled", resource.Name)
		},
		violated: func(meta map[string]any) bool {
			logsEnabled, ok := meta["access_logs_enabled"].(bool)
			return ok && !logsEnabled
		},
	}
}
