package storage

import "fmt"

// Finding represents a standards-based assessment result.
type Finding struct {
	ID             int64
	AssessmentID   string
	Severity       string
	ResourceID     string
	Description    string
	Recommendation string
	StandardName   string
	ControlID      string
	Category       string
}

// InsertFinding inserts a new finding into the database.
func (s *Store) InsertFinding(finding *Finding) error {
	result, err := s.DB.Exec(`
		INSERT INTO findings (assessment_id, severity, resource_id, description, recommendation, standard_name, control_id, category)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		finding.AssessmentID,
		finding.Severity,
		finding.ResourceID,
		finding.Description,
		finding.Recommendation,
		finding.StandardName,
		finding.ControlID,
		finding.Category,
	)
	if err != nil {
		return fmt.Errorf("inserting finding: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("getting last insert id: %w", err)
	}
	finding.ID = id
	return nil
}

// GetFindingsByAssessment returns all findings for a given assessment ID.
func (s *Store) GetFindingsByAssessment(assessmentID string) ([]Finding, error) {
	return queryAll(s.DB, "findings by assessment", scanFinding, `
		SELECT id, assessment_id, severity, resource_id, description, recommendation, standard_name, control_id, category
		FROM findings
		WHERE assessment_id = ?`, assessmentID)
}

// scanFinding scans the current row into a Finding struct.
func scanFinding(row rowScanner) (Finding, error) {
	var f Finding
	if err := row.Scan(&f.ID, &f.AssessmentID, &f.Severity, &f.ResourceID, &f.Description, &f.Recommendation, &f.StandardName, &f.ControlID, &f.Category); err != nil {
		return Finding{}, err
	}
	return f, nil
}
