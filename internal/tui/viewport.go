package tui

import (
	"fmt"
	"strings"
)

// clampScroll clamps *offset into [0, total-maxRows] so the window stays
// within the content.
func clampScroll(total int, offset *int, maxRows int) {
	if *offset > total-maxRows {
		*offset = max(0, total-maxRows)
	}
	if *offset < 0 {
		*offset = 0
	}
}

// clampCursorScroll adjusts *offset so that cursor stays visible within a
// window of maxRows rows.
func clampCursorScroll(cursor int, offset *int, maxRows int) {
	if cursor < *offset {
		*offset = cursor
	}
	if cursor >= *offset+maxRows {
		*offset = cursor - maxRows + 1
	}
	if *offset < 0 {
		*offset = 0
	}
}

// renderScrollable clamps *offset and returns the visible window of lines,
// each followed by a newline.
func renderScrollable(lines []string, offset *int, maxRows int) string {
	clampScroll(len(lines), offset, maxRows)
	end := min(*offset+maxRows, len(lines))
	var b strings.Builder
	for i := *offset; i < end; i++ {
		b.WriteString(lines[i])
		b.WriteString("\n")
	}
	return b.String()
}

// scrollPct returns the scroll position as a percentage of the scrollable
// range, or 0 when everything fits.
func scrollPct(total, offset, maxRows int) int {
	if total-maxRows > 0 {
		return (offset * 100) / (total - maxRows)
	}
	return 0
}

// scrollFooter returns the standard dim "↕ scroll N%" footer (preceded by a
// blank line), or "" when the content fits within maxRows.
func scrollFooter(total, offset, maxRows int) string {
	if total <= maxRows {
		return ""
	}
	return dimNavStyle.Render(fmt.Sprintf("\n  ↕ scroll %d%%", scrollPct(total, offset, maxRows)))
}
