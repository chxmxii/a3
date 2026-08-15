# v0.2.0 — TUI Improvements & Codebase Overhaul

## New Features

### TUI
- **Inventory search**: press `/` to live-filter resources by name, ID, type, or region (case-insensitive, composes with region/type filters)
- **Help overlay**: press `?` for a full keybinding reference per view
- Animated loading spinner while assessment data loads

### Security
- **Open egress rules are now highlighted**: security group rules allowing all outbound traffic (0.0.0.0/0, ::/0) are flagged in the SG detail panel, matching the existing inbound highlight
- **Checklist expanded from 11 to 28 controls**: CIS AWS Foundations, AWS Well-Architected, cost, reliability, and operations controls now appear in checklist output (previously only the 3A Security Baseline controls were tracked, silently dropping findings from the other standards)

### Configuration
- `steampipe.connection_string` in `~/.a3/config.yaml` is now honored (precedence: `--steampipe-conn` flag > config > default) — previously written by `a3 configure` but never read

## Internal

- Codebase reduced ~17% (12.2k → 10.1k lines of Go) with no behavior regressions
- All AWS pricing/instance data moved to one embedded JSON catalog (`internal/cost/data/aws_pricing.json`) — prices can be updated without code changes
- 23 of 28 security rules converted to a data-driven table; adding a simple rule is now a ~10-line entry
- Unified storage scan layer, Steampipe query cascade, TUI scroll/key handling, and metadata accessors
- First unit tests for TUI logic (SG rule parsing, search, scroll clamping)

---

# v0.1.0 — Initial Release

First public release of 3A (Agnostic Account Assessment).

## What is 3A

A terminal-first tool that assesses cloud accounts (AWS, OCI) using Steampipe. One command gives you a full picture: resource inventory, architecture map, security findings, cost analysis, and actionable reports.

## Features

### Discovery
- 25+ AWS resource types via Steampipe SQL queries
- Three-tier fallback for restricted IAM roles (SELECT * → specific columns → minimal)
- Concurrent multi-region discovery

### Security Assessment
- 26 rules covering EC2, RDS, S3, Lambda, VPC, ALB, EBS, IAM, EKS
- Standards: CIS Benchmarks, AWS Well-Architected Framework
- Categories: Security, Reliability, Cost Optimization, Operational Excellence

### Architecture
- Split views: Network (VPC/Subnet/SG/Gateways) and Resource (ALB/EKS/RDS relationships)
- Color-coded resource types in tree visualization

### Cost Analysis
- Primary: real billing data from AWS Cost Explorer via Steampipe
- Fallback: static pricing catalog (90+ instance types)
- Idle and oversized resource detection

### TUI
- 5 interactive views: Overview, Inventory, Architecture, Findings, Cost
- Region and type filtering (r/R, t/T)
- Resource detail panels (Enter) with type-aware rendering for SGs and Route Tables
- Scrollable views with keyboard navigation

### Reports
- Markdown and JSON export
- Excel workbook with 5 sheets (Summary, Inventory, Findings, Cost, Architecture)
- Auto-filters and formatted columns

### CLI
- `a3 assess <profile>` — full pipeline with animated progress
- `a3 configure` — interactive setup wizard (credentials, Steampipe config, regions)
- `a3 report <profile> --format excel` — generate reports without TUI
- `a3 profiles list/add` — profile management

## Install

```bash
curl -fsSL https://chxmxii.github.io/3a/install.sh | sh
```

Or build from source:
```bash
go install github.com/chxmxii/3a/cmd/3a@v0.1.0
```

## Requirements

- Steampipe with the AWS plugin installed and configured
- Read-only AWS credentials (ReadOnlyAccess or custom policy)

## Known Limitations

- OCI discovery is defined but untested (awaiting OCI Steampipe plugin setup)
- Cost estimation accuracy depends on `ce:GetCostAndUsage` permission
- Some Steampipe hydrate columns require permissions beyond basic ReadOnly
- Property-based tests not yet implemented

## What's Next (v0.2)

- Well-Architected pillar scoring
- 50+ assessment rules
- Tag compliance checks
- Assessment diff (compare runs)
