package aws

import (
	"fmt"

	"github.com/chxmxii/a3/internal/assessment"
	"github.com/chxmxii/a3/internal/metautil"
	"github.com/chxmxii/a3/internal/provider"
	"github.com/chxmxii/a3/internal/storage"
)

// lambdaVPCID returns the VPC ID a Lambda function is attached to, or "".
func lambdaVPCID(meta map[string]any) string {
	vpcID := metautil.GetString(meta, "vpc_id")
	if vpcID == "" {
		// Check nested vpc_config.
		if vpcCfg, ok := meta["vpc_config"].(map[string]any); ok {
			vpcID, _ = vpcCfg["vpc_id"].(string)
		}
	}
	return vpcID
}

// lambdaNoVPCRule checks for Lambda functions not attached to a VPC.
func lambdaNoVPCRule() *simpleRule {
	return &simpleRule{
		id:             "aws-lambda-no-vpc",
		standard:       "AWS Well-Architected",
		controlID:      "SEC-008",
		category:       assessment.CategorySecurity,
		severity:       assessment.SeverityLow,
		appliesTo:      []provider.ResourceType{provider.ResourceTypeLambda},
		recommendation: "Attach the function to a VPC if it accesses private resources",
		describe: func(resource storage.Resource) string {
			return fmt.Sprintf("Lambda function %s is not attached to a VPC", resource.Name)
		},
		violated: func(meta map[string]any) bool {
			return lambdaVPCID(meta) == ""
		},
	}
}

// lambdaOldRuntimeRule checks for Lambda functions using deprecated runtimes.
func lambdaOldRuntimeRule() *simpleRule {
	deprecated := map[string]bool{
		"python2.7":     true,
		"python3.6":     true,
		"python3.7":     true,
		"nodejs10.x":    true,
		"nodejs12.x":    true,
		"nodejs14.x":    true,
		"dotnetcore2.1": true,
		"dotnetcore3.1": true,
		"ruby2.5":       true,
		"ruby2.7":       true,
		"java8":         true,
		"go1.x":         true,
	}
	return &simpleRule{
		id:             "aws-lambda-old-runtime",
		standard:       "AWS Well-Architected",
		controlID:      "OPS-001",
		category:       assessment.CategoryOperationalExcellence,
		severity:       assessment.SeverityMedium,
		appliesTo:      []provider.ResourceType{provider.ResourceTypeLambda},
		recommendation: "Upgrade to a supported runtime version to receive security patches",
		describe: func(resource storage.Resource) string {
			runtime := metautil.GetString(resource.RawMetadata, "runtime")
			return fmt.Sprintf("Lambda %s uses deprecated runtime %s", resource.Name, runtime)
		},
		violated: func(meta map[string]any) bool {
			return deprecated[metautil.GetString(meta, "runtime")]
		},
	}
}

// lambdaHighMemoryRule checks for over-provisioned Lambda functions.
func lambdaHighMemoryRule() *simpleRule {
	return &simpleRule{
		id:             "aws-lambda-high-memory",
		standard:       "AWS Well-Architected",
		controlID:      "COST-001",
		category:       assessment.CategoryCostOptimization,
		severity:       assessment.SeverityLow,
		appliesTo:      []provider.ResourceType{provider.ResourceTypeLambda},
		recommendation: "Review actual memory usage and reduce allocation if peak usage is well below configured memory",
		describe: func(resource storage.Resource) string {
			memorySize, _ := resource.RawMetadata["memory_size"].(float64)
			return fmt.Sprintf("Lambda %s has %.0f MB memory allocated (potentially over-provisioned)", resource.Name, memorySize)
		},
		violated: func(meta map[string]any) bool {
			memorySize, ok := meta["memory_size"].(float64)
			return ok && memorySize >= 3008
		},
	}
}

// lambdaNoDeadLetterRule checks for Lambda without DLQ configured.
func lambdaNoDeadLetterRule() *simpleRule {
	return &simpleRule{
		id:             "aws-lambda-no-dlq",
		standard:       "AWS Well-Architected",
		controlID:      "REL-001",
		category:       assessment.CategoryReliability,
		severity:       assessment.SeverityLow,
		appliesTo:      []provider.ResourceType{provider.ResourceTypeLambda},
		recommendation: "Configure a DLQ (SQS or SNS) to capture failed async invocations",
		describe: func(resource storage.Resource) string {
			return fmt.Sprintf("Lambda %s has no dead letter queue configured", resource.Name)
		},
		violated: func(meta map[string]any) bool {
			dlqArn := metautil.GetString(meta, "dead_letter_config_target_arn")
			if dlqArn == "" {
				if dlq, ok := meta["dead_letter_config"].(map[string]any); ok {
					dlqArn, _ = dlq["target_arn"].(string)
				}
			}
			return dlqArn == ""
		},
	}
}
