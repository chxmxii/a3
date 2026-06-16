package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/chxmxii/3a/internal/config"
	"github.com/chxmxii/3a/internal/report"
	"github.com/chxmxii/3a/internal/storage"
)

func newReportCmd() *cobra.Command {
	var format string
	var output string

	cmd := &cobra.Command{
		Use:   "report <profile>",
		Short: "Generate a report from the latest assessment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]
			return runReport(profileName, format, output)
		},
	}

	cmd.Flags().StringVar(&format, "format", "markdown", "output format (markdown, json, or excel)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output file path (default: stdout, required for excel)")

	return cmd
}

func runReport(profileName, format, output string) error {
	cfgPath := config.DefaultConfigPath()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	dbFile := resolveDBPath(getDBPath())
	if cfg.DBPath != "" {
		dbFile = resolveDBPath(cfg.DBPath)
	}

	store, err := storage.Open(dbFile)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer store.Close()

	// Find the latest assessment for this profile.
	assessment, err := store.GetLatestAssessment(profileName)
	if err != nil {
		return fmt.Errorf("querying assessments: %w", err)
	}
	if assessment == nil {
		return fmt.Errorf("no assessments found for profile %q. Run '3a assess %s' first.", profileName, profileName)
	}

	// Determine format.
	var reportFormat report.Format
	switch format {
	case "markdown", "md":
		reportFormat = report.FormatMarkdown
	case "json":
		reportFormat = report.FormatJSON
	case "excel", "xlsx":
		// Excel needs a file path.
		if output == "" {
			output = fmt.Sprintf("%s-report.xlsx", profileName)
		}
		gen := report.NewGenerator(store)
		if err := gen.GenerateExcel(assessment.ID, output); err != nil {
			return fmt.Errorf("generating excel report: %w", err)
		}
		fmt.Printf("Excel report written to %s\n", output)
		return nil
	default:
		return fmt.Errorf("unsupported format: %s (use markdown, json, or excel)", format)
	}

	// Generate report.
	gen := report.NewGenerator(store)
	content, err := gen.Generate(assessment.ID, reportFormat)
	if err != nil {
		return fmt.Errorf("generating report: %w", err)
	}

	// Output.
	if output != "" {
		if err := os.WriteFile(output, []byte(content), 0644); err != nil {
			return fmt.Errorf("writing report: %w", err)
		}
		fmt.Printf("Report written to %s\n", output)
	} else {
		fmt.Print(content)
	}

	return nil
}
