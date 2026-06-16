package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/chxmxii/3a/internal/storage"
)

type overviewView struct {
	assessment *storage.Assessment
	resources  []storage.Resource
	findings   []storage.Finding
	costs      []storage.CostEstimate
}

func (v *overviewView) render(width, height int) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  📊 Assessment Overview"))
	b.WriteString("\n\n")

	if v.assessment != nil {
		b.WriteString(headerStyle.Render("  Assessment"))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("    Profile:   %s\n", v.assessment.Profile))
		b.WriteString(fmt.Sprintf("    Provider:  %s\n", v.assessment.Provider))
		b.WriteString(fmt.Sprintf("    Status:    %s\n", v.assessment.Status))
		b.WriteString(fmt.Sprintf("    Started:   %s\n", v.assessment.StartedAt.Format("2006-01-02 15:04:05")))
		if v.assessment.CompletedAt != nil {
			b.WriteString(fmt.Sprintf("    Completed: %s\n", v.assessment.CompletedAt.Format("2006-01-02 15:04:05")))
		}
		b.WriteString(fmt.Sprintf("    Regions:   %s\n", strings.Join(v.assessment.Regions, ", ")))
		b.WriteString("\n")
	}

	// Resources by type.
	b.WriteString(headerStyle.Render(fmt.Sprintf("  Resources (%d total)", len(v.resources))))
	b.WriteString("\n")
	typeCounts := make(map[string]int)
	regionCounts := make(map[string]int)
	for _, r := range v.resources {
		typeCounts[r.ResourceType]++
		regionCounts[r.Region]++
	}

	// Sort by count descending.
	type kv struct {
		key   string
		count int
	}
	var sorted []kv
	for k, c := range typeCounts {
		sorted = append(sorted, kv{k, c})
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].count > sorted[j].count })

	for _, item := range sorted {
		b.WriteString(fmt.Sprintf("    %-22s %d\n", item.key, item.count))
	}
	b.WriteString("\n")

	// Resources by region.
	b.WriteString(headerStyle.Render(fmt.Sprintf("  Regions (%d)", len(regionCounts))))
	b.WriteString("\n")
	var regionSorted []kv
	for k, c := range regionCounts {
		regionSorted = append(regionSorted, kv{k, c})
	}
	sort.Slice(regionSorted, func(i, j int) bool { return regionSorted[i].count > regionSorted[j].count })
	for _, item := range regionSorted {
		b.WriteString(fmt.Sprintf("    %-22s %d resources\n", item.key, item.count))
	}
	b.WriteString("\n")

	// Findings.
	b.WriteString(headerStyle.Render(fmt.Sprintf("  Findings (%d total)", len(v.findings))))
	b.WriteString("\n")
	if len(v.findings) == 0 {
		b.WriteString(passStyle.Render("    No findings — all checks passed"))
		b.WriteString("\n")
	} else {
		sevCounts := map[string]int{}
		for _, f := range v.findings {
			sevCounts[f.Severity]++
		}
		if c := sevCounts["critical"]; c > 0 {
			b.WriteString(fmt.Sprintf("    %s %d\n", severityCriticalStyle.Render("CRITICAL"), c))
		}
		if c := sevCounts["high"]; c > 0 {
			b.WriteString(fmt.Sprintf("    %s     %d\n", severityHighStyle.Render("HIGH"), c))
		}
		if c := sevCounts["medium"]; c > 0 {
			b.WriteString(fmt.Sprintf("    %s   %d\n", severityMediumStyle.Render("MEDIUM"), c))
		}
		if c := sevCounts["low"]; c > 0 {
			b.WriteString(fmt.Sprintf("    %s      %d\n", severityLowStyle.Render("LOW"), c))
		}
	}
	b.WriteString("\n")

	// Cost.
	b.WriteString(headerStyle.Render("  Monthly Cost Estimate"))
	b.WriteString("\n")
	totalCost := 0.0
	for _, c := range v.costs {
		if c.MonthlyCost != nil {
			totalCost += *c.MonthlyCost
		}
	}
	b.WriteString(fmt.Sprintf("    $%.2f/month (~$%.2f/year)\n", totalCost, totalCost*12))

	return b.String()
}
