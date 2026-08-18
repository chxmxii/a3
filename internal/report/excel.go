package report

import (
	"fmt"
	"sort"
	"strings"

	"github.com/xuri/excelize/v2"
	"github.com/chxmxii/a3/internal/storage"
)

// Color palette for a clean, professional look.
const (
	colorPrimary    = "#1B2A4A" // dark navy — title/header backgrounds
	colorAccent     = "#2E86AB" // teal accent — section headers
	colorLightGray  = "#F5F6F8" // alternating row fill
	colorWhite      = "#FFFFFF"
	colorBorder     = "#D0D5DD" // subtle borders
	colorCritical   = "#DC2626" // red
	colorHigh       = "#EA580C" // orange
	colorMedium     = "#CA8A04" // amber
	colorLow        = "#2563EB" // blue
	colorInfo       = "#6B7280" // gray
	colorGreen      = "#16A34A" // success/healthy
	colorCriticalBg = "#FEF2F2"
	colorHighBg     = "#FFF7ED"
	colorMediumBg   = "#FEFCE8"
	colorLowBg      = "#EFF6FF"
)

// RenderExcel generates a professionally styled Excel workbook from the report data.
func RenderExcel(data *ReportData, outputPath string) error {
	f := excelize.NewFile()
	defer f.Close()

	// Remove default "Sheet1".
	f.DeleteSheet("Sheet1")

	tech := BuildTechnicalReport(data)

	// Create styled sheets.
	writeSummarySheet(f, data, tech)
	writeInventorySheet(f, data)
	writeFindingsSheet(f, data)
	writeCostSheet(f, data)
	writeRelationshipsSheet(f, data)

	// Set active sheet to Summary.
	idx, _ := f.GetSheetIndex("Summary")
	f.SetActiveSheet(idx)

	if err := f.SaveAs(outputPath); err != nil {
		return fmt.Errorf("saving excel report: %w", err)
	}
	return nil
}

// -------------------------------------------------------------------
// Styles
// -------------------------------------------------------------------

func makeTitleStyle(f *excelize.File) int {
	s, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   18,
			Color:  colorPrimary,
			Family: "Segoe UI",
		},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	return s
}

func makeSubtitleStyle(f *excelize.File) int {
	s, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   11,
			Color:  "#6B7280",
			Family: "Segoe UI",
		},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	return s
}

func makeSectionHeaderStyle(f *excelize.File) int {
	s, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   12,
			Color:  colorAccent,
			Family: "Segoe UI",
		},
		Border: []excelize.Border{
			{Type: "bottom", Color: colorAccent, Style: 2},
		},
	})
	return s
}

func makeMetricLabelStyle(f *excelize.File) int {
	s, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   10,
			Color:  "#6B7280",
			Family: "Segoe UI",
		},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	return s
}

func makeMetricValueStyle(f *excelize.File) int {
	s, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   11,
			Color:  colorPrimary,
			Family: "Segoe UI",
		},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	return s
}

func makeTableHeaderStyle(f *excelize.File) int {
	s, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   10,
			Color:  colorWhite,
			Family: "Segoe UI",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{colorPrimary},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
		Border: []excelize.Border{
			{Type: "bottom", Color: colorPrimary, Style: 1},
		},
	})
	return s
}

func makeDataRowStyle(f *excelize.File, alt bool) int {
	fill := excelize.Fill{}
	if alt {
		fill = excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{colorLightGray}}
	}
	s, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   10,
			Family: "Segoe UI",
		},
		Fill: fill,
		Alignment: &excelize.Alignment{
			Vertical: "center",
			WrapText: true,
		},
		Border: []excelize.Border{
			{Type: "bottom", Color: colorBorder, Style: 1},
		},
	})
	return s
}

func makeSeverityStyle(f *excelize.File, severity string) int {
	fontColor := colorInfo
	bgColor := colorWhite
	switch strings.ToLower(severity) {
	case "critical":
		fontColor = colorCritical
		bgColor = colorCriticalBg
	case "high":
		fontColor = colorHigh
		bgColor = colorHighBg
	case "medium":
		fontColor = colorMedium
		bgColor = colorMediumBg
	case "low":
		fontColor = colorLow
		bgColor = colorLowBg
	}
	s, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   10,
			Color:  fontColor,
			Family: "Segoe UI",
		},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{bgColor}},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "bottom", Color: colorBorder, Style: 1},
		},
	})
	return s
}

