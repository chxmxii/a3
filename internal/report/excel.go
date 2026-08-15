package report

import (
	"fmt"
	"sort"
	"strings"

	"github.com/xuri/excelize/v2"
)

// RenderExcel generates an Excel workbook with multiple sheets from the report data.
func RenderExcel(data *ReportData, outputPath string) error {
	f := excelize.NewFile()
	defer f.Close()

	// Remove default "Sheet1".
	f.DeleteSheet("Sheet1")

	tech := BuildTechnicalReport(data)

	// Sheet 1: Summary.
	writeSummarySheet(f, data, tech)

	// Sheet 2: Inventory.
	writeInventorySheet(f, data)

	// Sheet 3: Findings.
	writeFindingsSheet(f, data)

	// Sheet 4: Cost.
	writeCostSheet(f, data)

	// Sheet 5: Relationships.
	writeRelationshipsSheet(f, data)

	// Save.
	if err := f.SaveAs(outputPath); err != nil {
		return fmt.Errorf("saving excel report: %w", err)
	}

	return nil
}

func writeSummarySheet(f *excelize.File, data *ReportData, tech TechnicalReport) {
	sheet := "Summary"
	f.NewSheet(sheet)
	exec := tech.Executive

	row := 1
	f.SetCellValue(sheet, cell("A", row), "3A Assessment Report")
	row++
	row++

	if data.Assessment != nil {
		f.SetCellValue(sheet, cell("A", row), "Profile")
		f.SetCellValue(sheet, cell("B", row), data.Assessment.Profile)
		row++
		f.SetCellValue(sheet, cell("A", row), "Provider")
		f.SetCellValue(sheet, cell("B", row), data.Assessment.Provider)
		row++
		f.SetCellValue(sheet, cell("A", row), "Status")
		f.SetCellValue(sheet, cell("B", row), data.Assessment.Status)
		row++
		f.SetCellValue(sheet, cell("A", row), "Started")
		f.SetCellValue(sheet, cell("B", row), data.Assessment.StartedAt.Format("2006-01-02 15:04:05"))
		row++
		if data.Assessment.CompletedAt != nil {
			f.SetCellValue(sheet, cell("A", row), "Completed")
			f.SetCellValue(sheet, cell("B", row), data.Assessment.CompletedAt.Format("2006-01-02 15:04:05"))
			row++
		}
		f.SetCellValue(sheet, cell("A", row), "Regions")
		f.SetCellValue(sheet, cell("B", row), strings.Join(data.Assessment.Regions, ", "))
		row++
	}

	row++
	f.SetCellValue(sheet, cell("A", row), "Total Resources")
	f.SetCellValue(sheet, cell("B", row), exec.TotalResources)
	row++
	f.SetCellValue(sheet, cell("A", row), "Total Findings")
	f.SetCellValue(sheet, cell("B", row), exec.TotalFindings)
	row++

	// Cost total.
	f.SetCellValue(sheet, cell("A", row), "Est. Monthly Cost")
	f.SetCellValue(sheet, cell("B", row), fmt.Sprintf("$%.2f", exec.MonthlyCost))
	row++
	row++

	// Findings by severity.
	f.SetCellValue(sheet, cell("A", row), "Findings by Severity")
	row++
	sevCounts := map[string]int{
		"critical": exec.CriticalCount,
		"high":     exec.HighCount,
		"medium":   exec.MediumCount,
		"low":      exec.LowCount,
	}
	// BuildExecutiveSummary does not count informational findings; derive that one here.
	for _, finding := range data.Findings {
		if finding.Severity == "informational" {
			sevCounts["informational"]++
		}
	}
	for _, sev := range []string{"critical", "high", "medium", "low", "informational"} {
		if c := sevCounts[sev]; c > 0 {
			f.SetCellValue(sheet, cell("A", row), strings.ToUpper(sev))
			f.SetCellValue(sheet, cell("B", row), c)
			row++
		}
	}

	row++
	// Resources by type.
	f.SetCellValue(sheet, cell("A", row), "Resources by Type")
	row++
	type kv struct {
		k string
		v int
	}
	var sorted []kv
	for k, v := range tech.ResourcesByType {
		sorted = append(sorted, kv{k, v})
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].v > sorted[j].v })
	for _, item := range sorted {
		f.SetCellValue(sheet, cell("A", row), item.k)
		f.SetCellValue(sheet, cell("B", row), item.v)
		row++
	}

	// Set column widths.
	f.SetColWidth(sheet, "A", "A", 25)
	f.SetColWidth(sheet, "B", "B", 40)
}

