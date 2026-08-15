package storage

import (
	"database/sql"
	"fmt"
)

// rowScanner is the subset of *sql.Rows needed by the shared query helpers
// and the per-entity scan functions. Scan functions receive it positioned on
// the current row and must only call Scan.
type rowScanner interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}

// queryAll runs the query, scans every row via scan, and returns the results.
// op names the operation for error messages (e.g. "findings by assessment").
func queryAll[T any](db *sql.DB, op string, scan func(rowScanner) (T, error), query string, args ...any) ([]T, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying %s: %w", op, err)
	}
	defer rows.Close()

	var results []T
	for rows.Next() {
		item, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning %s row: %w", op, err)
		}
		results = append(results, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating %s rows: %w", op, err)
	}
	return results, nil
}

// queryOne runs the query and scans at most one row via scan.
// It returns (nil, nil) when the query matches no rows.
func queryOne[T any](db *sql.DB, op string, scan func(rowScanner) (T, error), query string, args ...any) (*T, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying %s: %w", op, err)
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("querying %s: %w", op, err)
		}
		return nil, nil
	}
	item, err := scan(rows)
	if err != nil {
		return nil, fmt.Errorf("scanning %s: %w", op, err)
	}
	return &item, nil
}
