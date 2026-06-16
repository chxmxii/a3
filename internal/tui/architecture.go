package tui

import (
	"fmt"
	"strings"

	"github.com/chxmxii/3a/internal/storage"
)

type architectureView struct {
	resources     []storage.Resource
	relationships []storage.Relationship
	scrollOffset  int
}

func (v *architectureView) render(width, height int) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  🏗️  Architecture"))
	b.WriteString("\n\n")

	if len(v.relationships) == 0 {
		b.WriteString(normalStyle.Render("  No relationships discovered."))
		b.WriteString("\n")
		b.WriteString(dimNavStyle.Render("  This can happen when resources lack cross-references or permissions are limited."))
		return b.String()
	}

	b.WriteString(dimNavStyle.Render(fmt.Sprintf("  %d relationships mapped", len(v.relationships))))
	b.WriteString("\n\n")

	// Build name/type lookups.
	nameMap := make(map[string]string)
	typeMap := make(map[string]string)
	for _, r := range v.resources {
		display := r.Name
		if display == "" {
			display = r.ResourceID
		}
		nameMap[r.ResourceID] = display
		typeMap[r.ResourceID] = r.ResourceType
	}

	// Generate tree lines.
	lines := v.buildTreeLines(nameMap, typeMap)

	// Apply scroll.
	maxRows := height - 8
	if maxRows < 5 {
		maxRows = 5
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
		b.WriteString(dimNavStyle.Render(fmt.Sprintf("\n  ↕ scroll %d%% (%d lines)", pct, len(lines))))
	}

	return b.String()
}

func (v *architectureView) buildTreeLines(nameMap, typeMap map[string]string) []string {
	var lines []string

	// Group relationships by source.
	bySource := make(map[string][]storage.Relationship)
	for _, rel := range v.relationships {
		bySource[rel.SourceID] = append(bySource[rel.SourceID], rel)
	}

	// Find roots (sources that are never targets).
	targetSet := make(map[string]bool)
	for _, rel := range v.relationships {
		targetSet[rel.TargetID] = true
	}

	var roots []string
	for source := range bySource {
		if !targetSet[source] {
			roots = append(roots, source)
		}
	}
	if len(roots) == 0 {
		for source := range bySource {
			roots = append(roots, source)
		}
	}

	rendered := make(map[string]bool)
	for _, root := range roots {
		if rendered[root] {
			continue
		}
		v.buildTree(&lines, root, "  ", true, bySource, nameMap, typeMap, rendered, 0)
	}

	return lines
}

func (v *architectureView) buildTree(lines *[]string, resourceID, prefix string, isLast bool, bySource map[string][]storage.Relationship, nameMap, typeMap map[string]string, rendered map[string]bool, depth int) {
	if depth > 5 || rendered[resourceID] {
		return
	}
	rendered[resourceID] = true

	connector := "├── "
	if isLast {
		connector = "└── "
	}
	if depth == 0 {
		connector = ""
	}

	name := nameMap[resourceID]
	if name == "" {
		name = resourceID
	}
	if len(name) > 40 {
		name = name[:37] + "..."
	}
	rType := typeMap[resourceID]
	if rType == "" {
		rType = "?"
	}

	line := fmt.Sprintf("%s%s%s %s", prefix, connector, normalStyle.Render("["+rType+"]"), name)
	*lines = append(*lines, line)

	children := bySource[resourceID]
	childPrefix := prefix + "│   "
	if isLast || depth == 0 {
		childPrefix = prefix + "    "
	}

	for i, rel := range children {
		isChildLast := i == len(children)-1
		v.buildTree(lines, rel.TargetID, childPrefix, isChildLast, bySource, nameMap, typeMap, rendered, depth+1)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
