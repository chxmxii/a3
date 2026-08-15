package assessment

// Standard represents a compliance standard or framework.
type Standard struct {
	Name        string
	Version     string
	Description string
	Controls    []Control
}

// Control represents a specific control within a standard.
type Control struct {
	ID          string
	Name        string
	Description string
	Category    FindingCategory
}

// BuiltInStandards returns the compliance standards supported by 3A.
// Every (Standard, ControlID) pair referenced by a rule in
// internal/assessment/rules must have a matching control here; the
// checklist engine derives its checks from this catalog.
func BuiltInStandards() []Standard {
	return []Standard{
		{
			Name:        "3A Security Baseline",
			Version:     "1.0",
			Description: "3A built-in security assessment baseline",
			Controls: []Control{
				{ID: "SEC-001", Name: "S3 Public Access", Description: "S3 buckets should not allow public access", Category: CategorySecurity},
				{ID: "SEC-002", Name: "Security Group Open Access", Description: "Security groups should not allow unrestricted inbound access on dangerous ports", Category: CategorySecurity},
				{ID: "SEC-003", Name: "EBS Encryption", Description: "EBS volumes should be encrypted", Category: CategorySecurity},
				{ID: "SEC-004", Name: "RDS Public Access", Description: "RDS instances should not be publicly accessible", Category: CategorySecurity},
				{ID: "SEC-005", Name: "IAM MFA", Description: "IAM users should have MFA enabled", Category: CategorySecurity},
				{ID: "SEC-006", Name: "EKS Public Endpoint", Description: "EKS clusters should not have public API endpoints", Category: CategorySecurity},
				{ID: "SEC-007", Name: "S3 Encryption", Description: "S3 buckets should have default encryption enabled", Category: CategorySecurity},
				{ID: "SEC-008", Name: "OCI Public Bucket", Description: "Object storage buckets should not be public", Category: CategorySecurity},
				{ID: "SEC-009", Name: "OCI NSG Open Ingress", Description: "NSGs should not allow unrestricted ingress", Category: CategorySecurity},
				{ID: "SEC-010", Name: "OCI Volume Encryption", Description: "Block volumes should be encrypted", Category: CategorySecurity},
				{ID: "SEC-011", Name: "OCI DB Public Access", Description: "Database systems should not be publicly accessible", Category: CategorySecurity},
			},
		},
		{
			Name:        "CIS AWS Foundations",
			Version:     "1.0",
			Description: "CIS AWS Foundations benchmark controls checked by 3A",
			Controls: []Control{
				{ID: "SEC-009", Name: "EC2 Public IP", Description: "EC2 instances should not have public IP addresses assigned directly", Category: CategorySecurity},
				{ID: "SEC-010", Name: "EC2 IMDSv2 Enforcement", Description: "EC2 instances should enforce IMDSv2 for instance metadata access", Category: CategorySecurity},
				{ID: "SEC-011", Name: "RDS Storage Encryption", Description: "RDS instances should have storage encryption at rest enabled", Category: CategorySecurity},
				{ID: "SEC-012", Name: "VPC Flow Logs", Description: "VPCs should have flow logs enabled", Category: CategorySecurity},
				{ID: "SEC-013", Name: "Default Security Group Rules", Description: "Default security groups should not have inbound rules configured", Category: CategorySecurity},
				{ID: "SEC-016", Name: "Load Balancer Access Logs", Description: "Load balancers should have access logging enabled", Category: CategorySecurity},
			},
		},
		{
			Name:        "AWS Well-Architected",
			Version:     "1.0",
			Description: "AWS Well-Architected framework controls checked by 3A",
			Controls: []Control{
				{ID: "SEC-008", Name: "Lambda VPC Attachment", Description: "Lambda functions accessing private resources should be attached to a VPC", Category: CategorySecurity},
				{ID: "SEC-014", Name: "Subnet Public IP Auto-Assign", Description: "Subnets should not auto-assign public IPs to instances", Category: CategorySecurity},
				{ID: "SEC-015", Name: "ALB WAF Protection", Description: "Internet-facing ALBs should have a WAF Web ACL associated", Category: CategorySecurity},
				{ID: "COST-001", Name: "Lambda Memory Sizing", Description: "Lambda functions should not be over-provisioned with memory", Category: CategoryCostOptimization},
				{ID: "COST-002", Name: "Stopped EC2 Instances", Description: "Stopped EC2 instances should be terminated or their EBS volumes reclaimed", Category: CategoryCostOptimization},
				{ID: "REL-001", Name: "Lambda Dead Letter Queue", Description: "Lambda functions should have a dead letter queue configured", Category: CategoryReliability},
				{ID: "REL-002", Name: "RDS Multi-AZ", Description: "RDS instances should use Multi-AZ deployment for automatic failover", Category: CategoryReliability},
				{ID: "REL-003", Name: "RDS Automated Backups", Description: "RDS instances should have automated backups enabled", Category: CategoryReliability},
				{ID: "REL-004", Name: "Load Balancer Deletion Protection", Description: "Load balancers should have deletion protection enabled", Category: CategoryReliability},
				{ID: "OPS-001", Name: "Lambda Runtime Version", Description: "Lambda functions should use supported runtime versions", Category: CategoryOperationalExcellence},
				{ID: "OPS-002", Name: "RDS Auto Minor Upgrade", Description: "RDS instances should have auto minor version upgrade enabled", Category: CategoryOperationalExcellence},
			},
		},
	}
}
