package storage

import (
	"path/filepath"
	"strings"
	"testing"
)

// testStore creates an in-memory Store for testing. It inserts a test
// assessment record to satisfy the foreign key constraint on resources.
func testStore(t *testing.T) *Store {
	t.Helper()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	t.Cleanup(func() { store.Close() })

	// Insert a test assessment to satisfy foreign key constraints.
	_, err = store.DB.Exec(`
		INSERT INTO assessments (id, profile, provider, status, started_at, regions)
		VALUES ('test-assessment-1', 'default', 'aws', 'in_progress', '2024-01-01T00:00:00Z', '["us-east-1"]')`)
	if err != nil {
		t.Fatalf("inserting test assessment: %v", err)
	}

	return store
}

func TestInsertResource(t *testing.T) {
	store := testStore(t)

	r := &Resource{
		AssessmentID: "test-assessment-1",
		ProviderType: "aws",
		ResourceType: "ec2_instance",
		ResourceID:   "i-1234567890abcdef0",
		Region:       "us-east-1",
		Name:         "my-instance",
		Tags:         map[string]string{"env": "prod", "team": "platform"},
		RawMetadata:  map[string]any{"instance_type": "m5.large", "state": "running"},
	}

	err := store.InsertResource(r)
	if err != nil {
		t.Fatalf("InsertResource() error: %v", err)
	}

	if r.ID == 0 {
		t.Error("expected resource ID to be set after insert")
	}
}

func TestInsertResourceDuplicateConstraint(t *testing.T) {
	store := testStore(t)

	r := &Resource{
		AssessmentID: "test-assessment-1",
		ProviderType: "aws",
		ResourceType: "ec2_instance",
		ResourceID:   "i-duplicate",
		Region:       "us-east-1",
		Name:         "first",
		Tags:         map[string]string{},
		RawMetadata:  map[string]any{},
	}

	if err := store.InsertResource(r); err != nil {
		t.Fatalf("first InsertResource() error: %v", err)
	}

	// Attempt duplicate insert with same (assessment_id, resource_id).
	r2 := &Resource{
		AssessmentID: "test-assessment-1",
		ProviderType: "aws",
		ResourceType: "ec2_instance",
		ResourceID:   "i-duplicate",
		Region:       "us-east-1",
		Name:         "second",
		Tags:         map[string]string{},
		RawMetadata:  map[string]any{},
	}

	err := store.InsertResource(r2)
	if err == nil {
		t.Fatal("expected error on duplicate insert, got nil")
	}
	if !strings.Contains(err.Error(), "UNIQUE constraint") {
		t.Errorf("expected UNIQUE constraint error, got: %v", err)
	}
}

func TestInsertResourceNilMaps(t *testing.T) {
	store := testStore(t)

	r := &Resource{
		AssessmentID: "test-assessment-1",
		ProviderType: "aws",
		ResourceType: "s3_bucket",
		ResourceID:   "arn:aws:s3:::my-bucket",
		Region:       "us-east-1",
		Name:         "my-bucket",
		Tags:         nil,
		RawMetadata:  nil,
	}

	err := store.InsertResource(r)
	if err != nil {
		t.Fatalf("InsertResource() with nil maps error: %v", err)
	}

	// Retrieve and verify defaults.
	got, err := store.GetResourceByID("test-assessment-1", "arn:aws:s3:::my-bucket")
	if err != nil {
		t.Fatalf("GetResourceByID() error: %v", err)
	}
	if got.Tags == nil {
		t.Error("expected Tags to be non-nil after round-trip (should be empty map from null JSON)")
	}
}

