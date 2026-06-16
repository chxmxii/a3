package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/chxmxii/3a/internal/storage"
)

type costView struct {
	costs        []storage.CostEstimate
	resources    []storage.Resource
	scrollOffset int
}

func (v *costView) render(width, height int) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  💰 Cost Analysis"))
	b.WriteString("\n\n")

	if len(v.costs) == 0 {
		b.WriteString(normalStyle.Render("  No cost estimates available."))
		return b.String()
	}

	lines := v.buildLines()

	// Apply scroll.
	maxRows := height - 6
	if maxRows < 10 {
		maxRows = 10
	}
	if v.scrollOffset > len(lines)-maxRows {
		v.scrollOffset = max(0, len(lines)-maxRows)
	}
	end := v.scrollOffset + maxRows
	if end > len(lines) {
		end = len(lines)
	}

	for i := v.scrollOffset; i < end; i++ {
		b.WriteString(lines[i])
		b.WriteString("\n")
	}

	if len(lines) > maxRows {
		pct := 0
		if len(lines)-maxRows > 0 {
			pct = (v.scrollOffset * 100) / (len(lines) - maxRows)
		}
		b.WriteString(dimNavStyle.Render(fmt.Sprintf("\n  ↕ scroll %d%%", pct)))
	}

	return b.String()
}

func (v *costView) buildLines() []string {
	var lines []string

	totalCost := 0.0
	byCategory := make(map[string]float64)
	var idleResources []storage.CostEstimate
	var oversizedResources []storage.CostEstimate

	for _, c := range v.costs {
		if c.MonthlyCost != nil {
			totalCost += *c.MonthlyCost
			byCategory[c.Category] += *c.MonthlyCost
		}
		if c.IdleFlag {
			idleResources = append(idleResources, c)
		}
		if c.OversizedFlag {
			oversizedResources = append(oversizedResources, c)
		}
	}

	// Total.
	lines = append(lines, headerStyle.Render("  Monthly Total"))
	lines = append(lines, fmt.Sprintf("    $%.2f/month (~$%.2f/year)", totalCost, totalCost*12))
	lines = append(lines, "")

	// By category.
	lines = append(lines, headerStyle.Render("  Breakdown by Category"))
	type catCost struct {
		name string
		cost float64
	}
	var cats []catCost
	for k, c := range byCategory {
		cats = append(cats, catCost{k, c})
	}
	sort.Slice(cats, func(i, j int) bool { return cats[i].cost > cats[j].cost })
	for _, cat := range cats {
		pct := 0.0
		if totalCost > 0 {
			pct = (cat.cost / totalCost) * 100
		}
		bar := strings.Repeat("█", int(pct/5))
		lines = append(lines, fmt.Sprintf("    %-12s $%8.2f  %4.1f%%  %s", cat.name, cat.cost, pct, dimNavStyle.Render(bar)))
	}
	lines = append(lines, "")

	// Top cost drivers.
	lines = append(lines, headerStyle.Render("  Top Cost Drivers"))
	nameMap := make(map[string]string)
	for _, r := range v.resources {
		nameMap[r.ResourceID] = r.Name
	}

	type costItem struct {
		id      string
		cost    float64
		resType string
	}
	var items []costItem
	for _, c := range v.costs {
		if c.MonthlyCost != nil && *c.MonthlyCost > 0 {
			items = append(items, costItem{c.ResourceID, *c.MonthlyCost, c.ResourceType})
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].cost > items[j].cost })
	if len(items) > 10 {
		items = items[:10]
	}
	for i, item := range items {
		name := nameMap[item.id]
		if name == "" {
			name = item.id
		}
		if len(name) > 30 {
			name = name[:27] + "..."
		}
		lines = append(lines, fmt.Sprintf("    %2d. %-30s %-14s $%.2f/mo", i+1, name, item.resType, item.cost))
	}
	lines = append(lines, "")

	// Optimization.
	if len(idleResources) > 0 || len(oversizedResources) > 0 {
		lines = append(lines, headerStyle.Render("  Optimization Opportunities"))
		if len(idleResources) > 0 {
			lines = append(lines, warnStyle.Render(fmt.Sprintf("    ⚠ %d potentially idle resource(s)", len(idleResources))))
		}
		if len(oversizedResources) > 0 {
			lines = append(lines, warnStyle.Render(fmt.Sprintf("    ⚠ %d potentially oversized resource(s)", len(oversizedResources))))
		}
	}

	return lines
}
