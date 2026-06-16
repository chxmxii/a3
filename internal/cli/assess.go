package cli

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/chxmxii/3a/internal/architecture"
	"github.com/chxmxii/3a/internal/assessment"
	awsrules "github.com/chxmxii/3a/internal/assessment/rules/aws"
	ocirules "github.com/chxmxii/3a/internal/assessment/rules/oci"
	"github.com/chxmxii/3a/internal/checklist"
	"github.com/chxmxii/3a/internal/config"
	"github.com/chxmxii/3a/internal/cost"
	"github.com/chxmxii/3a/internal/discovery"
	"github.com/chxmxii/3a/internal/provider/steampipe"
	"github.com/chxmxii/3a/internal/sizing"
	"github.com/chxmxii/3a/internal/storage"
	"github.com/chxmxii/3a/internal/tui"
)

var (
	spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	stepStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED")).Bold(true)
	doneStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981"))
	failedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
)

func newAssessCmd() *cobra.Command {
	var connString string
	var noTUI bool

	cmd := &cobra.Command{
		Use:   "assess <profile>",
		Short: "Run a full assessment for a profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]
			return runAssessment(profileName, connString, noTUI)
		},
	}

	cmd.Flags().StringVar(&connString, "steampipe-conn", "postgres://steampipe@localhost:9193/steampipe", "Steampipe connection string")
	cmd.Flags().BoolVar(&noTUI, "no-tui", false, "skip TUI and print summary to stdout")

	return cmd
}

// step prints a step line with spinner-like prefix.
func step(icon, msg string) {
	fmt.Printf("  %s %s\n", stepStyle.Render(icon), msg)
}

// stepDone prints a completed step.
func stepDone(msg string) {
	fmt.Printf("  %s %s\n", doneStyle.Render("✓"), msg)
}

// stepFail prints a failed step (non-fatal).
func stepFail(msg string) {
	fmt.Printf("  %s %s\n", failedStyle.Render("✗"), dimStyle.Render(msg))
}

