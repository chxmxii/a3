package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/chxmxii/3a/internal/storage"
)

type inventoryView struct {
	resources    []storage.Resource
	regions      []string
	regionIdx    int    // -1 means "all regions"
	typeFilter   string // empty means all types
	typeIdx      int
	cursor       int
	offset       int
}

func (v *inventoryView) nextRegion() {
	v.regionIdx++
	if v.regionIdx >= len(v.regions) {
		v.regionIdx = -1 // wrap to "all"
	}
	v.cursor = 0
	v.offset = 0
}

func (v *inventoryView) prevRegion() {
	v.regionIdx--
	if v.regionIdx < -1 {
		v.regionIdx = len(v.regions) - 1
	}
	v.cursor = 0
	v.offset = 0
}

func (v *inventoryView) nextType() {
	types := v.availableTypes()
	v.typeIdx++
	if v.typeIdx >= len(types) {
		v.typeIdx = -1 // wrap to "all"
		v.typeFilter = ""
	} else {
		v.typeFilter = types[v.typeIdx]
	}
	v.cursor = 0
	v.offset = 0
}

func (v *inventoryView) clearFilters() {
	v.regionIdx = -1
	v.typeFilter = ""
	v.typeIdx = -1
	v.cursor = 0
	v.offset = 0
}

func (v *inventoryView) availableTypes() []string {
	typeSet := make(map[string]bool)
	for _, r := range v.resources {
		typeSet[r.ResourceType] = true
	}
	var types []string
	for t := range typeSet {
		types = append(types, t)
	}
	sort.Strings(types)
	return types
}

func (v *inventoryView) currentRegion() string {
	if v.regionIdx < 0 || v.regionIdx >= len(v.regions) {
		return ""
	}
	return v.regions[v.regionIdx]
}

func (v *inventoryView) filteredResources() []storage.Resource {
	region := v.currentRegion()
	var filtered []storage.Resource
	for _, r := range v.resources {
		if region != "" && r.Region != region {
			continue
		}
		if v.typeFilter != "" && r.ResourceType != v.typeFilter {
			continue
		}
		filtered = append(filtered, r)
	}
	return filtered
}

func (v *inventoryView) render(width, height int) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  📦 Resource Inventory"))
	b.WriteString("\n\n")

	// Filter status bar.
	region := v.currentRegion()
	filterParts := []string{}
	if region != "" {
		filterParts = append(filterParts, regionBadgeStyle.Render("Region: "+region))
	} else {
		filterParts = append(filterParts, dimNavStyle.Render("Region: all"))
	}
	if v.typeFilter != "" {
		filterParts = append(filterParts, regionBadgeStyle.Render("Type: "+v.typeFilter))
	} else {
		filterParts = append(filterParts, dimNavStyle.Render("Type: all"))
	}
	b.WriteString("  " + strings.Join(filterParts, "  "))
	b.WriteString("\n")

	filtered := v.filteredResources()
	b.WriteString(dimNavStyle.Render(fmt.Sprintf("  %d of %d resources", len(filtered), len(v.resources))))
	b.WriteString("\n\n")

	// Table header.
	header := fmt.Sprintf("  %-18s %-35s %-18s %s", "TYPE", "NAME", "REGION", "ID")
	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n")
	b.WriteString(dimNavStyle.Render("  " + strings.Repeat("─", min(width-4, 95))))
	b.WriteString("\n")

	// Visible rows.
	maxRows := height - 11
	if maxRows < 5 {
		maxRows = 5
	}

	// Adjust offset.
	if v.cursor < v.offset {
		v.offset = v.cursor
	}
	if v.cursor >= v.offset+maxRows {
		v.offset = v.cursor - maxRows + 1
	}
	if v.offset < 0 {
		v.offset = 0
	}

	end := v.offset + maxRows
	if end > len(filtered) {
		end = len(filtered)
	}

	for i := v.offset; i < end; i++ {
		r := filtered[i]
		shortID := r.ResourceID
		if len(shortID) > 20 {
			shortID = "..." + shortID[len(shortID)-17:]
		}
		name := r.Name
		if name == "" {
			name = "(unnamed)"
		}
		if len(name) > 33 {
			name = name[:30] + "..."
		}

		line := fmt.Sprintf("  %-18s %-35s %-18s %s", r.ResourceType, name, r.Region, shortID)
		if i == v.cursor {
			b.WriteString(selectedStyle.Render(line))
		} else {
			b.WriteString(normalStyle.Render(line))
		}
		b.WriteString("\n")
	}

	// Scroll indicator.
	if len(filtered) > maxRows {
		pct := 0
		if len(filtered)-maxRows > 0 {
			pct = (v.offset * 100) / (len(filtered) - maxRows)
		}
		b.WriteString(dimNavStyle.Render(fmt.Sprintf("\n  ↕ scroll %d%%", pct)))
	}

	return b.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
