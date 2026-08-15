package report


// ExecutiveSummary generates a high-level summary suitable for leadership.
type ExecutiveSummary struct {
	Provider       string
	Profile        string
	TotalResources int
	TotalFindings  int
	CriticalCount  int
	HighCount      int
	MediumCount    int
	LowCount       int
	MonthlyCost    float64
	TopRisks       []string
}

// BuildExecutiveSummary creates an executive summary from report data.
func BuildExecutiveSummary(data *ReportData) ExecutiveSummary {
	summary := ExecutiveSummary{
		Provider:       data.Assessment.Provider,
		Profile:        data.Assessment.Profile,
		TotalResources: len(data.Resources),
		TotalFindings:  len(data.Findings),
	}

	for _, f := range data.Findings {
		switch f.Severity {
		case "critical":
			summary.CriticalCount++
		case "high":
			summary.HighCount++
		case "medium":
			summary.MediumCount++
		case "low":
			summary.LowCount++
		}
	}

	for _, c := range data.Costs {
		if c.MonthlyCost != nil {
			summary.MonthlyCost += *c.MonthlyCost
		}
	}

	// Collect unique top risks (critical/high findings).
	seen := make(map[string]bool)
	for _, f := range data.Findings {
		if (f.Severity == "critical" || f.Severity == "high") && !seen[f.Description] {
			seen[f.Description] = true
			desc := f.Description
			if len(desc) > 100 {
				desc = desc[:97] + "..."
			}
			summary.TopRisks = append(summary.TopRisks, desc)
			if len(summary.TopRisks) >= 5 {
				break
			}
		}
	}

	return summary
}

// RiskLevel returns a human-readable risk assessment.
func (s ExecutiveSummary) RiskLevel() string {
	if s.CriticalCount > 0 {
		return "CRITICAL"
	}
	if s.HighCount > 3 {
		return "HIGH"
	}
	if s.HighCount > 0 || s.MediumCount > 5 {
		return "MEDIUM"
	}
	return "LOW"
}
