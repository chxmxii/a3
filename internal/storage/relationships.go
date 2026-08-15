package storage

import "fmt"

// Relationship represents a directed edge between two resources.
type Relationship struct {
	ID               int64
	AssessmentID     string
	SourceID         string
	TargetID         string
	RelationshipType string
	Status           string // "resolved" or "unresolved"
	UnresolvedReason string
	TargetRegion     string
	TargetAccount    string
}

// InsertRelationship inserts a relationship record into the database.
func (s *Store) InsertRelationship(rel *Relationship) error {
	result, err := s.DB.Exec(`
		INSERT INTO relationships (assessment_id, source_id, target_id, relationship_type, status, unresolved_reason, target_region, target_account)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		rel.AssessmentID,
		rel.SourceID,
		rel.TargetID,
		rel.RelationshipType,
		rel.Status,
		rel.UnresolvedReason,
		rel.TargetRegion,
		rel.TargetAccount,
	)
	if err != nil {
		return fmt.Errorf("inserting relationship: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("getting last insert id: %w", err)
	}
	rel.ID = id
	return nil
}

// GetRelationshipsByAssessment returns all relationships for a given assessment.
func (s *Store) GetRelationshipsByAssessment(assessmentID string) ([]Relationship, error) {
	return queryAll(s.DB, "relationships by assessment", scanRelationship, `
		SELECT id, assessment_id, source_id, target_id, relationship_type, status, unresolved_reason, target_region, target_account
		FROM relationships
		WHERE assessment_id = ?`, assessmentID)
}

// scanRelationship scans the current row into a Relationship struct.
func scanRelationship(row rowScanner) (Relationship, error) {
	var r Relationship
	if err := row.Scan(
		&r.ID,
		&r.AssessmentID,
		&r.SourceID,
		&r.TargetID,
		&r.RelationshipType,
		&r.Status,
		&r.UnresolvedReason,
		&r.TargetRegion,
		&r.TargetAccount,
	); err != nil {
		return Relationship{}, err
	}
	return r, nil
}
