package storage

import (
	"fmt"
)

// CostEstimate represents a cost estimate for a resource stored in SQLite.
type CostEstimate struct {
	ID            int64
	AssessmentID  string
	ResourceID    string
	ResourceType  string
	MonthlyCost   *float64 // nil if unestimable
	Confidence    *string  // nil if unestimable
	Category      string
	IdleFlag      bool
	OversizedFlag bool
	Unestimable   bool
}

// InsertCostEstimate inserts a cost estimate record into the database.
func (s *Store) InsertCostEstimate(est *CostEstimate) error {
	result, err := s.DB.Exec(`
		INSERT INTO cost_estimates (assessment_id, resource_id, resource_type, monthly_cost, confidence, category, idle_flag, oversized_flag, unestimable)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		est.AssessmentID,
		est.ResourceID,
		est.ResourceType,
		est.MonthlyCost,
		est.Confidence,
		est.Category,
		boolToInt(est.IdleFlag),
		boolToInt(est.OversizedFlag),
		boolToInt(est.Unestimable),
	)
	if err != nil {
		return fmt.Errorf("inserting cost estimate: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("getting last insert id: %w", err)
	}
	est.ID = id

	return nil
}

// GetCostsByAssessment returns all cost estimates for the given assessment ID.
func (s *Store) GetCostsByAssessment(assessmentID string) ([]CostEstimate, error) {
	return queryAll(s.DB, "costs by assessment", scanCostEstimate, `
		SELECT id, assessment_id, resource_id, resource_type, monthly_cost, confidence, category, idle_flag, oversized_flag, unestimable
		FROM cost_estimates
		WHERE assessment_id = ?`, assessmentID)
}

// scanCostEstimate scans the current row into a CostEstimate struct.
func scanCostEstimate(row rowScanner) (CostEstimate, error) {
	var est CostEstimate
	var idleFlag, oversizedFlag, unestimable int

	err := row.Scan(
		&est.ID,
		&est.AssessmentID,
		&est.ResourceID,
		&est.ResourceType,
		&est.MonthlyCost,
		&est.Confidence,
		&est.Category,
		&idleFlag,
		&oversizedFlag,
		&unestimable,
	)
	if err != nil {
		return CostEstimate{}, err
	}

	est.IdleFlag = idleFlag != 0
	est.OversizedFlag = oversizedFlag != 0
	est.Unestimable = unestimable != 0

	return est, nil
}

// boolToInt converts a bool to an integer (0 or 1) for SQLite storage.
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
