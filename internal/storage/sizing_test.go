package storage

import (
	"testing"
)

func TestInsertSizing(t *testing.T) {
	store := testStore(t)

	entry := &SizingEntry{
		AssessmentID: "test-assessment-1",
		Category:     "compute",
		ResourceID:   "i-1234567890abcdef0",
		Data: map[string]any{
			"instance_type":   "m5.large",
			"vcpus":           float64(2),
			"memory_gb":       float64(8),
			"cpu_utilization": float64(35.5),
		},
	}

	err := store.InsertSizing(entry)
	if err != nil {
		t.Fatalf("InsertSizing() error: %v", err)
	}

	if entry.ID == 0 {
		t.Error("expected sizing entry ID to be set after insert")
	}
}

func TestInsertSizingNilData(t *testing.T) {
	store := testStore(t)

	entry := &SizingEntry{
		AssessmentID: "test-assessment-1",
		Category:     "storage",
		ResourceID:   "vol-nil-data",
		Data:         nil,
	}

	err := store.InsertSizing(entry)
	if err != nil {
		t.Fatalf("InsertSizing() with nil data error: %v", err)
	}

	// Retrieve and verify round-trip.
	entries, err := store.GetSizingByAssessment("test-assessment-1")
	if err != nil {
		t.Fatalf("GetSizingByAssessment() error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 sizing entry, got %d", len(entries))
	}
	if entries[0].Data == nil {
		t.Error("expected Data to be non-nil after round-trip (should be empty map)")
	}
	if len(entries[0].Data) != 0 {
		t.Errorf("expected empty data map, got %v", entries[0].Data)
	}
}

func TestGetSizingByAssessment(t *testing.T) {
	store := testStore(t)

	entries := []*SizingEntry{
		{
			AssessmentID: "test-assessment-1",
			Category:     "compute",
			ResourceID:   "i-001",
			Data:         map[string]any{"instance_type": "m5.large", "vcpus": float64(2)},
		},
		{
			AssessmentID: "test-assessment-1",
			Category:     "database",
			ResourceID:   "db-001",
			Data:         map[string]any{"engine": "mysql", "instance_class": "db.r5.large"},
		},
		{
			AssessmentID: "test-assessment-1",
			Category:     "storage",
			ResourceID:   "vol-001",
			Data:         map[string]any{"storage_type": "gp3", "capacity_gb": float64(100)},
		},
	}

	for _, e := range entries {
		if err := store.InsertSizing(e); err != nil {
			t.Fatalf("InsertSizing() error: %v", err)
		}
	}

	got, err := store.GetSizingByAssessment("test-assessment-1")
	if err != nil {
		t.Fatalf("GetSizingByAssessment() error: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 sizing entries, got %d", len(got))
	}
}

func TestGetSizingByAssessmentEmpty(t *testing.T) {
	store := testStore(t)

	got, err := store.GetSizingByAssessment("nonexistent-assessment")
	if err != nil {
		t.Fatalf("GetSizingByAssessment() error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 sizing entries for nonexistent assessment, got %d", len(got))
	}
}

func TestSizingRoundTripWithComplexData(t *testing.T) {
	store := testStore(t)

	entry := &SizingEntry{
		AssessmentID: "test-assessment-1",
		Category:     "kubernetes",
		ResourceID:   "eks-cluster-1",
		Data: map[string]any{
			"node_count":  float64(5),
			"total_vcpus": float64(20),
			"total_mem_gb": float64(80),
			"node_types":  []any{"m5.xlarge", "c5.2xlarge"},
			"metadata": map[string]any{
				"version":  "1.28",
				"platform": "EKS",
			},
		},
	}

	if err := store.InsertSizing(entry); err != nil {
		t.Fatalf("InsertSizing() error: %v", err)
	}

	got, err := store.GetSizingByAssessment("test-assessment-1")
	if err != nil {
		t.Fatalf("GetSizingByAssessment() error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 sizing entry, got %d", len(got))
	}

	result := got[0]
	if result.AssessmentID != "test-assessment-1" {
		t.Errorf("AssessmentID = %q, want %q", result.AssessmentID, "test-assessment-1")
	}
	if result.Category != "kubernetes" {
		t.Errorf("Category = %q, want %q", result.Category, "kubernetes")
	}
	if result.ResourceID != "eks-cluster-1" {
		t.Errorf("ResourceID = %q, want %q", result.ResourceID, "eks-cluster-1")
	}

	// Verify nested data round-trips correctly.
	if result.Data["node_count"] != float64(5) {
		t.Errorf("Data[node_count] = %v, want 5", result.Data["node_count"])
	}

	nodeTypes, ok := result.Data["node_types"].([]any)
	if !ok {
		t.Fatal("expected node_types to be []any")
	}
	if len(nodeTypes) != 2 {
		t.Fatalf("expected 2 node types, got %d", len(nodeTypes))
	}
	if nodeTypes[0] != "m5.xlarge" {
		t.Errorf("node_types[0] = %v, want m5.xlarge", nodeTypes[0])
	}

	metadata, ok := result.Data["metadata"].(map[string]any)
	if !ok {
		t.Fatal("expected metadata to be map[string]any")
	}
	if metadata["version"] != "1.28" {
		t.Errorf("metadata.version = %v, want 1.28", metadata["version"])
	}
}
