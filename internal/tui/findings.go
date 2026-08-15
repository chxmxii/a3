package tui

import (
	"fmt"
	"strings"

	"github.com/chxmxii/a3/internal/storage"
)

type findingsView struct {
	findings       []storage.Finding
	severityFilter string
	cursor         int
	offset         int
}

// severityKeys maps key presses to severity filters.
var severityKeys = map[string]string{
	"c": "critical",
	"h": "high",
	"m": "medium",
	"l": "low",
}

// handleKey handles a key press for the findings view. Returns true if the
// key was handled. The scroll offset is clamped to the cursor at render time.
func (v *findingsView) handleKey(key string) bool {
	switch key {
	case "up", "k":
		if v.cursor > 0 {
			v.cursor--
		}
	case "down", "j":
		if v.cursor < len(v.filteredFindings())-1 {
			v.cursor++
		}
	case "x":
		v.severityFilter = ""
		v.cursor = 0
	default:
		severity, ok := severityKeys[key]
		if !ok {
			return false
		}
		v.severityFilter = severity
		v.cursor = 0
	}
	return true
}

func (v *findingsView) render(width, height int) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  🔍 Security Findings"))
	b.WriteString("\n\n")

	// Filter bar.
	if v.severityFilter != "" {
		b.WriteString(regionBadgeStyle.Render(fmt.Sprintf("  Filter: %s", strings.ToUpper(v.severityFilter))))
		b.WriteString(dimNavStyle.Render("  (x to clear)"))
	} else {
		b.WriteString(dimNavStyle.Render("  c:critical  h:high  m:medium  l:low  x:clear"))
	}
	b.WriteString("\n")

	filtered := v.filteredFindings()
	b.WriteString(dimNavStyle.Render(fmt.Sprintf("  %d of %d findings", len(filtered), len(v.findings))))
	b.WriteString("\n\n")

	if len(filtered) == 0 {
		if len(v.findings) == 0 {
			b.WriteString(passStyle.Render("  ✓ No security findings — all checks passed!"))
		} else {
			b.WriteString(normalStyle.Render("  No findings match the current filter."))
		}
		return b.String()
	}

	// Visible rows.
	maxRows := height - 11
	if maxRows < 5 {
		maxRows = 5
	}

	clampCursorScroll(v.cursor, &v.offset, maxRows)

	end := min(v.offset+maxRows, len(filtered))

	for i := v.offset; i < end; i++ {
		f := filtered[i]

		var sevLabel string
		switch f.Severity {
		case "critical":
			sevLabel = severityCriticalStyle.Render("CRIT")
		case "high":
			sevLabel = severityHighStyle.Render("HIGH")
		case "medium":
			sevLabel = severityMediumStyle.Render("MED ")
		case "low":
			sevLabel = severityLowStyle.Render("LOW ")
		default:
			sevLabel = normalStyle.Render("INFO")
		}

		desc := f.Description
		maxDesc := width - 20
		if maxDesc < 40 {
			maxDesc = 40
		}
		if len(desc) > maxDesc {
			desc = desc[:maxDesc-3] + "..."
		}

		line := fmt.Sprintf("  %s  %s", sevLabel, desc)
		if i == v.cursor {
			b.WriteString(selectedStyle.Render(line))
		} else {
			b.WriteString(normalStyle.Render(line))
		}
		b.WriteString("\n")

		// Show details for selected item.
		if i == v.cursor {
			b.WriteString(dimNavStyle.Render(fmt.Sprintf("         Resource: %s", truncate(f.ResourceID, 60))))
			b.WriteString("\n")
			if f.Recommendation != "" {
				b.WriteString(dimNavStyle.Render(fmt.Sprintf("         Fix: %s", truncate(f.Recommendation, 70))))
				b.WriteString("\n")
			}
			if f.StandardName != "" {
				b.WriteString(dimNavStyle.Render(fmt.Sprintf("         Standard: %s [%s]", f.StandardName, f.ControlID)))
				b.WriteString("\n")
			}
		}
	}

	// Scroll indicator.
	b.WriteString(scrollFooter(len(filtered), v.offset, maxRows))

	return b.String()
}

func (v *findingsView) filteredFindings() []storage.Finding {
	if v.severityFilter == "" {
		return v.findings
	}
	var filtered []storage.Finding
	for _, f := range v.findings {
		if f.Severity == v.severityFilter {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
