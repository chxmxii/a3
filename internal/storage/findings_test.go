package storage

import (
	"path/filepath"
	"testing"
)

// helper to create a store with a test assessment already inserted.
func setupFindingsTest(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}

	// Insert a parent assessment since findings reference assessments via FK.
	_, err = store.DB.Exec(`
		INSERT INTO assessments (id, profile, provider, status, started_at, regions)
		VALUES ('assess-001', 'test-profile', 'aws', 'completed', '2024-01-01T00:00:00Z', '["us-east-1"]')`)
	if err != nil {
		t.Fatalf("inserting test assessment: %v", err)
	}

	t.Cleanup(func() { store.Close() })
	return store
}

func TestInsertFinding(t *testing.T) {
	store := setupFindingsTest(t)

	f := &Finding{
		AssessmentID:   "assess-001",
		Severity:       "high",
		ResourceID:     "arn:aws:ec2:us-east-1:123456789012:instance/i-abc123",
		Description:    "Security group allows unrestricted SSH access",
		Recommendation: "Restrict SSH access to specific IP ranges",
		StandardName:   "CIS AWS 1.5",
		ControlID:      "5.2.1",
		Category:       "Security",
	}

	if err := store.InsertFinding(f); err != nil {
		t.Fatalf("InsertFinding() error: %v", err)
	}

	if f.ID == 0 {
		t.Error("expected finding ID to be set after insert")
	}
}

func TestInsertFindingForeignKeyViolation(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer store.Close()

	// Insert finding referencing non-existent assessment; should fail due to FK constraint.
	f := &Finding{
		AssessmentID:   "non-existent",
		Severity:       "low",
		ResourceID:     "res-1",
		Description:    "test",
		Recommendation: "test",
		StandardName:   "test",
		ControlID:      "1.1",
		Category:       "Security",
	}

	if err := store.InsertFinding(f); err == nil {
		t.Error("expected foreign key violation error, got nil")
	}
}

func TestGetFindingsByAssessment(t *testing.T) {
	store := setupFindingsTest(t)

	// Insert multiple findings.
	findings := []*Finding{
		{
			AssessmentID:   "assess-001",
			Severity:       "critical",
			ResourceID:     "res-1",
			Description:    "Root account without MFA",
			Recommendation: "Enable MFA on root account",
			StandardName:   "CIS AWS 1.5",
			ControlID:      "1.1.1",
			Category:       "Security",
		},
		{
			AssessmentID:   "assess-001",
			Severity:       "medium",
			ResourceID:     "res-2",
			Description:    "S3 bucket without versioning",
			Recommendation: "Enable versioning on the bucket",
			StandardName:   "AWS Well-Architected",
			ControlID:      "REL-3",
			Category:       "Reliability",
		},
		{
			AssessmentID:   "assess-001",
			Severity:       "low",
			ResourceID:     "res-3",
			Description:    "Instance without detailed monitoring",
			Recommendation: "Enable detailed monitoring",
			StandardName:   "AWS Well-Architected",
			ControlID:      "PERF-1",
			Category:       "Performance",
		},
	}

	for _, f := range findings {
		if err := store.InsertFinding(f); err != nil {
			t.Fatalf("InsertFinding() error: %v", err)
		}
	}

	got, err := store.GetFindingsByAssessment("assess-001")
	if err != nil {
		t.Fatalf("GetFindingsByAssessment() error: %v", err)
	}

	if len(got) != 3 {
		t.Fatalf("expected 3 findings, got %d", len(got))
	}

	// Verify fields on first finding.
	if got[0].Severity != "critical" {
		t.Errorf("expected severity 'critical', got %q", got[0].Severity)
	}
	if got[0].ResourceID != "res-1" {
		t.Errorf("expected resource_id 'res-1', got %q", got[0].ResourceID)
	}
	if got[0].Category != "Security" {
		t.Errorf("expected category 'Security', got %q", got[0].Category)
	}
}

func TestGetFindingsByAssessmentEmpty(t *testing.T) {
	store := setupFindingsTest(t)

	got, err := store.GetFindingsByAssessment("assess-001")
	if err != nil {
		t.Fatalf("GetFindingsByAssessment() error: %v", err)
	}

	if got != nil {
		t.Errorf("expected nil for empty result, got %v", got)
	}
}

func TestInsertFindingRoundTrip(t *testing.T) {
	store := setupFindingsTest(t)

	original := &Finding{
		AssessmentID:   "assess-001",
		Severity:       "critical",
		ResourceID:     "arn:aws:iam::123456789012:root",
		Description:    "Root account MFA is not enabled",
		Recommendation: "Enable hardware MFA on the root account to prevent unauthorized access",
		StandardName:   "CIS AWS Foundations Benchmark v1.5",
		ControlID:      "1.5",
		Category:       "Security",
	}

	if err := store.InsertFinding(original); err != nil {
		t.Fatalf("InsertFinding() error: %v", err)
	}

	got, err := store.GetFindingsByAssessment("assess-001")
	if err != nil {
		t.Fatalf("GetFindingsByAssessment() error: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(got))
	}

	f := got[0]
	if f.ID != original.ID {
		t.Errorf("ID: got %d, want %d", f.ID, original.ID)
	}
	if f.AssessmentID != original.AssessmentID {
		t.Errorf("AssessmentID: got %q, want %q", f.AssessmentID, original.AssessmentID)
	}
	if f.Severity != original.Severity {
		t.Errorf("Severity: got %q, want %q", f.Severity, original.Severity)
	}
	if f.ResourceID != original.ResourceID {
		t.Errorf("ResourceID: got %q, want %q", f.ResourceID, original.ResourceID)
	}
	if f.Description != original.Description {
		t.Errorf("Description: got %q, want %q", f.Description, original.Description)
	}
	if f.Recommendation != original.Recommendation {
		t.Errorf("Recommendation: got %q, want %q", f.Recommendation, original.Recommendation)
	}
	if f.StandardName != original.StandardName {
		t.Errorf("StandardName: got %q, want %q", f.StandardName, original.StandardName)
	}
	if f.ControlID != original.ControlID {
		t.Errorf("ControlID: got %q, want %q", f.ControlID, original.ControlID)
	}
	if f.Category != original.Category {
		t.Errorf("Category: got %q, want %q", f.Category, original.Category)
	}
}