func makeCurrencyStyle(f *excelize.File, alt bool) int {
	fill := excelize.Fill{}
	if alt {
		fill = excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{colorLightGray}}
	}
	s, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   10,
			Family: "Segoe UI",
		},
		Fill:      fill,
		NumFmt:    4, // #,##0.00
		Alignment: &excelize.Alignment{Vertical: "center", Horizontal: "right"},
		Border: []excelize.Border{
			{Type: "bottom", Color: colorBorder, Style: 1},
		},
	})
	return s
}

func makeRiskBadgeStyle(f *excelize.File, level string) int {
	fontColor := colorGreen
	bgColor := "#F0FDF4"
	switch level {
	case "CRITICAL":
		fontColor = colorCritical
		bgColor = colorCriticalBg
	case "HIGH":
		fontColor = colorHigh
		bgColor = colorHighBg
	case "MEDIUM":
		fontColor = colorMedium
		bgColor = colorMediumBg
	}
	s, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   12,
			Color:  fontColor,
			Family: "Segoe UI",
		},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{bgColor}},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	return s
}

// -------------------------------------------------------------------
// Summary Sheet
// -------------------------------------------------------------------

func writeSummarySheet(f *excelize.File, data *ReportData, tech TechnicalReport) {
	sheet := "Summary"
	f.NewSheet(sheet)
	exec := tech.Executive

	titleStyle := makeTitleStyle(f)
	subtitleStyle := makeSubtitleStyle(f)
	sectionStyle := makeSectionHeaderStyle(f)
	labelStyle := makeMetricLabelStyle(f)
	valueStyle := makeMetricValueStyle(f)

	f.SetColWidth(sheet, "A", "A", 3)  // left margin
	f.SetColWidth(sheet, "B", "B", 22) // labels
	f.SetColWidth(sheet, "C", "C", 30) // values
	f.SetColWidth(sheet, "D", "D", 5)  // spacer
	f.SetColWidth(sheet, "E", "E", 22) // labels
	f.SetColWidth(sheet, "F", "F", 30) // values

	row := 2
	// Title.
	f.SetCellValue(sheet, cell("B", row), "Cloud Assessment Report")
	f.SetCellStyle(sheet, cell("B", row), cell("B", row), titleStyle)
	row++
	if data.Assessment != nil {
		subtitle := fmt.Sprintf("%s  •  %s  •  %s",
			strings.ToUpper(data.Assessment.Provider),
			data.Assessment.Profile,
			data.Assessment.StartedAt.Format("Jan 2, 2006"))
		f.SetCellValue(sheet, cell("B", row), subtitle)
		f.SetCellStyle(sheet, cell("B", row), cell("B", row), subtitleStyle)
	}
	row += 2

	// Risk level badge.
	riskLevel := exec.RiskLevel()
	riskStyle := makeRiskBadgeStyle(f, riskLevel)
	f.SetCellValue(sheet, cell("B", row), "Overall Risk")
	f.SetCellStyle(sheet, cell("B", row), cell("B", row), labelStyle)
	f.SetCellValue(sheet, cell("C", row), riskLevel)
	f.SetCellStyle(sheet, cell("C", row), cell("C", row), riskStyle)
	row += 2

	// Key metrics section.
	f.SetCellValue(sheet, cell("B", row), "Key Metrics")
	f.SetCellStyle(sheet, cell("B", row), cell("F", row), sectionStyle)
	row++

	// Two-column metrics layout.
	writeMetric(f, sheet, row, "B", "C", "Total Resources", fmt.Sprintf("%d", exec.TotalResources), labelStyle, valueStyle)
	writeMetric(f, sheet, row, "E", "F", "Total Findings", fmt.Sprintf("%d", exec.TotalFindings), labelStyle, valueStyle)
	row++
	writeMetric(f, sheet, row, "B", "C", "Est. Monthly Cost", fmt.Sprintf("$%.2f", exec.MonthlyCost), labelStyle, valueStyle)
	writeMetric(f, sheet, row, "E", "F", "Relationships Mapped", fmt.Sprintf("%d", tech.Relationships), labelStyle, valueStyle)
	row += 2

	// Assessment info.
	if data.Assessment != nil {
		f.SetCellValue(sheet, cell("B", row), "Assessment Details")
		f.SetCellStyle(sheet, cell("B", row), cell("F", row), sectionStyle)
		row++
		writeMetric(f, sheet, row, "B", "C", "Assessment ID", data.Assessment.ID, labelStyle, valueStyle)
		row++
		writeMetric(f, sheet, row, "B", "C", "Status", strings.ToUpper(data.Assessment.Status), labelStyle, valueStyle)
		writeMetric(f, sheet, row, "E", "F", "Provider", strings.ToUpper(data.Assessment.Provider), labelStyle, valueStyle)
		row++
		writeMetric(f, sheet, row, "B", "C", "Started", data.Assessment.StartedAt.Format("2006-01-02 15:04:05"), labelStyle, valueStyle)
		if data.Assessment.CompletedAt != nil {
			writeMetric(f, sheet, row, "E", "F", "Completed", data.Assessment.CompletedAt.Format("2006-01-02 15:04:05"), labelStyle, valueStyle)
		}
		row++
		writeMetric(f, sheet, row, "B", "C", "Regions", strings.Join(data.Assessment.Regions, ", "), labelStyle, valueStyle)
		row += 2
	}

	// Findings by severity.
	f.SetCellValue(sheet, cell("B", row), "Findings by Severity")
	f.SetCellStyle(sheet, cell("B", row), cell("F", row), sectionStyle)
	row++

	sevCounts := map[string]int{
		"critical": exec.CriticalCount,
		"high":     exec.HighCount,
		"medium":   exec.MediumCount,
		"low":      exec.LowCount,
	}
	for _, finding := range data.Findings {
		if finding.Severity == "informational" {
			sevCounts["informational"]++
		}
	}
	for _, sev := range []string{"critical", "high", "medium", "low", "informational"} {
		if c := sevCounts[sev]; c > 0 {
			sevStyle := makeSeverityStyle(f, sev)
			f.SetCellValue(sheet, cell("B", row), strings.ToUpper(sev))
			f.SetCellStyle(sheet, cell("B", row), cell("B", row), sevStyle)
			f.SetCellValue(sheet, cell("C", row), c)
			f.SetCellStyle(sheet, cell("C", row), cell("C", row), valueStyle)
			row++
		}
	}
	row++

	// Resources by type.
	f.SetCellValue(sheet, cell("B", row), "Resources by Type")
	f.SetCellStyle(sheet, cell("B", row), cell("F", row), sectionStyle)
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
		f.SetCellValue(sheet, cell("B", row), item.k)
		f.SetCellStyle(sheet, cell("B", row), cell("B", row), labelStyle)
		f.SetCellValue(sheet, cell("C", row), item.v)
		f.SetCellStyle(sheet, cell("C", row), cell("C", row), valueStyle)
		row++
	}
	row++

	// Top risks.
	if len(exec.TopRisks) > 0 {
		f.SetCellValue(sheet, cell("B", row), "Top Risks")
		f.SetCellStyle(sheet, cell("B", row), cell("F", row), sectionStyle)
		row++
		for i, risk := range exec.TopRisks {
			f.SetCellValue(sheet, cell("B", row), fmt.Sprintf("%d.", i+1))
			f.SetCellStyle(sheet, cell("B", row), cell("B", row), valueStyle)
			f.SetCellValue(sheet, cell("C", row), risk)
			f.SetCellStyle(sheet, cell("C", row), cell("C", row), labelStyle)
			row++
		}
	}
}

