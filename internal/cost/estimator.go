package cost

import (
	"log"
	"sort"

	"github.com/chxmxii/a3/internal/metautil"
	"github.com/chxmxii/a3/internal/storage"
)

// Estimator calculates cost estimates for discovered resources.
type Estimator struct {
	store *storage.Store
}

// CostSummary holds aggregated cost information.
type CostSummary struct {
	TotalMonthlyCost float64
	ByCategory       map[CostCategory]float64
	TopDrivers       []CostDriver
	IdleCount        int
	OversizedCount   int
}

// CostDriver represents a top cost-driving resource.
type CostDriver struct {
	ResourceID   string
	ResourceType string
	Name         string
	MonthlyCost  float64
}

// NewEstimator creates a new cost estimator.
func NewEstimator(store *storage.Store) *Estimator {
	return &Estimator{store: store}
}

// Estimate processes all resources and calculates cost estimates.
func (e *Estimator) Estimate(assessmentID string) (*CostSummary, error) {
	resources, err := e.store.GetResourcesByAssessment(assessmentID)
	if err != nil {
		return nil, err
	}

	summary := &CostSummary{
		ByCategory: make(map[CostCategory]float64),
	}

	var allDrivers []CostDriver

	for _, res := range resources {
		est := e.estimateResource(res)
		if est == nil {
			continue
		}

		est.AssessmentID = assessmentID
		if err := e.store.InsertCostEstimate(est); err != nil {
			log.Printf("[cost] failed to store estimate for %s: %v", res.ResourceID, err)
			continue
		}

		if est.MonthlyCost != nil {
			cost := *est.MonthlyCost
			summary.TotalMonthlyCost += cost
			summary.ByCategory[CostCategory(est.Category)] += cost

			allDrivers = append(allDrivers, CostDriver{
				ResourceID:   res.ResourceID,
				ResourceType: res.ResourceType,
				Name:         res.Name,
				MonthlyCost:  cost,
			})
		}

		if est.IdleFlag {
			summary.IdleCount++
		}
		if est.OversizedFlag {
			summary.OversizedCount++
		}
	}

	// Find top 5 cost drivers.
	sort.Slice(allDrivers, func(i, j int) bool {
		return allDrivers[i].MonthlyCost > allDrivers[j].MonthlyCost
	})
	if len(allDrivers) > 5 {
		allDrivers = allDrivers[:5]
	}
	summary.TopDrivers = allDrivers

	return summary, nil
}

func (e *Estimator) estimateResource(res storage.Resource) *storage.CostEstimate {
	category := string(GetCategory(res.ResourceType))

	switch res.ResourceType {
	case "ec2_instance":
		return e.estimateEC2(res, category)
	case "rds_instance":
		return e.estimateRDS(res, category)
	case "ebs_volume":
		return e.estimateEBS(res, category)
	case "nat_gateway":
		return e.estimateNATGW(res, category)
	case "alb":
		return e.estimateALB(res, category)
	case "nlb":
		return e.estimateNLB(res, category)
	case "eks_cluster":
		return e.estimateEKS(res, category)
	case "lambda_function":
		return e.estimateLambda(res, category)
	default:
		return nil
	}
}

func (e *Estimator) estimateEC2(res storage.Resource, category string) *storage.CostEstimate {
	instanceType := metautil.GetString(res.RawMetadata, "instance_type", "instanceType", "instance_type_name")

	est := &storage.CostEstimate{
		ResourceID:   res.ResourceID,
		ResourceType: res.ResourceType,
		Category:     category,
	}

	// Check if instance is stopped (idle).
	state := metautil.GetString(res.RawMetadata, "instance_state", "state")
	if state == "stopped" {
		est.IdleFlag = true
	}

	if instanceType == "" {
		est.Unestimable = true
		return est
	}

	hourly, ok := instanceHourly(instanceType)
	if !ok {
		// Try without dot suffix variations (e.g., "m5.large" vs "m5large").
		est.Unestimable = true
		conf := "low"
		est.Confidence = &conf
		return est
	}

	monthly := hourly * HoursPerMonth
	est.MonthlyCost = &monthly
	conf := "medium"
	est.Confidence = &conf

	return est
}

