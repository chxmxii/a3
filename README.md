<p align="center">
  <img src="docs/assets/logo.png" alt="A3" width="180">
</p>

# A3 / Cloud-Agnostic Account Assessment

Terminal-based cloud account assessment tool. Discovers resources, maps architecture, evaluates security posture, estimates costs, and generates reports.

Currently supports AWS and OCI via Steampipe.

## Requirements

- Steampipe with the AWS or OCI plugin installed and configured
- Valid cloud credentials (read-only access is sufficient)
- Go 1.24+

## Install

One-line install (Linux and macOS):

```bash
curl -fsSL https://chxmxii.github.io/a3/install.sh | bash
```

Using Go:

```bash
go install github.com/chxmxii/a3/cmd/a3@latest
```

Using Docker/Podman:

```bash
docker pull ghcr.io/chxmxii/a3:latest
docker run --rm -v ~/.aws:/root/.aws -v ~/.a3:/root/.a3 ghcr.io/chxmxii/a3 assess <profile>
```

From source:

```bash
git clone https://github.com/chxmxii/a3.git
cd a3
go build -o bin/a3 ./cmd/a3/ 
sudo cp bin/a3 /usr/local/bin/
```

## Usage

```sh
# Setup Wizard
a3 configure

# Run a full assessment
a3 assess <profile-name>

# Run without TUI
a3 assess <profile-name> --no-tui

# Generate reports
a3 report <profile-name> --format <markdown|json|excel> -o report.<md,json,xlsx>

# Profile management
a3 profiles list
a3 profiles add production --provider aws --regions us-east-1,eu-west-1 --aws-profile prod-readonly
```

## What it does

1. Connects to Steampipe and queries 25+ cloud resource types via SQL
2. Reconstructs architecture as two views: Network and Resources.
3. Evaluates 26 security rules against CIS Benchmarks and Well-Architected Framework
4. Fetches real billing data from AWS Cost Explorer (falls back to static estimates)
5. Calculates infrastructure sizing (vCPUs, memory, storage)
6. Generates an adaptive checklist (PASS/FAIL/WARN)
7. Displays everything in an interactive TUI or exports as Markdown/JSON/Excel

## TUI Controls

| Key | Action |
|-----|--------|
| 1-5 | Switch views (Overview, Inventory, Architecture, Findings, Cost) |
| j/k | Scroll up/down |
| Enter | View resource details (type-aware for SGs, Route Tables, IAM Policies) |
| / | Search inventory (type to filter live, Enter to apply, Esc to cancel) |
| Esc/x | Close detail panel / clear filters and search |
| r/R | Cycle regions (Inventory) |
| t/T | Cycle resource types (Inventory) |
| n/v | Switch Network/Resource architecture view |
| c/h/m/l | Filter by severity (Findings) |
| ? | Toggle help overlay |
| q | Quit (Ctrl+C also works during search) |

## Configuration

Profiles are stored in `~/.a3/config.yaml`:

```yaml
db_path: ~/.a3/a3.db
steampipe:
  connection_string: postgres://steampipe@localhost:9193/steampipe
profiles:
  - name: production
    provider: aws
    aws_profile: prod-readonly
    regions:
      - us-east-1
      - eu-west-1
```

Assessment data is stored in SQLite at `~/.a3/a3.db`.

## Steampipe Setup

```bash
steampipe plugin install <plugin>
steampipe service start
```

Configure credentials in `~/.steampipe/config/aws.spc`:

```hcl
connection "aws" {
  plugin  = "aws"
  profile = "your-aws-profile"
  regions = ["*"] # All regions
}
```

Or use `a3 configure` to set everything up with a3 wizard.

## License

MIT