func writeMetric(f *excelize.File, sheet string, row int, labelCol, valueCol, label, value string, labelStyle, valueStyle int) {
	f.SetCellValue(sheet, cell(labelCol, row), label)
	f.SetCellStyle(sheet, cell(labelCol, row), cell(labelCol, row), labelStyle)
	f.SetCellValue(sheet, cell(valueCol, row), value)
	f.SetCellStyle(sheet, cell(valueCol, row), cell(valueCol, row), valueStyle)
}

// -------------------------------------------------------------------
// Inventory Sheet
// -------------------------------------------------------------------

func writeInventorySheet(f *excelize.File, data *ReportData) {
	sheet := "Inventory"
	headers := []string{"Type", "Name", "Region", "Resource ID", "Tags"}
	widths := []float64{22, 35, 16, 55, 45}

	rows := make([][]any, 0, len(data.Resources))
	for _, r := range data.Resources {
		var tagParts []string
		for k, v := range r.Tags {
			tagParts = append(tagParts, k+"="+v)
		}
		sort.Strings(tagParts)
		rows = append(rows, []any{r.ResourceType, r.Name, r.Region, r.ResourceID, strings.Join(tagParts, "; ")})
	}

	writeStyledTable(f, sheet, headers, widths, rows, -1, -1)
}

// -------------------------------------------------------------------
// Findings Sheet
// -------------------------------------------------------------------