func runAssessment(profileName, connString string, noTUI bool) error {
	ctx := context.Background()

	// Suppress steampipe log spam during assessment.
	log.SetOutput(io.Discard)
	defer log.SetOutput(os.Stderr)

	// Load config.
	cfgPath := config.DefaultConfigPath()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		cfg = &config.Config{
			DBPath: resolveDBPath(getDBPath()),
			Profiles: []config.AccountProfile{
				{
					Name:     profileName,
					Provider: "aws",
					Regions:  []string{"us-east-1"},
				},
			},
		}
	}

	profile, err := config.GetProfile(cfg, profileName)
	if err != nil {
		return fmt.Errorf("profile error: %w", err)
	}

	// Open storage.
	dbFile := resolveDBPath(getDBPath())
	if cfg.DBPath != "" {
		dbFile = resolveDBPath(cfg.DBPath)
	}
	store, err := storage.Open(dbFile)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer store.Close()

	// Create assessment record.
	assessmentID := uuid.New().String()
	now := time.Now()
	a := &storage.Assessment{
		ID:        assessmentID,
		Profile:   profileName,
		Provider:  profile.Provider,
		Status:    "in_progress",
		StartedAt: now,
		Regions:   profile.Regions,
	}
	if err := store.CreateAssessment(a); err != nil {
		return fmt.Errorf("creating assessment: %w", err)
	}

	// Header.
	fmt.Println()
	fmt.Printf("  %s\n", stepStyle.Render("3A — Agnostic Account Assessment"))
	fmt.Printf("  %s\n", dimStyle.Render(fmt.Sprintf("Profile: %s | Provider: %s | ID: %s", profileName, profile.Provider, assessmentID[:8])))
	fmt.Println()

	// Step 1: Connect to Steampipe.
	step("→", "Connecting to Steampipe...")
	sp, err := steampipe.NewSteampipeProvider(connString, profile.Provider)
	if err != nil {
		return fmt.Errorf("creating steampipe provider: %w", err)
	}
	defer sp.Close()

	if err := sp.Authenticate(ctx); err != nil {
		stepFail("Connection failed")
		return fmt.Errorf("connecting to steampipe: %w", err)
	}

	// Step 2: Validate profile.
	step("→", "Validating credentials...")
	if err := sp.ValidateProfile(ctx); err != nil {
		stepFail("Validation failed")
		_ = store.UpdateAssessmentStatus(assessmentID, "failed", nil)
		return fmt.Errorf("profile validation failed:\n\n%w", err)
	}
	stepDone("Credentials validated")

	// Step 3: Discovery.
	step("→", "Discovering resources...")
	engine := discovery.NewEngine(sp, store)
	summary, err := engine.Run(ctx, assessmentID, profile.Regions)
	if err != nil {
		stepFail("Discovery failed")
		return fmt.Errorf("discovery failed: %w", err)
	}

	if summary.TotalResources == 0 {
		_ = store.UpdateAssessmentStatus(assessmentID, "failed", nil)
		stepFail("No resources discovered")
		return fmt.Errorf("discovery returned 0 resources — check Steampipe credentials and configuration")
	}
	stepDone(fmt.Sprintf("Discovered %d resources across %d regions", summary.TotalResources, len(summary.ByRegion)))

	// Step 4: Architecture.
	step("→", "Reconstructing architecture...")
	reconstructor := architecture.NewReconstructor(store, profile.Provider)
	if err := reconstructor.Reconstruct(assessmentID); err != nil {
		stepFail("Architecture: " + err.Error())
	} else {
		rels, _ := store.GetRelationshipsByAssessment(assessmentID)
		stepDone(fmt.Sprintf("Mapped %d relationships", len(rels)))
	}

	// Step 5: Assessment.
	step("→", "Running security assessment...")
	var rules []assessment.Rule
	switch profile.Provider {
	case "aws":
		rules = awsrules.AllRules()
	case "oci":
		rules = ocirules.AllRules()
	}
	assessEngine := assessment.NewEngine(store, rules)
	if err := assessEngine.Run(ctx, assessmentID); err != nil {
		stepFail("Assessment: " + err.Error())
	}
	findings, _ := store.GetFindingsByAssessment(assessmentID)
	stepDone(fmt.Sprintf("Evaluated rules — %d findings", len(findings)))

	// Step 6: Sizing.
	step("→", "Analyzing infrastructure sizing...")
	sizingAnalyzer := sizing.NewAnalyzer(store)
	sizingSummary, err := sizingAnalyzer.Analyze(assessmentID)
	if err != nil {
		stepFail("Sizing: " + err.Error())
	} else {
		stepDone(fmt.Sprintf("Sizing: %d vCPUs, %.1f GB memory", sizingSummary.TotalVCPUs, sizingSummary.TotalMemoryGB))
	}

	// Step 7: Cost.
	step("→", "Estimating costs...")
	costEstimator := cost.NewEstimator(store)
	costSummary, err := costEstimator.Estimate(assessmentID)
	if err != nil {
		stepFail("Cost: " + err.Error())
	} else {
		stepDone(fmt.Sprintf("Estimated $%.2f/month", costSummary.TotalMonthlyCost))
	}

	// Step 8: Checklist.
	step("→", "Generating checklist...")
	checkEngine := checklist.NewEngine(store)
	checkSummary, err := checkEngine.Generate(assessmentID)
	if err != nil {
		stepFail("Checklist: " + err.Error())
	} else {
		stepDone(fmt.Sprintf("Checklist: %d pass, %d fail, %d warn", checkSummary.PassCount, checkSummary.FailCount, checkSummary.WarnCount))
	}

	// Mark complete.
	completedAt := time.Now()
	_ = store.UpdateAssessmentStatus(assessmentID, "completed", &completedAt)

	elapsed := time.Since(now).Round(time.Millisecond)
	fmt.Println()
	fmt.Printf("  %s %s\n", doneStyle.Render("✓ Assessment complete"), dimStyle.Render(fmt.Sprintf("(%s)", elapsed)))
	fmt.Println()

	if noTUI {
		return nil
	}

	// Launch TUI.
	fmt.Printf("  %s\n\n", dimStyle.Render("Launching interactive view... (press q to quit)"))
	model := tui.NewModel(store, assessmentID)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	return nil
}

func resolveDBPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

func canaryTableForProvider(providerType string) string {
	switch providerType {
	case "aws":
		return "aws_account"
	case "oci":
		return "oci_identity_compartment"
	default:
		return "unknown"
	}
}
