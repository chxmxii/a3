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

	// Sheet 1: Summary.
	writeSummarySheet(f, data)

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

func writeSummarySheet(f *excelize.File, data *ReportData) {
	sheet := "Summary"
	f.NewSheet(sheet)

	// Header styles.
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#4F46E5"}},
		Alignment: &excelize.Alignment{Horizontal: "left"},
	})
	_ = headerStyle

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
	f.SetCellValue(sheet, cell("B", row), len(data.Resources))
	row++
	f.SetCellValue(sheet, cell("A", row), "Total Findings")
	f.SetCellValue(sheet, cell("B", row), len(data.Findings))
	row++

	// Cost total.
	totalCost := 0.0
	for _, c := range data.Costs {
		if c.MonthlyCost != nil {
			totalCost += *c.MonthlyCost
		}
	}
	f.SetCellValue(sheet, cell("A", row), "Est. Monthly Cost")
	f.SetCellValue(sheet, cell("B", row), fmt.Sprintf("$%.2f", totalCost))
	row++
	row++

	// Findings by severity.
	f.SetCellValue(sheet, cell("A", row), "Findings by Severity")
	row++
	sevCounts := map[string]int{}
	for _, finding := range data.Findings {
		sevCounts[finding.Severity]++
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
	typeCounts := map[string]int{}
	for _, r := range data.Resources {
		typeCounts[r.ResourceType]++
	}
	type kv struct {
		k string
		v int
	}
	var sorted []kv
	for k, v := range typeCounts {
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
	f.NewSheet(sheet)

	// Headers.
	headers := []string{"Type", "Name", "Region", "Resource ID", "Tags"}
	for i, h := range headers {
		f.SetCellValue(sheet, cell(colLetter(i), 1), h)
	}

	// Data rows.
	for i, r := range data.Resources {
		row := i + 2
		f.SetCellValue(sheet, cell("A", row), r.ResourceType)
		f.SetCellValue(sheet, cell("B", row), r.Name)
		f.SetCellValue(sheet, cell("C", row), r.Region)
		f.SetCellValue(sheet, cell("D", row), r.ResourceID)

		// Tags as key=value pairs.
		var tagParts []string
		for k, v := range r.Tags {
			tagParts = append(tagParts, k+"="+v)
		}
		sort.Strings(tagParts)
		f.SetCellValue(sheet, cell("E", row), strings.Join(tagParts, "; "))
	}

	f.SetColWidth(sheet, "A", "A", 20)
	f.SetColWidth(sheet, "B", "B", 35)
	f.SetColWidth(sheet, "C", "C", 18)
	f.SetColWidth(sheet, "D", "D", 50)
	f.SetColWidth(sheet, "E", "E", 40)

	// Auto-filter.
	if len(data.Resources) > 0 {
		f.AutoFilter(sheet, fmt.Sprintf("A1:E%d", len(data.Resources)+1), nil)
	}
}

func writeFindingsSheet(f *excelize.File, data *ReportData) {
	sheet := "Findings"
	f.NewSheet(sheet)

	headers := []string{"Severity", "Category", "Resource ID", "Description", "Recommendation", "Standard", "Control"}
	for i, h := range headers {
		f.SetCellValue(sheet, cell(colLetter(i), 1), h)
	}

	for i, finding := range data.Findings {
		row := i + 2
		f.SetCellValue(sheet, cell("A", row), strings.ToUpper(finding.Severity))
		f.SetCellValue(sheet, cell("B", row), finding.Category)
		f.SetCellValue(sheet, cell("C", row), finding.ResourceID)
		f.SetCellValue(sheet, cell("D", row), finding.Description)
		f.SetCellValue(sheet, cell("E", row), finding.Recommendation)
		f.SetCellValue(sheet, cell("F", row), finding.StandardName)
		f.SetCellValue(sheet, cell("G", row), finding.ControlID)
	}

	f.SetColWidth(sheet, "A", "A", 12)
	f.SetColWidth(sheet, "B", "B", 20)
	f.SetColWidth(sheet, "C", "C", 50)
	f.SetColWidth(sheet, "D", "D", 60)
	f.SetColWidth(sheet, "E", "E", 60)
	f.SetColWidth(sheet, "F", "F", 25)
	f.SetColWidth(sheet, "G", "G", 15)

	if len(data.Findings) > 0 {
		f.AutoFilter(sheet, fmt.Sprintf("A1:G%d", len(data.Findings)+1), nil)
	}
}

func writeCostSheet(f *excelize.File, data *ReportData) {
	sheet := "Cost"
	f.NewSheet(sheet)

	headers := []string{"Resource ID", "Resource Type", "Category", "Monthly Cost ($)", "Confidence", "Idle", "Oversized"}
	for i, h := range headers {
		f.SetCellValue(sheet, cell(colLetter(i), 1), h)
	}

	for i, c := range data.Costs {
		row := i + 2
		f.SetCellValue(sheet, cell("A", row), c.ResourceID)
		f.SetCellValue(sheet, cell("B", row), c.ResourceType)
		f.SetCellValue(sheet, cell("C", row), c.Category)
		if c.MonthlyCost != nil {
			f.SetCellValue(sheet, cell("D", row), *c.MonthlyCost)
		} else {
			f.SetCellValue(sheet, cell("D", row), "N/A")
		}
		conf := ""
		if c.Confidence != nil {
			conf = *c.Confidence
		}
		f.SetCellValue(sheet, cell("E", row), conf)
		f.SetCellValue(sheet, cell("F", row), boolToYesNo(c.IdleFlag))
		f.SetCellValue(sheet, cell("G", row), boolToYesNo(c.OversizedFlag))
	}

	f.SetColWidth(sheet, "A", "A", 50)
	f.SetColWidth(sheet, "B", "B", 20)
	f.SetColWidth(sheet, "C", "C", 15)
	f.SetColWidth(sheet, "D", "D", 15)
	f.SetColWidth(sheet, "E", "E", 12)
	f.SetColWidth(sheet, "F", "F", 8)
	f.SetColWidth(sheet, "G", "G", 10)

	if len(data.Costs) > 0 {
		f.AutoFilter(sheet, fmt.Sprintf("A1:G%d", len(data.Costs)+1), nil)
	}
}

func writeRelationshipsSheet(f *excelize.File, data *ReportData) {
	sheet := "Architecture"
	f.NewSheet(sheet)

	headers := []string{"Source ID", "Target ID", "Relationship Type", "Status", "Reason"}
	for i, h := range headers {
		f.SetCellValue(sheet, cell(colLetter(i), 1), h)
	}

	for i, rel := range data.Relationships {
		row := i + 2
		f.SetCellValue(sheet, cell("A", row), rel.SourceID)
		f.SetCellValue(sheet, cell("B", row), rel.TargetID)
		f.SetCellValue(sheet, cell("C", row), rel.RelationshipType)
		f.SetCellValue(sheet, cell("D", row), rel.Status)
		f.SetCellValue(sheet, cell("E", row), rel.UnresolvedReason)
	}

	f.SetColWidth(sheet, "A", "A", 50)
	f.SetColWidth(sheet, "B", "B", 50)
	f.SetColWidth(sheet, "C", "C", 25)
	f.SetColWidth(sheet, "D", "D", 12)
	f.SetColWidth(sheet, "E", "E", 30)
}

// Helpers.

func cell(col string, row int) string {
	return fmt.Sprintf("%s%d", col, row)
}

func colLetter(idx int) string {
	return string(rune('A' + idx))
}

func boolToYesNo(b bool) string {
	if b {
		return "Yes"
	}
	return ""
}