func writeFindingsSheet(f *excelize.File, data *ReportData) {
	sheet := "Findings"
	headers := []string{"Severity", "Category", "Resource ID", "Description", "Recommendation", "Standard", "Control"}
	widths := []float64{13, 22, 50, 55, 55, 22, 12}

	// Sort findings: critical first, then high, medium, low, informational.
	sortedFindings := make([]storage.Finding, len(data.Findings))
	copy(sortedFindings, data.Findings)
	sevOrder := map[string]int{"critical": 0, "high": 1, "medium": 2, "low": 3, "informational": 4}
	sort.SliceStable(sortedFindings, func(i, j int) bool {
		return sevOrder[sortedFindings[i].Severity] < sevOrder[sortedFindings[j].Severity]
	})

	rows := make([][]any, 0, len(sortedFindings))
	for _, finding := range sortedFindings {
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

	// Severity column index = 0 (first col has severity-specific styling).
	writeStyledTable(f, sheet, headers, widths, rows, 0, -1)
}

// -------------------------------------------------------------------
// Cost Sheet
// -------------------------------------------------------------------

func writeCostSheet(f *excelize.File, data *ReportData) {
	sheet := "Cost"
	headers := []string{"Resource ID", "Resource Type", "Category", "Monthly Cost ($)", "Confidence", "Idle", "Oversized"}
	widths := []float64{50, 22, 18, 16, 14, 8, 11}

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

	// Cost column index = 3 (for currency formatting).
	writeStyledTable(f, sheet, headers, widths, rows, -1, 3)
}

// -------------------------------------------------------------------
// Relationships Sheet
// -------------------------------------------------------------------

func writeRelationshipsSheet(f *excelize.File, data *ReportData) {
	sheet := "Architecture"
	headers := []string{"Source ID", "Target ID", "Relationship Type", "Status", "Reason"}
	widths := []float64{50, 50, 25, 14, 35}

	rows := make([][]any, 0, len(data.Relationships))
	for _, rel := range data.Relationships {
		rows = append(rows, []any{rel.SourceID, rel.TargetID, rel.RelationshipType, rel.Status, rel.UnresolvedReason})
	}

	writeStyledTable(f, sheet, headers, widths, rows, -1, -1)
}

// -------------------------------------------------------------------
// Core table writer with styling
// -------------------------------------------------------------------

// writeStyledTable creates a professional-looking table with:
// - Styled header row with dark background
// - Alternating row colors
// - Freeze panes on the header row
// - Auto-filter
// - Optional severity-colored column (severityCol index, -1 to skip)
// - Optional currency-formatted column (currencyCol index, -1 to skip)
func writeStyledTable(f *excelize.File, sheet string, headers []string, widths []float64, rows [][]any, severityCol, currencyCol int) {
	f.NewSheet(sheet)

	headerStyle := makeTableHeaderStyle(f)
	rowStyleEven := makeDataRowStyle(f, false)
	rowStyleOdd := makeDataRowStyle(f, true)

	// Set column widths.
	for i, w := range widths {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetColWidth(sheet, col, col, w)
	}

	// Write header.
	for i, h := range headers {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetCellValue(sheet, cell(col, 1), h)
		f.SetCellStyle(sheet, cell(col, 1), cell(col, 1), headerStyle)
	}
	f.SetRowHeight(sheet, 1, 24)

	// Write data rows.
	for i, vals := range rows {
		rowNum := i + 2
		alt := i%2 == 1
		baseStyle := rowStyleEven
		if alt {
			baseStyle = rowStyleOdd
		}

		for j, v := range vals {
			col, _ := excelize.ColumnNumberToName(j + 1)
			cellRef := cell(col, rowNum)

			f.SetCellValue(sheet, cellRef, v)

			// Apply severity styling to the severity column.
			if j == severityCol {
				if sev, ok := v.(string); ok {
					f.SetCellStyle(sheet, cellRef, cellRef, makeSeverityStyle(f, sev))
					continue
				}
			}

			// Apply currency formatting to the cost column.
			if j == currencyCol {
				if _, ok := v.(float64); ok {
					f.SetCellStyle(sheet, cellRef, cellRef, makeCurrencyStyle(f, alt))
					continue
				}
			}

			f.SetCellStyle(sheet, cellRef, cellRef, baseStyle)
		}
	}

	// Freeze the header row.
	f.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      0,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	})

	// Auto-filter.
	if len(rows) > 0 {
		lastCol, _ := excelize.ColumnNumberToName(len(headers))
		f.AutoFilter(sheet, fmt.Sprintf("A1:%s%d", lastCol, len(rows)+1), nil)
	}
}

// -------------------------------------------------------------------
// Helpers
// -------------------------------------------------------------------

func cell(col string, row int) string {
	return fmt.Sprintf("%s%d", col, row)
}

func boolToYesNo(b bool) string {
	if b {
		return "Yes"
	}
	return ""
}
