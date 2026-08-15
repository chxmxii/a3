package storage

import (
	"os"
	"path/filepath"
	"testing"
)

// newTestStore creates an in-memory Store for testing.
func newTestStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("opening test store: %v", err)
	}
	t.Cleanup(func() {
		store.Close()
		os.Remove(dbPath)
	})
	return store
}

// insertTestAssessment inserts a minimal assessment row needed for foreign key constraints.
func insertTestAssessment(t *testing.T, store *Store, id string) {
	t.Helper()
	_, err := store.DB.Exec(`
		INSERT INTO assessments (id, profile, provider, status, started_at, regions)
		VALUES (?, 'test-profile', 'aws', 'in_progress', '2024-01-01T00:00:00Z', '["us-east-1"]')`,
		id)
	if err != nil {
		t.Fatalf("inserting test assessment: %v", err)
	}
}

func TestInsertRelationship(t *testing.T) {
	store := newTestStore(t)
	insertTestAssessment(t, store, "assess-1")

	rel := &Relationship{
		AssessmentID:     "assess-1",
		SourceID:         "arn:aws:ec2:us-east-1:123456789012:instance/i-abc123",
		TargetID:         "arn:aws:ec2:us-east-1:123456789012:security-group/sg-xyz",
		RelationshipType: "uses_security_group",
		Status:           "resolved",
		UnresolvedReason: "",
		TargetRegion:     "us-east-1",
		TargetAccount:    "123456789012",
	}

	err := store.InsertRelationship(rel)
	if err != nil {
		t.Fatalf("InsertRelationship: %v", err)
	}

	if rel.ID == 0 {
		t.Error("expected non-zero ID after insert")
	}
}

func TestInsertRelationship_ForeignKeyViolation(t *testing.T) {
	store := newTestStore(t)

	rel := &Relationship{
		AssessmentID:     "nonexistent-assessment",
		SourceID:         "source-1",
		TargetID:         "target-1",
		RelationshipType: "references",
		Status:           "resolved",
	}

	err := store.InsertRelationship(rel)
	if err == nil {
		t.Fatal("expected error for foreign key violation, got nil")
	}
}

func TestGetRelationshipsByAssessment(t *testing.T) {
	store := newTestStore(t)
	insertTestAssessment(t, store, "assess-1")
	insertTestAssessment(t, store, "assess-2")

	// Insert relationships for assess-1
	rels := []*Relationship{
		{
			AssessmentID:     "assess-1",
			SourceID:         "source-1",
			TargetID:         "target-1",
			RelationshipType: "uses_security_group",
			Status:           "resolved",
		},
		{
			AssessmentID:     "assess-1",
			SourceID:         "source-2",
			TargetID:         "target-2",
			RelationshipType: "routes_to",
			Status:           "unresolved",
			UnresolvedReason: "target not found in inventory",
			TargetRegion:     "eu-west-1",
		},
	}

	// Insert a relationship for assess-2 (should not be returned)
	otherRel := &Relationship{
		AssessmentID:     "assess-2",
		SourceID:         "source-x",
		TargetID:         "target-x",
		RelationshipType: "attached_to",
		Status:           "resolved",
	}

	for _, r := range rels {
		if err := store.InsertRelationship(r); err != nil {
			t.Fatalf("InsertRelationship: %v", err)
		}
	}
	if err := store.InsertRelationship(otherRel); err != nil {
		t.Fatalf("InsertRelationship: %v", err)
	}

	// Query for assess-1
	got, err := store.GetRelationshipsByAssessment("assess-1")
	if err != nil {
		t.Fatalf("GetRelationshipsByAssessment: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 relationships, got %d", len(got))
	}

	// Verify first relationship
	if got[0].SourceID != "source-1" {
		t.Errorf("expected SourceID 'source-1', got %q", got[0].SourceID)
	}
	if got[0].RelationshipType != "uses_security_group" {
		t.Errorf("expected RelationshipType 'uses_security_group', got %q", got[0].RelationshipType)
	}
	if got[0].Status != "resolved" {
		t.Errorf("expected Status 'resolved', got %q", got[0].Status)
	}

	// Verify second relationship (unresolved)
	if got[1].Status != "unresolved" {
		t.Errorf("expected Status 'unresolved', got %q", got[1].Status)
	}
	if got[1].UnresolvedReason != "target not found in inventory" {
		t.Errorf("expected UnresolvedReason, got %q", got[1].UnresolvedReason)
	}
	if got[1].TargetRegion != "eu-west-1" {
		t.Errorf("expected TargetRegion 'eu-west-1', got %q", got[1].TargetRegion)
	}
}

func TestGetRelationshipsByAssessment_Empty(t *testing.T) {
	store := newTestStore(t)
	insertTestAssessment(t, store, "assess-empty")

	got, err := store.GetRelationshipsByAssessment("assess-empty")
	if err != nil {
		t.Fatalf("GetRelationshipsByAssessment: %v", err)
	}

	if got != nil {
		t.Errorf("expected nil slice for no results, got %v", got)
	}
}

func TestInsertRelationship_UnresolvedWithCrossAccount(t *testing.T) {
	store := newTestStore(t)
	insertTestAssessment(t, store, "assess-1")

	rel := &Relationship{
		AssessmentID:     "assess-1",
		SourceID:         "vpc-123",
		TargetID:         "vpc-456",
		RelationshipType: "peering",
		Status:           "unresolved",
		UnresolvedReason: "target in different account",
		TargetRegion:     "ap-southeast-1",
		TargetAccount:    "987654321098",
	}

	err := store.InsertRelationship(rel)
	if err != nil {
		t.Fatalf("InsertRelationship: %v", err)
	}

	// Retrieve and verify all fields round-trip
	got, err := store.GetRelationshipsByAssessment("assess-1")
	if err != nil {
		t.Fatalf("GetRelationshipsByAssessment: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("expected 1 relationship, got %d", len(got))
	}

	r := got[0]
	if r.AssessmentID != "assess-1" {
		t.Errorf("AssessmentID: got %q, want %q", r.AssessmentID, "assess-1")
	}
	if r.SourceID != "vpc-123" {
		t.Errorf("SourceID: got %q, want %q", r.SourceID, "vpc-123")
	}
	if r.TargetID != "vpc-456" {
		t.Errorf("TargetID: got %q, want %q", r.TargetID, "vpc-456")
	}
	if r.RelationshipType != "peering" {
		t.Errorf("RelationshipType: got %q, want %q", r.RelationshipType, "peering")
	}
	if r.Status != "unresolved" {
		t.Errorf("Status: got %q, want %q", r.Status, "unresolved")
	}
	if r.UnresolvedReason != "target in different account" {
		t.Errorf("UnresolvedReason: got %q, want %q", r.UnresolvedReason, "target in different account")
	}
	if r.TargetRegion != "ap-southeast-1" {
		t.Errorf("TargetRegion: got %q, want %q", r.TargetRegion, "ap-southeast-1")
	}
	if r.TargetAccount != "987654321098" {
		t.Errorf("TargetAccount: got %q, want %q", r.TargetAccount, "987654321098")
	}
}
