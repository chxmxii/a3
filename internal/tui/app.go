package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/chxmxii/a3/internal/storage"
)

// View represents the currently active TUI view.
type View int

const (
	ViewOverview View = iota
	ViewInventory
	ViewArchitecture
	ViewFindings
	ViewCost
)

// Model is the root Bubble Tea model for the 3A TUI.
type Model struct {
	store        *storage.Store
	assessmentID string
	activeView   View
	width        int
	height       int

	// View models.
	overview     overviewView
	inventory    inventoryView
	architecture architectureView
	findings     findingsView
	cost         costView

	spinner  spinner.Model
	showHelp bool
	loaded   bool
	err      error
}

// dataLoadedMsg is sent when data has been loaded from the store.
type dataLoadedMsg struct {
	overview     overviewView
	inventory    inventoryView
	architecture architectureView
	findings     findingsView
	cost         costView
}

// errMsg wraps an error for the TUI.
type errMsg struct{ err error }

// NewModel creates a new TUI model for the given assessment.
func NewModel(store *storage.Store, assessmentID string) Model {
	return Model{
		store:        store,
		assessmentID: assessmentID,
		activeView:   ViewOverview,
		spinner:      spinner.New(spinner.WithSpinner(spinner.Dot), spinner.WithStyle(spinnerStyle)),
	}
}

// Init starts the TUI.
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.loadData, m.spinner.Tick)
}

func (m Model) loadData() tea.Msg {
	assessment, err := m.store.GetAssessment(m.assessmentID)
	if err != nil {
		return errMsg{err}
	}

	resources, err := m.store.GetResourcesByAssessment(m.assessmentID)
	if err != nil {
		return errMsg{err}
	}

	findings, err := m.store.GetFindingsByAssessment(m.assessmentID)
	if err != nil {
		return errMsg{err}
	}

	relationships, err := m.store.GetRelationshipsByAssessment(m.assessmentID)
	if err != nil {
		return errMsg{err}
	}

	costs, err := m.store.GetCostsByAssessment(m.assessmentID)
	if err != nil {
		return errMsg{err}
	}

	// Build region list from resources.
	regionSet := make(map[string]bool)
	for _, r := range resources {
		if r.Region != "" {
			regionSet[r.Region] = true
		}
	}
	var regions []string
	for reg := range regionSet {
		regions = append(regions, reg)
	}
	sort.Strings(regions)

	return dataLoadedMsg{
		overview: overviewView{
			assessment: assessment,
			resources:  resources,
			findings:   findings,
			costs:      costs,
		},
		inventory: inventoryView{
			resources: resources,
			regions:   regions,
		},
		architecture: architectureView{
			resources:     resources,
			relationships: relationships,
		},
		findings: findingsView{findings: findings},
		cost:     costView{costs: costs, resources: resources},
	}
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()

		// Always allow ctrl+c to quit, even in text-entry mode.
		if key == "ctrl+c" {
			return m, tea.Quit
		}

		// A view in text-entry mode (e.g. inventory search) gets every key,
		// including the global ones.
		if m.viewCapturesInput() {
			m.handleViewKey(key)
			return m, nil
		}

		// Any key closes the help overlay.
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}

		switch key {
		case "q":
			return m, tea.Quit
		case "?":
			m.showHelp = true
		case "1":
			m.activeView = ViewOverview
		case "2":
			m.activeView = ViewInventory
		case "3":
			m.activeView = ViewArchitecture
		case "4":
			m.activeView = ViewFindings
		case "5":
			m.activeView = ViewCost
		default:
			// Forward everything else to the active view.
			m.handleViewKey(key)
		}

	case spinner.TickMsg:
		if !m.loaded && m.err == nil {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case dataLoadedMsg:
		m.loaded = true
		m.overview = msg.overview
		m.inventory = msg.inventory
		m.architecture = msg.architecture
		m.findings = msg.findings
		m.cost = msg.cost

	case errMsg:
		m.err = msg.err
	}

	return m, nil
}

// viewCapturesInput reports whether the active view is in a text-entry mode
// that must receive every key press, including global ones.
func (m *Model) viewCapturesInput() bool {
	return m.activeView == ViewInventory && m.inventory.capturesInput()
}

