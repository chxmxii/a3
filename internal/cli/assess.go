package cli

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/chxmxii/a3/internal/architecture"
	"github.com/chxmxii/a3/internal/assessment"
	awsrules "github.com/chxmxii/a3/internal/assessment/rules/aws"
	ocirules "github.com/chxmxii/a3/internal/assessment/rules/oci"
	"github.com/chxmxii/a3/internal/checklist"
	"github.com/chxmxii/a3/internal/config"
	"github.com/chxmxii/a3/internal/cost"
	"github.com/chxmxii/a3/internal/discovery"
	"github.com/chxmxii/a3/internal/provider/steampipe"
	"github.com/chxmxii/a3/internal/sizing"
	"github.com/chxmxii/a3/internal/storage"
	"github.com/chxmxii/a3/internal/tui"
)

var (
	frames      = []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}
	accentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED")).Bold(true)
	doneStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981"))
	failStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444"))
	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
	warnStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B"))
)

// spinner runs an animated spinner while a task executes.
type spinner struct {
	mu      sync.Mutex
	msg     string
	running bool
	done    chan struct{}
}

func newSpinner(msg string) *spinner {
	s := &spinner{msg: msg, done: make(chan struct{})}
	s.start()
	return s
}

func (s *spinner) start() {
	s.running = true
	go func() {
		i := 0
		for {
			select {
			case <-s.done:
				return
			default:
				s.mu.Lock()
				frame := accentStyle.Render(frames[i%len(frames)])
				fmt.Printf("\r  %s %s", frame, s.msg)
				s.mu.Unlock()
				i++
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()
}

func (s *spinner) succeed(result string) {
	close(s.done)
	s.running = false
	// Clear the full line then print result.
	fmt.Printf("\r\033[K  %s %s\n", doneStyle.Render("✓"), result)
}

func (s *spinner) fail(result string) {
	close(s.done)
	s.running = false
	fmt.Printf("\r\033[K  %s %s\n", failStyle.Render("✗"), dimStyle.Render(result))
}

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

	cmd.Flags().StringVar(&connString, "steampipe-conn", "", "Steampipe connection string (default \""+defaultSteampipeConn+"\")")
	cmd.Flags().BoolVar(&noTUI, "no-tui", false, "skip TUI and print summary to stdout")

	return cmd
}

// printBanner prints the ASCII header and assessment identity line.
func printBanner(profileName, provider, assessmentID string) {
	fmt.Println()
	fmt.Println(accentStyle.Render("    ___   _____   "))
	fmt.Println(accentStyle.Render("   / _ \\ |____ |  "))
	fmt.Println(accentStyle.Render("  / /_\\ \\    / /  "))
	fmt.Println(accentStyle.Render("  |  _  |    \\ \\  "))
	fmt.Println(accentStyle.Render("  | | | |.___/ /  "))
	fmt.Println(accentStyle.Render("  \\_| |_/\\____/   "))
	fmt.Println()
	fmt.Printf("  %s\n", dimStyle.Render("Agnostic Account Assessment"))
	fmt.Printf("  %s\n\n", dimStyle.Render(fmt.Sprintf("Profile: %s | Provider: %s | ID: %s", profileName, provider, assessmentID[:8])))
}

// runStep runs fn under a spinner. fn returns the message to display and an
// error; on error the message is shown as the failure line. A non-nil error is
// returned to the caller only when fatal is true — non-fatal steps log the
// failure and let the assessment continue.
func runStep(label string, fatal bool, fn func() (string, error)) error {
	s := newSpinner(label)
	msg, err := fn()
	if err != nil {
		s.fail(msg)
		if fatal {
			return err
		}
		return nil
	}
	s.succeed(msg)
	return nil
}

func runAssessment(profileName, connString string, noTUI bool) error {
	ctx := context.Background()

	// Suppress steampipe log noise.
	log.SetOutput(io.Discard)
	defer log.SetOutput(os.Stderr)

	// Ensure ~/.a3 directory exists.
	if _, err := config.EnsureConfigDir(); err != nil {
		return err
	}

	// Load config. Without a config file, fall back to a default profile so a
	// bare `a3 assess <name>` still works against aws/us-east-1.
	cfg, err := config.Load(config.DefaultConfigPath())
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

	store, err := openStore(cfg)
	if err != nil {
		return err
	}
	defer store.Close()

	connString = steampipeConnString(connString, cfg)

	// Create assessment.
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

	printBanner(profileName, profile.Provider, assessmentID)

	// Step 1: Connect and validate.
	sp1 := newSpinner("Connecting to Steampipe and validating credentials...")
	sp, err := steampipe.NewSteampipeProvider(connString, profile.Provider)
	if err != nil {
		sp1.fail("Connection failed")
		return fmt.Errorf("creating steampipe provider: %w", err)
	}
	defer sp.Close()
	if err := sp.Authenticate(ctx); err != nil {
		sp1.fail("Connection failed")
		return fmt.Errorf("connecting to steampipe: %w", err)
	}
	warn, err := sp.ValidateProfile(ctx)
	if err != nil {
		sp1.fail("Validation failed")
		_ = store.UpdateAssessmentStatus(assessmentID, "failed", nil)
		return fmt.Errorf("profile validation failed:\n\n%w", err)
	}
	if warn != "" {
		sp1.succeed("Connected (limited access)")
		fmt.Printf("  %s\n", warnStyle.Render("⚠ "+warn))
	} else {
		sp1.succeed("Connected and validated")
	}

	// Step 2: Discover.
	sp3 := newSpinner("Discovering resources (this may take a moment)...")
	engine := discovery.NewEngine(sp, store)
	summary, err := engine.Run(ctx, assessmentID, profile.Regions)
	if err != nil {
		sp3.fail("Discovery failed")
		return fmt.Errorf("discovery failed: %w", err)
	}
	if summary.TotalResources == 0 {
		sp3.fail("No resources found")
		_ = store.UpdateAssessmentStatus(assessmentID, "failed", nil)
		return fmt.Errorf("discovery returned 0 resources — check Steampipe credentials")
	}
	sp3.succeed(fmt.Sprintf("Discovered %d resources across %d regions", summary.TotalResources, len(summary.ByRegion)))

	// Step 4: Architecture (non-fatal).
	_ = runStep("Reconstructing architecture...", false, func() (string, error) {
		reconstructor := architecture.NewReconstructor(store, profile.Provider)
		if err := reconstructor.Reconstruct(assessmentID); err != nil {
			return "Architecture: " + err.Error(), err
		}
		rels, _ := store.GetRelationshipsByAssessment(assessmentID)
		return fmt.Sprintf("Mapped %d relationships", len(rels)), nil
	})

	// Step 5: Assessment (non-fatal).
	_ = runStep("Running security assessment...", false, func() (string, error) {
		var rules []assessment.Rule
		switch profile.Provider {
		case "aws":
			rules = awsrules.AllRules()
		case "oci":
			rules = ocirules.AllRules()
		}
		assessEngine := assessment.NewEngine(store, rules)
		_ = assessEngine.Run(ctx, assessmentID)
		findings, _ := store.GetFindingsByAssessment(assessmentID)
		return fmt.Sprintf("Security: %d findings", len(findings)), nil
	})

	// Step 6: Sizing (non-fatal).
	_ = runStep("Analyzing infrastructure sizing...", false, func() (string, error) {
		sizingSummary, err := sizing.NewAnalyzer(store).Analyze(assessmentID)
		if err != nil {
			return "Sizing unavailable", err
		}
		return fmt.Sprintf("Sizing: %d vCPUs, %.1f GB memory", sizingSummary.TotalVCPUs, sizingSummary.TotalMemoryGB), nil
	})

	// Step 7: Cost (non-fatal). Try real billing data from AWS Cost Explorer
	// (via Steampipe) first, then fall back to static estimation.
	_ = runStep("Fetching cost data...", false, func() (string, error) {
		billingData, billingErr := cost.QueryBilling(ctx, sp.Pool())
		if billingErr == nil && billingData != nil {
			cost.StoreBillingCosts(store, assessmentID, billingData)
			return fmt.Sprintf("Actual cost: $%.2f/month (%s)", billingData.TotalMonthlyCost, billingData.Message), nil
		}
		costSummary, err := cost.NewEstimator(store).Estimate(assessmentID)
		if err != nil {
			return "Cost estimation unavailable", err
		}
		return fmt.Sprintf("Estimated $%.2f/month (static catalog — real billing data unavailable)", costSummary.TotalMonthlyCost), nil
	})

	// Step 8: Checklist (non-fatal).
	_ = runStep("Generating checklist...", false, func() (string, error) {
		checkSummary, err := checklist.NewEngine(store).Generate(assessmentID)
		if err != nil {
			return "Checklist unavailable", err
		}
		return fmt.Sprintf("Checklist: %d pass, %d fail, %d warn", checkSummary.PassCount, checkSummary.FailCount, checkSummary.WarnCount), nil
	})

	// Done.
	completedAt := time.Now()
	_ = store.UpdateAssessmentStatus(assessmentID, "completed", &completedAt)
	elapsed := time.Since(now).Round(time.Millisecond)

	fmt.Println()
	fmt.Printf("  %s  %s\n\n", doneStyle.Render("✓ Assessment complete"), dimStyle.Render(elapsed.String()))

	if noTUI {
		return nil
	}

	fmt.Printf("  %s\n\n", dimStyle.Render("Launching TUI... (q to quit)"))
	time.Sleep(300 * time.Millisecond)

	model := tui.NewModel(store, assessmentID)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	return nil
}
