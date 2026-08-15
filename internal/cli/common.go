package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chxmxii/a3/internal/config"
	"github.com/chxmxii/a3/internal/storage"
)

const defaultSteampipeConn = "postgres://steampipe@localhost:9193/steampipe"

// openStore opens the SQLite store. Note: a DBPath from config.yaml takes
// precedence over the --db flag (long-standing behavior, preserved here).
func openStore(cfg *config.Config) (*storage.Store, error) {
	dbFile := resolveDBPath(getDBPath())
	if cfg.DBPath != "" {
		dbFile = resolveDBPath(cfg.DBPath)
	}
	store, err := storage.Open(dbFile)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}
	return store, nil
}

// steampipeConnString resolves the Steampipe connection string:
// --steampipe-conn flag > config.yaml steampipe.connection_string > default.
func steampipeConnString(flagValue string, cfg *config.Config) string {
	if flagValue != "" {
		return flagValue
	}
	if cfg.Steampipe.ConnectionString != "" {
		return cfg.Steampipe.ConnectionString
	}
	return defaultSteampipeConn
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
