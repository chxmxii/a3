package aws

import (
	"fmt"

	"github.com/chxmxii/a3/internal/assessment"
	"github.com/chxmxii/a3/internal/metautil"
	"github.com/chxmxii/a3/internal/provider"
	"github.com/chxmxii/a3/internal/storage"
)

// ec2PublicIPRule checks for EC2 instances with public IPs.
func ec2PublicIPRule() *simpleRule {
	return &simpleRule{
		id:             "aws-ec2-public-ip",
		standard:       "CIS AWS Foundations",
		controlID:      "SEC-009",
		category:       assessment.CategorySecurity,
		severity:       assessment.SeverityMedium,
		appliesTo:      []provider.ResourceType{provider.ResourceTypeEC2Instance},
		recommendation: "Use a load balancer or NAT gateway instead of direct public IP assignment",
		describe: func(resource storage.Resource) string {
			publicIP := metautil.GetString(resource.RawMetadata, "public_ip_address")
			return fmt.Sprintf("EC2 instance %s has a public IP (%s)", resource.Name, publicIP)
		},
		violated: func(meta map[string]any) bool {
			return metautil.GetString(meta, "public_ip_address") != ""
		},
	}
}

// ec2StoppedInstanceRule checks for instances that have been stopped (wasting money on EBS).
func ec2StoppedInstanceRule() *simpleRule {
	return &simpleRule{
		id:             "aws-ec2-stopped",
		standard:       "AWS Well-Architected",
		controlID:      "COST-002",
		category:       assessment.CategoryCostOptimization,
		severity:       assessment.SeverityLow,
		appliesTo:      []provider.ResourceType{provider.ResourceTypeEC2Instance},
		recommendation: "Terminate the instance if no longer needed, or snapshot and delete EBS volumes",
		describe: func(resource storage.Resource) string {
			return fmt.Sprintf("EC2 instance %s is stopped (EBS volumes still incur charges)", resource.Name)
		},
		violated: func(meta map[string]any) bool {
			return metautil.GetString(meta, "instance_state", "state") == "stopped"
		},
	}
}

// ec2NoIMDSv2Rule checks if instance metadata service v2 is not enforced.
func ec2NoIMDSv2Rule() *simpleRule {
	return &simpleRule{
		id:             "aws-ec2-no-imdsv2",
		standard:       "CIS AWS Foundations",
		controlID:      "SEC-010",
		category:       assessment.CategorySecurity,
		severity:       assessment.SeverityMedium,
		appliesTo:      []provider.ResourceType{provider.ResourceTypeEC2Instance},
		recommendation: "Set HttpTokens to 'required' to enforce IMDSv2 and prevent SSRF-based credential theft",
		describe: func(resource storage.Resource) string {
			return fmt.Sprintf("EC2 instance %s does not enforce IMDSv2", resource.Name)
		},
		violated: func(meta map[string]any) bool {
			httpTokens := metautil.GetString(meta, "metadata_options_http_tokens")
			if httpTokens == "" {
				if opts, ok := meta["metadata_options"].(map[string]any); ok {
					httpTokens, _ = opts["http_tokens"].(string)
				}
			}
			// "required" means IMDSv2 is enforced. "optional" means v1 is still allowed.
			return httpTokens == "optional" || httpTokens == ""
		},
	}
}
