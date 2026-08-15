package discovery

import (
	"context"
	"fmt"

	"github.com/chxmxii/a3/internal/provider"
	"github.com/chxmxii/a3/internal/storage"
)

// DiscoverySummary reports what was found during discovery.
type DiscoverySummary struct {
	TotalResources int
	ByType         map[provider.ResourceType]int
	ByRegion       map[string]int
	Errors         []DiscoveryError
}

// DiscoveryError records a failed service/region scan.
type DiscoveryError struct {
	Service string
	Region  string
	Err     error
}

// resultsBuffer sizes the channel between the discoverer goroutine and the
// persisting loop.
const resultsBuffer = 100

// Engine orchestrates resource discovery across regions.
type Engine struct {
	provider provider.Provider
	store    *storage.Store
}

// NewEngine creates a new discovery Engine with the given provider and store.
func NewEngine(p provider.Provider, store *storage.Store) *Engine {
	return &Engine{
		provider: p,
		store:    store,
	}
}

// Run executes discovery for the given assessment and regions, returning a summary.
func (e *Engine) Run(ctx context.Context, assessmentID string, regions []string) (*DiscoverySummary, error) {
	discoverer := e.provider.Discoverer()
	if discoverer == nil {
		return nil, fmt.Errorf("provider %s returned nil discoverer", e.provider.Name())
	}

	// Create a buffered results channel.
	results := make(chan provider.DiscoveredResource, resultsBuffer)

	// Launch discovery in a separate goroutine so we can read results concurrently.
	var discoverErr error
	done := make(chan struct{})
	go func() {
		defer close(done)
		discoverErr = discoverer.DiscoverResources(ctx, regions, results)
		close(results)
	}()

	summary := &DiscoverySummary{
		ByType:   make(map[provider.ResourceType]int),
		ByRegion: make(map[string]int),
	}

	// Read from the results channel and persist each resource.
	for res := range results {
		resource := &storage.Resource{
			AssessmentID: assessmentID,
			ProviderType: res.ProviderType,
			ResourceType: string(res.ResourceType),
			ResourceID:   res.ResourceID,
			Region:       res.Region,
			Name:         res.Name,
			Tags:         res.Tags,
			RawMetadata:  res.RawMetadata,
		}

		if err := e.store.InsertResource(resource); err != nil {
			// Record error but continue processing remaining resources.
			summary.Errors = append(summary.Errors, DiscoveryError{
				Service: string(res.ResourceType),
				Region:  res.Region,
				Err:     fmt.Errorf("persisting resource %s: %w", res.ResourceID, err),
			})
			continue
		}

		summary.TotalResources++
		summary.ByType[res.ResourceType]++
		summary.ByRegion[res.Region]++
	}

	// Wait for the discoverer goroutine to finish.
	<-done

	if discoverErr != nil {
		summary.Errors = append(summary.Errors, DiscoveryError{
			Service: "discovery",
			Region:  "",
			Err:     discoverErr,
		})
	}

	return summary, nil
}