func writeInventorySheet(f *excelize.File, data *ReportData) {
	sheet := "Inventory"

	rows := make([][]any, 0, len(data.Resources))
	for _, r := range data.Resources {
		// Tags as key=value pairs.
		var tagParts []string
		for k, v := range r.Tags {
			tagParts = append(tagParts, k+"="+v)
		}
		sort.Strings(tagParts)
		rows = append(rows, []any{r.ResourceType, r.Name, r.Region, r.ResourceID, strings.Join(tagParts, "; ")})
	}

	writeTable(f, sheet,
		[]string{"Type", "Name", "Region", "Resource ID", "Tags"},
		[]float64{20, 35, 18, 50, 40},
		rows)

	// Auto-filter.
	if len(rows) > 0 {
		f.AutoFilter(sheet, fmt.Sprintf("A1:E%d", len(rows)+1), nil)
	}
}

func writeFindingsSheet(f *excelize.File, data *ReportData) {
	sheet := "Findings"

	rows := make([][]any, 0, len(data.Findings))
	for _, finding := range data.Findings {
		rows = append(rows, []any{
			strings.ToUpper(finding.Severity),
			finding.Category,
			finding.ResourceID,
			finding.Description,
			finding.Recommendation,
			finding.StandardName,
			finding.ControlID,
		})
	}

	writeTable(f, sheet,
		[]string{"Severity", "Category", "Resource ID", "Description", "Recommendation", "Standard", "Control"},
		[]float64{12, 20, 50, 60, 60, 25, 15},
		rows)

	if len(rows) > 0 {
		f.AutoFilter(sheet, fmt.Sprintf("A1:G%d", len(rows)+1), nil)
	}
}

func writeCostSheet(f *excelize.File, data *ReportData) {
	sheet := "Cost"

	rows := make([][]any, 0, len(data.Costs))
	for _, c := range data.Costs {
		var monthly any = "N/A"
		if c.MonthlyCost != nil {
			monthly = *c.MonthlyCost
		}
		conf := ""
		if c.Confidence != nil {
			conf = *c.Confidence
		}
		rows = append(rows, []any{
			c.ResourceID,
			c.ResourceType,
			c.Category,
			monthly,
			conf,
			boolToYesNo(c.IdleFlag),
			boolToYesNo(c.OversizedFlag),
		})
	}

	writeTable(f, sheet,
		[]string{"Resource ID", "Resource Type", "Category", "Monthly Cost ($)", "Confidence", "Idle", "Oversized"},
		[]float64{50, 20, 15, 15, 12, 8, 10},
		rows)

	if len(rows) > 0 {
		f.AutoFilter(sheet, fmt.Sprintf("A1:G%d", len(rows)+1), nil)
	}
}

func writeRelationshipsSheet(f *excelize.File, data *ReportData) {
	rows := make([][]any, 0, len(data.Relationships))
	for _, rel := range data.Relationships {
		rows = append(rows, []any{rel.SourceID, rel.TargetID, rel.RelationshipType, rel.Status, rel.UnresolvedReason})
	}

	writeTable(f, "Architecture",
		[]string{"Source ID", "Target ID", "Relationship Type", "Status", "Reason"},
		[]float64{50, 50, 25, 12, 30},
		rows)
}

// Helpers.

// writeTable creates a sheet and writes a header row, data rows, and column widths.
func writeTable(f *excelize.File, sheet string, headers []string, widths []float64, rows [][]any) {
	f.NewSheet(sheet)

	for i, h := range headers {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetCellValue(sheet, cell(col, 1), h)
	}

	for i, vals := range rows {
		row := i + 2
		for j, v := range vals {
			col, _ := excelize.ColumnNumberToName(j + 1)
			f.SetCellValue(sheet, cell(col, row), v)
		}
	}

	for i, w := range widths {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetColWidth(sheet, col, col, w)
	}
}

func cell(col string, row int) string {
	return fmt.Sprintf("%s%d", col, row)
}

func boolToYesNo(b bool) string {
	if b {
		return "Yes"
	}
	return ""
}
