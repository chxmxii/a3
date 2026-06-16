package tui

import "github.com/charmbracelet/lipgloss"

var (
	primaryColor = lipgloss.Color("#7C3AED")
	successColor = lipgloss.Color("#10B981")
	warningColor = lipgloss.Color("#F59E0B")
	dangerColor  = lipgloss.Color("#EF4444")
	mutedColor   = lipgloss.Color("#6B7280")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#F9FAFB"))

	helpStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			Background(lipgloss.Color("#374151"))

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#D1D5DB"))

	dimNavStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF"))

	severityCriticalStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(dangerColor)

	severityHighStyle = lipgloss.NewStyle().
				Foreground(dangerColor)

	severityMediumStyle = lipgloss.NewStyle().
				Foreground(warningColor)

	severityLowStyle = lipgloss.NewStyle().
				Foreground(successColor)

	passStyle = lipgloss.NewStyle().Foreground(successColor)
	failStyle = lipgloss.NewStyle().Foreground(dangerColor)
	warnStyle = lipgloss.NewStyle().Foreground(warningColor)

	regionBadgeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#818CF8")).
				Bold(true)
)