// handleViewKey forwards a non-global key to the active view. Returns true
// if the view handled it.
func (m *Model) handleViewKey(key string) bool {
	switch m.activeView {
	case ViewOverview:
		return m.overview.handleKey(key)
	case ViewInventory:
		return m.inventory.handleKey(key)
	case ViewArchitecture:
		return m.architecture.handleKey(key)
	case ViewFindings:
		return m.findings.handleKey(key)
	case ViewCost:
		return m.cost.handleKey(key)
	}
	return false
}

// View renders the TUI.
func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("\n  Error: %v\n\n  Press q to quit.\n", m.err)
	}

	if !m.loaded {
		return "\n  " + m.spinner.View() + " Loading assessment data...\n"
	}

	// Fixed layout: nav (2 lines) + content (fills) + help (1 line).
	nav := m.renderNav()
	help := m.renderHelp()

	// Content area height = total height - nav (2) - help (2) - borders.
	contentHeight := m.height - 5
	if contentHeight < 10 {
		contentHeight = 10
	}

	var content string
	if m.showHelp {
		content = m.renderHelpOverlay(contentHeight)
	} else {
		switch m.activeView {
		case ViewOverview:
			content = m.overview.render(m.width, contentHeight)
		case ViewInventory:
			content = m.inventory.render(m.width, contentHeight)
		case ViewArchitecture:
			content = m.architecture.render(m.width, contentHeight)
		case ViewFindings:
			content = m.findings.render(m.width, contentHeight)
		case ViewCost:
			content = m.cost.render(m.width, contentHeight)
		}
	}

	// Truncate content if it exceeds available height.
	contentLines := strings.Split(content, "\n")
	if len(contentLines) > contentHeight {
		contentLines = contentLines[:contentHeight]
	}
	content = strings.Join(contentLines, "\n")

	return nav + "\n" + content + "\n\n" + help
}

func (m Model) renderNav() string {
	tabs := []struct {
		key  string
		name string
		view View
	}{
		{"1", "Overview", ViewOverview},
		{"2", "Inventory", ViewInventory},
		{"3", "Architecture", ViewArchitecture},
		{"4", "Findings", ViewFindings},
		{"5", "Cost", ViewCost},
	}

	var parts []string
	for _, tab := range tabs {
		label := fmt.Sprintf(" %s %s ", tab.key, tab.name)
		if tab.view == m.activeView {
			parts = append(parts, selectedStyle.Render(label))
		} else {
			parts = append(parts, dimNavStyle.Render(label))
		}
	}

	return "\n " + strings.Join(parts, dimNavStyle.Render("│"))
}

func (m Model) renderHelp() string {
	if m.showHelp {
		return helpStyle.Render("  press any key to close help")
	}
	base := "q:quit  ↑↓:scroll  1-5:views  ?:help"
	switch m.activeView {
	case ViewInventory:
		if m.inventory.searchInput {
			return helpStyle.Render("  type to search  enter:apply  esc:cancel")
		}
		if m.inventory.showDetail {
			base += "  esc/x:back  ↑↓:scroll"
		} else {
			base += "  enter:details  /:search  r/R:region  t/T:type  x:clear"
		}
	case ViewArchitecture:
		base += "  n:network  v:resource"
	case ViewFindings:
		base += "  c/h/m/l:severity  x:clear"
	}
	return helpStyle.Render("  " + base)
}

// renderHelpOverlay renders a centered box listing all keybindings.
func (m Model) renderHelpOverlay(contentHeight int) string {
	rows := []string{
		titleStyle.Render("Keybindings"),
		"",
		keyStyle.Render("Global"),
		"  q / ctrl+c    quit",
		"  1-5           switch view",
		"  ?             toggle this help",
		"  ↑↓ / k j      scroll / move cursor",
		"",
		keyStyle.Render("Inventory"),
		"  enter         resource details",
		"  esc           back to list",
		"  /             search (enter:apply, esc:cancel)",
		"  r / R         cycle region filter",
		"  t / T         cycle type filter",
		"  x             clear filters / close details",
		"",
		keyStyle.Render("Architecture"),
		"  n             network view",
		"  v             resource view",
		"",
		keyStyle.Render("Findings"),
		"  c / h / m / l filter by severity",
		"  x             clear filter",
		"",
		dimNavStyle.Render("press any key to close"),
	}
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Padding(0, 2).
		Render(strings.Join(rows, "\n"))
	return lipgloss.Place(m.width, contentHeight, lipgloss.Center, lipgloss.Center, box)
}
