package storage

import (
	"encoding/json"
	"fmt"
)

// SizingEntry represents a sizing data record stored in SQLite.
type SizingEntry struct {
	ID           int64
	AssessmentID string
	Category     string         // "compute", "kubernetes", "database", "storage"
	ResourceID   string
	Data         map[string]any // Category-specific fields, serialized as JSON
}

// InsertSizing inserts a sizing entry into the database, JSON-serializing the Data field.
func (s *Store) InsertSizing(entry *SizingEntry) error {
	data := entry.Data
	if data == nil {
		data = map[string]any{}
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshaling sizing data: %w", err)
	}

	result, err := s.DB.Exec(`
		INSERT INTO sizing (assessment_id, category, resource_id, data)
		VALUES (?, ?, ?, ?)`,
		entry.AssessmentID,
		entry.Category,
		entry.ResourceID,
		string(dataJSON),
	)
	if err != nil {
		return fmt.Errorf("inserting sizing entry: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("getting last insert id: %w", err)
	}
	entry.ID = id

	return nil
}

// GetSizingByAssessment returns all sizing entries for the given assessment ID.
func (s *Store) GetSizingByAssessment(assessmentID string) ([]SizingEntry, error) {
	return queryAll(s.DB, "sizing by assessment", scanSizingEntry, `
		SELECT id, assessment_id, category, resource_id, data
		FROM sizing
		WHERE assessment_id = ?`, assessmentID)
}

// scanSizingEntry scans the current row into a SizingEntry struct.
func scanSizingEntry(row rowScanner) (SizingEntry, error) {
	var entry SizingEntry
	var dataJSON string

	err := row.Scan(
		&entry.ID,
		&entry.AssessmentID,
		&entry.Category,
		&entry.ResourceID,
		&dataJSON,
	)
	if err != nil {
		return SizingEntry{}, err
	}

	if err := json.Unmarshal([]byte(dataJSON), &entry.Data); err != nil {
		return SizingEntry{}, fmt.Errorf("unmarshaling sizing data: %w", err)
	}

	return entry, nil
}
