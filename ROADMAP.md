# Roadmap

## v0.1 — Foundation (current)

- [x] Steampipe-based resource discovery (AWS)
- [x] Architecture reconstruction (network + resource views)
- [x] 26 assessment rules (Security, Reliability, Cost, Operations)
- [x] Infrastructure sizing analysis
- [x] Cost estimation (AWS Cost Explorer + static fallback)
- [x] Bubble Tea TUI with 5 views
- [x] Report generation (Markdown, JSON, Excel)
- [x] Profile management and interactive configure wizard
- [x] SQLite persistence with historical assessments

## v0.2 — Accuracy and Coverage

- [ ] Well-Architected Framework pillar scoring (percentage per pillar)
- [ ] 50+ assessment rules (CloudTrail, Config, GuardDuty, WAF)
- [ ] OCI assessment rules (matching AWS coverage)
- [ ] Improved cost accuracy with data transfer and IOPS estimation
- [ ] ECS/Fargate resource discovery and sizing
- [ ] CloudWatch alarm coverage checks
- [ ] Tag compliance rules (required tags, naming conventions)

## v0.3 — Multi-Account and Comparison

- [ ] AWS Organizations support (assess multiple accounts)
- [ ] Assessment diff (compare two runs, show what changed)
- [ ] Trend tracking (cost and findings over time)
- [ ] Custom rule engine (user-defined rules via YAML)
- [ ] Remediation scripts (auto-generate fix commands)

## v0.4 — OCI Parity

- [ ] OCI resource discovery via Steampipe
- [ ] OCI architecture reconstruction
- [ ] OCI Well-Architected assessment rules
- [ ] OCI cost estimation
- [ ] Cross-cloud comparison reports

## v0.5 — Enterprise Features

- [ ] PDF report generation
- [ ] Scheduled assessments (cron-based)
- [ ] Webhook notifications (Slack, Teams)
- [ ] SARIF output for CI integration
- [ ] Compliance report templates (SOC2, ISO 27001, HIPAA)

## Future Considerations

- Azure support
- GCP support
- Terraform state import (assess without live API access)
- Policy-as-code integration (OPA/Rego)
- Team collaboration (shared assessment database)
