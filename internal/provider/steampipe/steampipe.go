package steampipe

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/chxmxii/a3/internal/provider"
)

// SteampipeProvider implements the provider.Provider interface using Steampipe's
// PostgreSQL-compatible endpoint for resource discovery.
type SteampipeProvider struct {
	pool         *pgxpool.Pool
	connString   string
	providerType string // "aws", "oci", or "azure"
}

// NewSteampipeProvider creates a new Steampipe-backed provider.
func NewSteampipeProvider(connString string, providerType string) (*SteampipeProvider, error) {
	if connString == "" {
		connString = "postgres://steampipe@localhost:9193/steampipe"
	}
	return &SteampipeProvider{
		connString:   connString,
		providerType: providerType,
	}, nil
}

// Name returns the provider type identifier.
func (s *SteampipeProvider) Name() string { return s.providerType }

// Authenticate connects to the Steampipe database and verifies connectivity.
func (s *SteampipeProvider) Authenticate(ctx context.Context) error {
	pool, err := pgxpool.New(ctx, s.connString)
	if err != nil {
		return fmt.Errorf("failed to connect to steampipe: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return fmt.Errorf("steampipe connection test failed: %w", err)
	}
	s.pool = pool
	return nil
}

// Discoverer returns the Steampipe-based resource discoverer.
func (s *SteampipeProvider) Discoverer() provider.Discoverer {
	return &SteampipeDiscoverer{pool: s.pool, providerType: s.providerType}
}

// MetricsClient returns nil — Steampipe discovery doesn't provide metrics.
func (s *SteampipeProvider) MetricsClient() provider.MetricsClient { return nil }

// PricingClient returns nil — pricing is handled separately.
func (s *SteampipeProvider) PricingClient() provider.PricingClient { return nil }

// ValidateProfile checks that Steampipe can actually query data for the configured
// provider type. Returns an error with actionable guidance if the plugin isn't
// installed, the connection isn't configured, or no data is returned.
func (s *SteampipeProvider) ValidateProfile(ctx context.Context) error {
	if s.pool == nil {
		return fmt.Errorf("steampipe connection not established — call Authenticate first")
	}

	// Pick a lightweight "canary" table per provider to test connectivity.
	var canaryTable string
	switch s.providerType {
	case "aws":
		canaryTable = "aws_account"
	case "oci":
		canaryTable = "oci_identity_compartment"
	case "azure":
		canaryTable = "azure_subscription"
	default:
		return fmt.Errorf("unsupported provider type: %s", s.providerType)
	}

	// Check if the table exists.
	var tableExists bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_name = $1
		)
	`, canaryTable).Scan(&tableExists)
	if err != nil {
		return fmt.Errorf("failed to check steampipe tables: %w\n\nIs the %s plugin installed? Run:\n  steampipe plugin install %s", err, s.providerType, s.providerType)
	}
	if !tableExists {
		return fmt.Errorf("steampipe table %q not found\n\nThe %s plugin may not be installed or configured. Run:\n  steampipe plugin install %s\n  steampipe plugin list", canaryTable, s.providerType, s.providerType)
	}

	// Check if we can actually get rows (credentials are working).
	var rowCount int
	query := fmt.Sprintf("SELECT count(*) FROM %s", canaryTable)
	err = s.pool.QueryRow(ctx, query).Scan(&rowCount)
	if err != nil {
		return fmt.Errorf("steampipe query to %s failed: %w\n\nThis usually means cloud credentials aren't configured.\nCheck your Steampipe connection config:\n  cat ~/.steampipe/config/%s.spc", canaryTable, err, s.providerType)
	}
	if rowCount == 0 {
		return fmt.Errorf("steampipe returned 0 rows from %s\n\nPossible causes:\n  • Cloud credentials not configured in ~/.steampipe/config/%s.spc\n  • The credentials don't have permission to list resources\n  • The account has no resources of this type\n\nTest manually:\n  steampipe query \"SELECT * FROM %s\"", canaryTable, s.providerType, canaryTable)
	}

	return nil
}

// Close releases the connection pool.
func (s *SteampipeProvider) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}

// Pool returns the underlying connection pool for direct queries.
func (s *SteampipeProvider) Pool() *pgxpool.Pool {
	return s.pool
}
