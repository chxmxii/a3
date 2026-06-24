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
				{ID: "SEC-012", Name: "Azure Storage Public Blob", Description: "Storage accounts should not allow public blob access", Category: CategorySecurity},
				{ID: "SEC-013", Name: "Azure NSG Open Ingress", Description: "Network security groups should not allow unrestricted inbound access", Category: CategorySecurity},
				{ID: "SEC-014", Name: "Azure Disk Encryption", Description: "Managed disks should use customer-managed encryption keys for sensitive data", Category: CategorySecurity},
				{ID: "SEC-015", Name: "Azure SQL Public Access", Description: "SQL servers should not be publicly accessible", Category: CategorySecurity},
			},
		},
	}
}