func (e *Estimator) estimateRDS(res storage.Resource, category string) *storage.CostEstimate {
	instanceClass := metautil.GetString(res.RawMetadata, "db_instance_class", "instanceClass")

	est := &storage.CostEstimate{
		ResourceID:   res.ResourceID,
		ResourceType: res.ResourceType,
		Category:     category,
	}

	hourly, ok := instanceHourly(instanceClass)
	if !ok {
		est.Unestimable = true
		return est
	}

	monthly := hourly * HoursPerMonth
	est.MonthlyCost = &monthly
	conf := "medium"
	est.Confidence = &conf

	return est
}

func (e *Estimator) estimateEBS(res storage.Resource, category string) *storage.CostEstimate {
	est := &storage.CostEstimate{
		ResourceID:   res.ResourceID,
		ResourceType: res.ResourceType,
		Category:     category,
	}

	volType := metautil.GetString(res.RawMetadata, "volume_type")
	if volType == "" {
		volType = "gp2"
	}

	sizeGB := 0.0
	if size, ok := res.RawMetadata["size"].(float64); ok {
		sizeGB = size
	}

	pricePerGB, ok := storageGBMonth(volType)
	if !ok {
		pricePerGB = 0.10 // default to gp2 pricing
	}

	monthly := sizeGB * pricePerGB
	est.MonthlyCost = &monthly
	conf := "medium"
	est.Confidence = &conf

	// Flag unattached volumes as idle.
	if attachments, ok := res.RawMetadata["attachments"].([]any); ok && len(attachments) == 0 {
		est.IdleFlag = true
	}
	if state := metautil.GetString(res.RawMetadata, "state"); state == "available" {
		est.IdleFlag = true
	}

	return est
}

func (e *Estimator) estimateNATGW(res storage.Resource, category string) *storage.CostEstimate {
	monthly := serviceHourly(svcNATGateway) * HoursPerMonth
	conf := "high"
	return &storage.CostEstimate{
		ResourceID:   res.ResourceID,
		ResourceType: res.ResourceType,
		Category:     category,
		MonthlyCost:  &monthly,
		Confidence:   &conf,
	}
}

func (e *Estimator) estimateALB(res storage.Resource, category string) *storage.CostEstimate {
	monthly := serviceHourly(svcALB) * HoursPerMonth
	conf := "low"
	return &storage.CostEstimate{
		ResourceID:   res.ResourceID,
		ResourceType: res.ResourceType,
		Category:     category,
		MonthlyCost:  &monthly,
		Confidence:   &conf,
	}
}

func (e *Estimator) estimateNLB(res storage.Resource, category string) *storage.CostEstimate {
	monthly := serviceHourly(svcNLB) * HoursPerMonth
	conf := "low"
	return &storage.CostEstimate{
		ResourceID:   res.ResourceID,
		ResourceType: res.ResourceType,
		Category:     category,
		MonthlyCost:  &monthly,
		Confidence:   &conf,
	}
}

func (e *Estimator) estimateEKS(res storage.Resource, category string) *storage.CostEstimate {
	monthly := serviceHourly(svcEKSControlPlane) * HoursPerMonth
	conf := "high"
	return &storage.CostEstimate{
		ResourceID:   res.ResourceID,
		ResourceType: res.ResourceType,
		Category:     category,
		MonthlyCost:  &monthly,
		Confidence:   &conf,
	}
}

func (e *Estimator) estimateLambda(res storage.Resource, category string) *storage.CostEstimate {
	// Lambda cost is usage-based — provide a minimal estimate.
	monthly := 0.0
	conf := "low"
	return &storage.CostEstimate{
		ResourceID:   res.ResourceID,
		ResourceType: res.ResourceType,
		Category:     category,
		MonthlyCost:  &monthly,
		Confidence:   &conf,
	}
}
