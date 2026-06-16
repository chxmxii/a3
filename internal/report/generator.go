package report

import (
	"fmt"

	"github.com/chxmxii/3a/internal/storage"
)

// Format specifies the output format for reports.
type Format string

const (
	FormatMarkdown Format = "markdown"
	FormatJSON     Format = "json"
	FormatExcel    Format = "excel"
)

// Generator orchestrates report generation.
type Generator struct {
	store *storage.Store
}

// NewGenerator creates a new report generator.
func NewGenerator(store *storage.Store) *Generator {
	return &Generator{store: store}
}

// Generate creates a report for the given assessment in the specified format.
func (g *Generator) Generate(assessmentID string, format Format) (string, error) {
	data, err := g.gatherData(assessmentID)
	if err != nil {
		return "", fmt.Errorf("gathering report data: %w", err)
	}

	switch format {
	case FormatMarkdown:
		return renderMarkdown(data), nil
	case FormatJSON:
		return renderJSON(data)
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}

// GenerateExcel creates an Excel report at the specified path.
func (g *Generator) GenerateExcel(assessmentID string, outputPath string) error {
	data, err := g.gatherData(assessmentID)
	if err != nil {
		return fmt.Errorf("gathering report data: %w", err)
	}
	return RenderExcel(data, outputPath)
}

// ReportData holds all data needed to render a report.
type ReportData struct {
	Assessment    *storage.Assessment
	Resources     []storage.Resource
	Findings      []storage.Finding
	Relationships []storage.Relationship
	Costs         []storage.CostEstimate
	Sizing        []storage.SizingEntry
}

func (g *Generator) gatherData(assessmentID string) (*ReportData, error) {
	assessment, err := g.store.GetAssessment(assessmentID)
	if err != nil {
		return nil, err
	}
	if assessment == nil {
		return nil, fmt.Errorf("assessment %s not found", assessmentID)
	}

	resources, err := g.store.GetResourcesByAssessment(assessmentID)
	if err != nil {
		return nil, err
	}

	findings, err := g.store.GetFindingsByAssessment(assessmentID)
	if err != nil {
		return nil, err
	}

	relationships, err := g.store.GetRelationshipsByAssessment(assessmentID)
	if err != nil {
		return nil, err
	}

	costs, err := g.store.GetCostsByAssessment(assessmentID)
	if err != nil {
		return nil, err
	}

	sizing, err := g.store.GetSizingByAssessment(assessmentID)
	if err != nil {
		return nil, err
	}

	return &ReportData{
		Assessment:    assessment,
		Resources:     resources,
		Findings:      findings,
		Relationships: relationships,
		Costs:         costs,
		Sizing:        sizing,
	}, nil
}
