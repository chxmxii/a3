package oci

import (
	"context"

	"github.com/chxmxii/a3/internal/assessment"
	"github.com/chxmxii/a3/internal/provider"
	"github.com/chxmxii/a3/internal/storage"
)

// newFinding builds a Finding for rule r, copying the rule's standard,
// control ID, and category so call sites don't repeat them.
func newFinding(r assessment.Rule, severity assessment.Severity, resource storage.Resource, description, recommendation string) assessment.Finding {
	return assessment.Finding{
		Severity:       severity,
		ResourceID:     resource.ResourceID,
		Description:    description,
		Recommendation: recommendation,
		StandardName:   r.Standard(),
		ControlID:      r.ControlID(),
		Category:       r.Category(),
	}
}

// simpleRule is a data-driven assessment rule: it checks resource metadata
// with violated and, on violation, emits a single finding with a fixed
// severity and recommendation and a description computed by describe.
type simpleRule struct {
	id             string
	standard       string
	controlID      string
	category       assessment.FindingCategory
	severity       assessment.Severity
	appliesTo      []provider.ResourceType
	recommendation string
	describe       func(storage.Resource) string
	violated       func(meta map[string]any) bool
}

func (r *simpleRule) ID() string                           { return r.id }
func (r *simpleRule) Standard() string                     { return r.standard }
func (r *simpleRule) ControlID() string                    { return r.controlID }
func (r *simpleRule) Category() assessment.FindingCategory { return r.category }
func (r *simpleRule) AppliesTo() []provider.ResourceType   { return r.appliesTo }

func (r *simpleRule) Evaluate(_ context.Context, resource storage.Resource) ([]assessment.Finding, error) {
	if r.violated(resource.RawMetadata) {
		return []assessment.Finding{newFinding(r, r.severity, resource, r.describe(resource), r.recommendation)}, nil
	}
	return nil, nil
}
