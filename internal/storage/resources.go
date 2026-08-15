package storage

import (
	"encoding/json"
	"fmt"
)

// Resource represents a discovered cloud resource stored in SQLite.
type Resource struct {
	ID           int64
	AssessmentID string
	ProviderType string
	ResourceType string
	ResourceID   string
	Region       string
	Name         string
	Tags         map[string]string
	RawMetadata  map[string]any
}

// InsertResource inserts a resource into the database, JSON-serializing the
// tags and raw_metadata fields. Returns an error if the (assessment_id, resource_id)
// pair already exists.
func (s *Store) InsertResource(resource *Resource) error {
	tags := resource.Tags
	if tags == nil {
		tags = map[string]string{}
	}
	rawMeta := resource.RawMetadata
	if rawMeta == nil {
		rawMeta = map[string]any{}
	}

	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return fmt.Errorf("marshaling tags: %w", err)
	}

	metadataJSON, err := json.Marshal(rawMeta)
	if err != nil {
		return fmt.Errorf("marshaling raw_metadata: %w", err)
	}

	result, err := s.DB.Exec(`
		INSERT INTO resources (assessment_id, provider_type, resource_type, resource_id, region, name, tags, raw_metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		resource.AssessmentID,
		resource.ProviderType,
		resource.ResourceType,
		resource.ResourceID,
		resource.Region,
		resource.Name,
		string(tagsJSON),
		string(metadataJSON),
	)
	if err != nil {
		return fmt.Errorf("inserting resource: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("getting last insert id: %w", err)
	}
	resource.ID = id

	return nil
}

// GetResourcesByAssessment returns all resources for the given assessment ID.
func (s *Store) GetResourcesByAssessment(assessmentID string) ([]Resource, error) {
	return queryAll(s.DB, "resources by assessment", scanResource, `
		SELECT id, assessment_id, provider_type, resource_type, resource_id, region, name, tags, raw_metadata
		FROM resources
		WHERE assessment_id = ?`, assessmentID)
}

// GetResourceByID returns a specific resource by assessment ID and resource ID.
// Returns nil if no matching resource is found.
func (s *Store) GetResourceByID(assessmentID, resourceID string) (*Resource, error) {
	return queryOne(s.DB, "resource by id", scanResource, `
		SELECT id, assessment_id, provider_type, resource_type, resource_id, region, name, tags, raw_metadata
		FROM resources
		WHERE assessment_id = ? AND resource_id = ?`, assessmentID, resourceID)
}

// scanResource scans the current row into a Resource struct.
func scanResource(row rowScanner) (Resource, error) {
	var r Resource
	var tagsJSON, metadataJSON string

	err := row.Scan(
		&r.ID,
		&r.AssessmentID,
		&r.ProviderType,
		&r.ResourceType,
		&r.ResourceID,
		&r.Region,
		&r.Name,
		&tagsJSON,
		&metadataJSON,
	)
	if err != nil {
		return Resource{}, err
	}

	if err := json.Unmarshal([]byte(tagsJSON), &r.Tags); err != nil {
		return Resource{}, fmt.Errorf("unmarshaling tags: %w", err)
	}
	if err := json.Unmarshal([]byte(metadataJSON), &r.RawMetadata); err != nil {
		return Resource{}, fmt.Errorf("unmarshaling raw_metadata: %w", err)
	}

	return r, nil
}
