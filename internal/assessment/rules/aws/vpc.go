package aws

import (
	"fmt"

	"github.com/chxmxii/a3/internal/assessment"
	"github.com/chxmxii/a3/internal/metautil"
	"github.com/chxmxii/a3/internal/provider"
	"github.com/chxmxii/a3/internal/storage"
)

// vpcFlowLogsRule checks if VPC flow logs are enabled.
func vpcFlowLogsRule() *simpleRule {
	return &simpleRule{
		id:             "aws-vpc-no-flow-logs",
		standard:       "CIS AWS Foundations",
		controlID:      "SEC-012",
		category:       assessment.CategorySecurity,
		severity:       assessment.SeverityMedium,
		appliesTo:      []provider.ResourceType{provider.ResourceTypeVPC},
		recommendation: "Enable VPC Flow Logs to capture network traffic for security analysis",
		describe: func(resource storage.Resource) string {
			return fmt.Sprintf("VPC %s does not have flow logs enabled", resource.Name)
		},
		violated: func(meta map[string]any) bool {
			// Steampipe: flow_logs is an array. Empty or nil means no flow logs.
			flowLogs, ok := meta["flow_logs"].([]any)
			return !ok || len(flowLogs) == 0
		},
	}
}

// vpcDefaultSGRule checks if the default security group allows traffic.
func vpcDefaultSGRule() *simpleRule {
	return &simpleRule{
		id:             "aws-vpc-default-sg",
		standard:       "CIS AWS Foundations",
		controlID:      "SEC-013",
		category:       assessment.CategorySecurity,
		severity:       assessment.SeverityMedium,
		appliesTo:      []provider.ResourceType{provider.ResourceTypeSecurityGroup},
		recommendation: "Remove all inbound/outbound rules from the default security group. Use custom SGs instead.",
		describe: func(storage.Resource) string {
			return "Default security group in VPC has inbound rules configured"
		},
		violated: func(meta map[string]any) bool {
			if metautil.GetString(meta, "group_name") != "default" {
				return false
			}
			// Check if default SG has any ingress rules.
			ipPerms, _ := meta["ip_permissions"].([]any)
			return len(ipPerms) > 0
		},
	}
}

// subnetPublicIPAutoAssignRule checks subnets that auto-assign public IPs.
func subnetPublicIPAutoAssignRule() *simpleRule {
	return &simpleRule{
		id:             "aws-subnet-public-ip",
		standard:       "AWS Well-Architected",
		controlID:      "SEC-014",
		category:       assessment.CategorySecurity,
		severity:       assessment.SeverityLow,
		appliesTo:      []provider.ResourceType{provider.ResourceTypeSubnet},
		recommendation: "Disable auto-assign public IP unless the subnet is intentionally public-facing",
		describe: func(resource storage.Resource) string {
			return fmt.Sprintf("Subnet %s auto-assigns public IPs to instances", resource.Name)
		},
		violated: func(meta map[string]any) bool {
			autoAssign, ok := meta["map_public_ip_on_launch"].(bool)
			return ok && autoAssign
		},
	}
}
