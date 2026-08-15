package storage

import (
	"testing"
)

func TestInsertCostEstimate(t *testing.T) {
	store := testStore(t)

	monthlyCost := 150.50
	confidence := "high"
	est := &CostEstimate{
		AssessmentID:  "test-assessment-1",
		ResourceID:    "i-1234567890abcdef0",
		ResourceType:  "ec2_instance",
		MonthlyCost:   &monthlyCost,
		Confidence:    &confidence,
		Category:      "Compute",
		IdleFlag:      false,
		OversizedFlag: true,
		Unestimable:   false,
	}

	err := store.InsertCostEstimate(est)
	if err != nil {
		t.Fatalf("InsertCostEstimate() error: %v", err)
	}

	if est.ID == 0 {
		t.Error("expected cost estimate ID to be set after insert")
	}
}

func TestInsertCostEstimateUnestimable(t *testing.T) {
	store := testStore(t)

	est := &CostEstimate{
		AssessmentID:  "test-assessment-1",
		ResourceID:    "custom-resource-1",
		ResourceType:  "custom_type",
		MonthlyCost:   nil,
		Confidence:    nil,
		Category:      "Other",
		IdleFlag:      false,
		OversizedFlag: false,
		Unestimable:   true,
	}

	err := store.InsertCostEstimate(est)
	if err != nil {
		t.Fatalf("InsertCostEstimate() error: %v", err)
	}

	// Retrieve and verify nil fields round-trip.
	costs, err := store.GetCostsByAssessment("test-assessment-1")
	if err != nil {
		t.Fatalf("GetCostsByAssessment() error: %v", err)
	}
	if len(costs) != 1 {
		t.Fatalf("expected 1 cost estimate, got %d", len(costs))
	}

	got := costs[0]
	if got.MonthlyCost != nil {
		t.Errorf("MonthlyCost = %v, want nil", got.MonthlyCost)
	}
	if got.Confidence != nil {
		t.Errorf("Confidence = %v, want nil", got.Confidence)
	}
	if !got.Unestimable {
		t.Error("expected Unestimable to be true")
	}
}

func TestGetCostsByAssessment(t *testing.T) {
	store := testStore(t)

	monthlyCost1 := 100.0
	confidence1 := "high"
	monthlyCost2 := 50.0
	confidence2 := "medium"

	estimates := []*CostEstimate{
		{
			AssessmentID:  "test-assessment-1",
			ResourceID:    "i-001",
			ResourceType:  "ec2_instance",
			MonthlyCost:   &monthlyCost1,
			Confidence:    &confidence1,
			Category:      "Compute",
			IdleFlag:      true,
			OversizedFlag: false,
			Unestimable:   false,
		},
		{
			AssessmentID:  "test-assessment-1",
			ResourceID:    "vol-001",
			ResourceType:  "ebs_volume",
			MonthlyCost:   &monthlyCost2,
			Confidence:    &confidence2,
			Category:      "Storage",
			IdleFlag:      false,
			OversizedFlag: false,
			Unestimable:   false,
		},
	}

	for _, est := range estimates {
		if err := store.InsertCostEstimate(est); err != nil {
			t.Fatalf("InsertCostEstimate() error: %v", err)
		}
	}

	got, err := store.GetCostsByAssessment("test-assessment-1")
	if err != nil {
		t.Fatalf("GetCostsByAssessment() error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 cost estimates, got %d", len(got))
	}
}

func TestGetCostsByAssessmentEmpty(t *testing.T) {
	store := testStore(t)

	got, err := store.GetCostsByAssessment("nonexistent-assessment")
	if err != nil {
		t.Fatalf("GetCostsByAssessment() error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 cost estimates for nonexistent assessment, got %d", len(got))
	}
}

func TestCostEstimateRoundTrip(t *testing.T) {
	store := testStore(t)

	monthlyCost := 325.75
	confidence := "low"
	est := &CostEstimate{
		AssessmentID:  "test-assessment-1",
		ResourceID:    "i-roundtrip-1",
		ResourceType:  "ec2_instance",
		MonthlyCost:   &monthlyCost,
		Confidence:    &confidence,
		Category:      "Compute",
		IdleFlag:      true,
		OversizedFlag: true,
		Unestimable:   false,
	}

	if err := store.InsertCostEstimate(est); err != nil {
		t.Fatalf("InsertCostEstimate() error: %v", err)
	}

	costs, err := store.GetCostsByAssessment("test-assessment-1")
	if err != nil {
		t.Fatalf("GetCostsByAssessment() error: %v", err)
	}
	if len(costs) != 1 {
		t.Fatalf("expected 1 cost estimate, got %d", len(costs))
	}

	got := costs[0]
	if got.AssessmentID != "test-assessment-1" {
		t.Errorf("AssessmentID = %q, want %q", got.AssessmentID, "test-assessment-1")
	}
	if got.ResourceID != "i-roundtrip-1" {
		t.Errorf("ResourceID = %q, want %q", got.ResourceID, "i-roundtrip-1")
	}
	if got.ResourceType != "ec2_instance" {
		t.Errorf("ResourceType = %q, want %q", got.ResourceType, "ec2_instance")
	}
	if got.MonthlyCost == nil || *got.MonthlyCost != 325.75 {
		t.Errorf("MonthlyCost = %v, want 325.75", got.MonthlyCost)
	}
	if got.Confidence == nil || *got.Confidence != "low" {
		t.Errorf("Confidence = %v, want low", got.Confidence)
	}
	if got.Category != "Compute" {
		t.Errorf("Category = %q, want %q", got.Category, "Compute")
	}
	if !got.IdleFlag {
		t.Error("expected IdleFlag to be true")
	}
	if !got.OversizedFlag {
		t.Error("expected OversizedFlag to be true")
	}
	if got.Unestimable {
		t.Error("expected Unestimable to be false")
	}
}
