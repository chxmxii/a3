package cost

// This file loads the embedded pricing catalog (data/aws_pricing.json).
// All prices assume us-east-1 on-demand pricing, as the previous hardcoded
// Go literals did. Instance rows carry the hourly price and, where known,
// vCPU/memory specs shared with internal/sizing.

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"
)

//go:embed data/aws_pricing.json
var pricingJSON []byte

// InstancePricing is one row of the embedded catalog: an instance type's
// hourly on-demand price plus optional hardware specs (zero/empty when the
// catalog has no spec data for the type).
type InstancePricing struct {
	Type      string  `json:"type"`
	VCPUs     int     `json:"vcpus,omitempty"`
	MemGB     float64 `json:"mem_gb,omitempty"`
	Family    string  `json:"family,omitempty"`
	HourlyUSD float64 `json:"hourly_usd"`
}

type pricingData struct {
	Instances      []InstancePricing  `json:"instances"`
	StorageGBMonth map[string]float64 `json:"storage_gb_month"`
	ServicesHourly map[string]float64 `json:"services_hourly"`
}

// Keys into the services_hourly section of the catalog.
const (
	svcNATGateway      = "nat_gateway"
	svcALB             = "alb"
	svcNLB             = "nlb"
	svcEKSControlPlane = "eks_control_plane"
)

var (
	pricingOnce   sync.Once
	pricing       pricingData
	instanceIndex map[string]InstancePricing
)

func loadPricing() {
	pricingOnce.Do(func() {
		if err := json.Unmarshal(pricingJSON, &pricing); err != nil {
			panic(fmt.Sprintf("cost: invalid embedded pricing data: %v", err))
		}
		instanceIndex = make(map[string]InstancePricing, len(pricing.Instances))
		for _, row := range pricing.Instances {
			instanceIndex[row.Type] = row
		}
	})
}

// LookupInstance returns the catalog row for an instance type (EC2 or RDS
// class), or false if the type is unknown.
func LookupInstance(instanceType string) (InstancePricing, bool) {
	loadPricing()
	row, ok := instanceIndex[instanceType]
	return row, ok
}

// instanceHourly returns the hourly on-demand price for an instance type.
func instanceHourly(instanceType string) (float64, bool) {
	row, ok := LookupInstance(instanceType)
	return row.HourlyUSD, ok
}

// storageGBMonth returns the monthly per-GB price for a storage type.
func storageGBMonth(volumeType string) (float64, bool) {
	loadPricing()
	price, ok := pricing.StorageGBMonth[volumeType]
	return price, ok
}

// serviceHourly returns the flat hourly price for a service (svc* keys).
func serviceHourly(name string) float64 {
	loadPricing()
	return pricing.ServicesHourly[name]
}
