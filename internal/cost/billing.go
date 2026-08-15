package cost

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/chxmxii/a3/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BillingCost represents actual cost from AWS Cost Explorer.
type BillingCost struct {
	Service     string
	MonthlyCost float64
}

// BillingSummary holds real billing data from AWS Cost Explorer via Steampipe.
type BillingSummary struct {
	TotalMonthlyCost float64
	ByService        []BillingCost
	Source           string // "billing_api" or "static_estimate"
	Message          string // informational message about accuracy
}

// QueryBilling attempts to get real cost data from Steampipe's aws_cost_by_service_monthly table.
// Returns nil if the table doesn't exist or the query fails.
func QueryBilling(ctx context.Context, pool *pgxpool.Pool) (*BillingSummary, error) {
	// Check if the cost table exists.
	var exists bool
	err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_name = 'aws_cost_by_service_monthly'
		)
	`).Scan(&exists)
	if err != nil || !exists {
		return nil, fmt.Errorf("billing table not available")
	}

	// Query the most recent month's costs grouped by service.
	rows, err := pool.Query(ctx, `
		SELECT
			service,
			blended_cost_amount::numeric
		FROM aws_cost_by_service_monthly
		WHERE period_start >= date_trunc('month', current_date - interval '1 month')
		  AND period_start < date_trunc('month', current_date)
		ORDER BY blended_cost_amount DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("querying billing data: %w", err)
	}
	defer rows.Close()

	summary := &BillingSummary{
		Source:  "billing_api",
		Message: "Based on last month's actual AWS billing data via Cost Explorer",
	}

	for rows.Next() {
		var service string
		var cost float64
		if err := rows.Scan(&service, &cost); err != nil {
			continue
		}
		if cost <= 0 {
			continue
		}
		summary.ByService = append(summary.ByService, BillingCost{
			Service:     service,
			MonthlyCost: cost,
		})
		summary.TotalMonthlyCost += cost
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading billing rows: %w", err)
	}

	if len(summary.ByService) == 0 {
		return nil, fmt.Errorf("no billing data returned")
	}

	// Sort by cost descending.
	sort.Slice(summary.ByService, func(i, j int) bool {
		return summary.ByService[i].MonthlyCost > summary.ByService[j].MonthlyCost
	})

	return summary, nil
}

// StoreBillingCosts persists billing data to the cost_estimates table.
func StoreBillingCosts(store *storage.Store, assessmentID string, billing *BillingSummary) {
	for _, svc := range billing.ByService {
		cost := svc.MonthlyCost
		conf := "high"
		est := &storage.CostEstimate{
			AssessmentID: assessmentID,
			ResourceID:   "service:" + svc.Service,
			ResourceType: "aws_service",
			Category:     mapServiceToCategory(svc.Service),
			MonthlyCost:  &cost,
			Confidence:   &conf,
		}
		_ = store.InsertCostEstimate(est)
	}
}

func mapServiceToCategory(service string) string {
	switch {
	case strings.Contains(service, "EC2") || strings.Contains(service, "Compute"):
		return string(CostCategoryCompute)
	case strings.Contains(service, "RDS") || strings.Contains(service, "Database") || strings.Contains(service, "DynamoDB") || strings.Contains(service, "ElastiCache"):
		return string(CostCategoryDatabase)
	case strings.Contains(service, "S3") || strings.Contains(service, "EFS") || strings.Contains(service, "Backup") || strings.Contains(service, "Storage"):
		return string(CostCategoryStorage)
	case strings.Contains(service, "VPC") || strings.Contains(service, "CloudFront") || strings.Contains(service, "Route 53") || strings.Contains(service, "Transfer") || strings.Contains(service, "ELB"):
		return string(CostCategoryNetworking)
	case strings.Contains(service, "EKS") || strings.Contains(service, "ECS") || strings.Contains(service, "Fargate"):
		return string(CostCategoryKubernetes)
	case strings.Contains(service, "Lambda"):
		return string(CostCategoryServerless)
	default:
		return string(CostCategoryOther)
	}
}
