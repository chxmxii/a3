package checklist

import (
	"fmt"

	"github.com/chxmxii/a3/internal/assessment"
	"github.com/chxmxii/a3/internal/storage"
)

// Engine generates a compliance checklist from assessment findings.
type Engine struct {
	store *storage.Store
}

// NewEngine creates a new checklist engine.
func NewEngine(store *storage.Store) *Engine {
	return &Engine{store: store}
}

// Generate creates a checklist summary from findings in the given assessment.
func (e *Engine) Generate(assessmentID string) (*ChecklistSummary, error) {
	findings, err := e.store.GetFindingsByAssessment(assessmentID)
	if err != nil {
		return nil, fmt.Errorf("loading findings: %w", err)
	}

	// Group findings by control_id.
	byControl := make(map[string][]storage.Finding)
	for _, f := range findings {
		key := f.ControlID + ": " + f.StandardName
		byControl[key] = append(byControl[key], f)
	}

	summary := &ChecklistSummary{
		ByCategory: make(map[string][]CheckItem),
	}

	// Define all checks.
	checks := allChecks()

	for _, check := range checks {
		item := CheckItem{
			Name:        check.Name,
			Description: check.Description,
			Category:    check.Category,
			Status:      StatusPass, // default to pass
		}

		// Look for findings matching this check.
		for key, controlFindings := range byControl {
			if matchesCheck(key, check) {
				// Determine status based on severity.
				for _, f := range controlFindings {
					item.ResourceIDs = append(item.ResourceIDs, f.ResourceID)
					switch f.Severity {
					case "critical", "high":
						item.Status = StatusFail
					case "medium":
						if item.Status != StatusFail {
							item.Status = StatusWarn
						}
					case "low", "informational":
						if item.Status == StatusPass {
							item.Status = StatusWarn
						}
					}
				}
				item.Details = fmt.Sprintf("%d resource(s) affected", len(controlFindings))
			}
		}

		summary.Items = append(summary.Items, item)
		summary.ByCategory[item.Category] = append(summary.ByCategory[item.Category], item)

		switch item.Status {
		case StatusPass:
			summary.PassCount++
		case StatusFail:
			summary.FailCount++
		case StatusWarn:
			summary.WarnCount++
		}
	}

	return summary, nil
}

type checkDef struct {
	Name        string
	Description string
	Category    string
	ControlIDs  []string
}

func matchesCheck(key string, check checkDef) bool {
	for _, id := range check.ControlIDs {
		if len(key) >= len(id) && key[:len(id)] == id {
			return true
		}
	}
	return false
}

// allChecks derives the checklist from the assessment standards catalog, so
// every control referenced by a rule appears in the generated checklist.
func allChecks() []checkDef {
	var checks []checkDef
	for _, std := range assessment.BuiltInStandards() {
		for _, control := range std.Controls {
			checks = append(checks, checkDef{
				Name:        control.Name,
				Description: control.Description,
				Category:    string(control.Category),
				ControlIDs:  []string{control.ID},
			})
		}
	}
	return checks
}