func TestGetResourcesByAssessment(t *testing.T) {
	store := testStore(t)

	resources := []*Resource{
		{
			AssessmentID: "test-assessment-1",
			ProviderType: "aws",
			ResourceType: "ec2_instance",
			ResourceID:   "i-001",
			Region:       "us-east-1",
			Name:         "instance-1",
			Tags:         map[string]string{"env": "prod"},
			RawMetadata:  map[string]any{"state": "running"},
		},
		{
			AssessmentID: "test-assessment-1",
			ProviderType: "aws",
			ResourceType: "s3_bucket",
			ResourceID:   "arn:aws:s3:::bucket-1",
			Region:       "us-west-2",
			Name:         "bucket-1",
			Tags:         map[string]string{},
			RawMetadata:  map[string]any{},
		},
	}

	for _, r := range resources {
		if err := store.InsertResource(r); err != nil {
			t.Fatalf("InsertResource() error: %v", err)
		}
	}

	got, err := store.GetResourcesByAssessment("test-assessment-1")
	if err != nil {
		t.Fatalf("GetResourcesByAssessment() error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(got))
	}
}

func TestGetResourcesByAssessmentEmpty(t *testing.T) {
	store := testStore(t)

	got, err := store.GetResourcesByAssessment("nonexistent-assessment")
	if err != nil {
		t.Fatalf("GetResourcesByAssessment() error: %v", err)
	}

	if len(got) != 0 {
		t.Errorf("expected 0 resources for nonexistent assessment, got %d", len(got))
	}
}

func TestGetResourceByID(t *testing.T) {
	store := testStore(t)

	r := &Resource{
		AssessmentID: "test-assessment-1",
		ProviderType: "aws",
		ResourceType: "ec2_instance",
		ResourceID:   "i-abc123",
		Region:       "eu-west-1",
		Name:         "web-server",
		Tags:         map[string]string{"role": "web", "env": "staging"},
		RawMetadata:  map[string]any{"instance_type": "t3.micro", "vpc_id": "vpc-123"},
	}

	if err := store.InsertResource(r); err != nil {
		t.Fatalf("InsertResource() error: %v", err)
	}

	got, err := store.GetResourceByID("test-assessment-1", "i-abc123")
	if err != nil {
		t.Fatalf("GetResourceByID() error: %v", err)
	}
	if got == nil {
		t.Fatal("expected resource, got nil")
	}

	// Verify all fields round-trip correctly.
	if got.AssessmentID != "test-assessment-1" {
		t.Errorf("AssessmentID = %q, want %q", got.AssessmentID, "test-assessment-1")
	}
	if got.ProviderType != "aws" {
		t.Errorf("ProviderType = %q, want %q", got.ProviderType, "aws")
	}
	if got.ResourceType != "ec2_instance" {
		t.Errorf("ResourceType = %q, want %q", got.ResourceType, "ec2_instance")
	}
	if got.ResourceID != "i-abc123" {
		t.Errorf("ResourceID = %q, want %q", got.ResourceID, "i-abc123")
	}
	if got.Region != "eu-west-1" {
		t.Errorf("Region = %q, want %q", got.Region, "eu-west-1")
	}
	if got.Name != "web-server" {
		t.Errorf("Name = %q, want %q", got.Name, "web-server")
	}
	if got.Tags["role"] != "web" {
		t.Errorf("Tags[role] = %q, want %q", got.Tags["role"], "web")
	}
	if got.Tags["env"] != "staging" {
		t.Errorf("Tags[env] = %q, want %q", got.Tags["env"], "staging")
	}
	if got.RawMetadata["instance_type"] != "t3.micro" {
		t.Errorf("RawMetadata[instance_type] = %v, want %q", got.RawMetadata["instance_type"], "t3.micro")
	}
	if got.RawMetadata["vpc_id"] != "vpc-123" {
		t.Errorf("RawMetadata[vpc_id] = %v, want %q", got.RawMetadata["vpc_id"], "vpc-123")
	}
}

func TestGetResourceByIDNotFound(t *testing.T) {
	store := testStore(t)

	got, err := store.GetResourceByID("test-assessment-1", "nonexistent-resource")
	if err != nil {
		t.Fatalf("GetResourceByID() error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for nonexistent resource, got %+v", got)
	}
}

func TestResourceRoundTripWithComplexMetadata(t *testing.T) {
	store := testStore(t)

	r := &Resource{
		AssessmentID: "test-assessment-1",
		ProviderType: "oci",
		ResourceType: "compute_instance",
		ResourceID:   "ocid1.instance.oc1.phx.abc123",
		Region:       "us-phoenix-1",
		Name:         "complex-instance",
		Tags: map[string]string{
			"env":        "production",
			"team":       "infra",
			"cost-center": "12345",
		},
		RawMetadata: map[string]any{
			"shape":          "VM.Standard2.1",
			"ocpus":          float64(1),
			"memory_in_gbs":  float64(15),
			"source_details": map[string]any{"source_type": "image", "image_id": "ocid1.image.oc1"},
			"vnic_attachments": []any{
				map[string]any{"vnic_id": "ocid1.vnic.oc1", "subnet_id": "ocid1.subnet.oc1"},
			},
		},
	}

	if err := store.InsertResource(r); err != nil {
		t.Fatalf("InsertResource() error: %v", err)
	}

	got, err := store.GetResourceByID("test-assessment-1", "ocid1.instance.oc1.phx.abc123")
	if err != nil {
		t.Fatalf("GetResourceByID() error: %v", err)
	}
	if got == nil {
		t.Fatal("expected resource, got nil")
	}

	// Verify nested metadata round-trips correctly.
	sourceDetails, ok := got.RawMetadata["source_details"].(map[string]any)
	if !ok {
		t.Fatal("expected source_details to be map[string]any")
	}
	if sourceDetails["source_type"] != "image" {
		t.Errorf("source_details.source_type = %v, want %q", sourceDetails["source_type"], "image")
	}

	vnicAttachments, ok := got.RawMetadata["vnic_attachments"].([]any)
	if !ok {
		t.Fatal("expected vnic_attachments to be []any")
	}
	if len(vnicAttachments) != 1 {
		t.Fatalf("expected 1 vnic attachment, got %d", len(vnicAttachments))
	}
}
