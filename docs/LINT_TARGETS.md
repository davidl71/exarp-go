# Lint Targets Reference

All lint-related Make targets in one place.

## Go Code Quality

| Target | Description |
|--------|-------------|
| `make fmt` | Format code with exarp-go (gofmt/goimports) |
| `make lint` | Lint code with exarp-go (auto-detect linter) |
| `make lint-fix` | Lint and auto-fix code with exarp-go |
| `make lint-all` | Lint everything: Go + docs + .cursor/plans + navigability |
| `make lint-all-fix` | Lint and fix everything: Go + docs + .cursor/plans |
| `make go-fmt` | Format Go code with gofmt directly (no exarp-go binary needed) |
| `make go-vet` | Check Go code with go vet |
| `make golangci-lint-check` | Check code with golangci-lint |
| `make golangci-lint-fix` | Fix code with golangci-lint (auto-fix) |

## Shell, YAML, Ansible

| Target | Description | Requires |
|--------|-------------|----------|
| `make lint-shellcheck` | Run shellcheck on `scripts/*.sh` and `ansible/run-dev-setup.sh` | `shellcheck` |
| `make lint-yaml` | Run yamllint on `.github` and `ansible` directories | `yamllint` |
| `make lint-ansible` | Run ansible-lint on playbooks/roles (run `make ansible-galaxy` first) | `ansible-lint` |

## AI Navigability

| Target | Description |
|--------|-------------|
| `make lint-nav` | Check file comments, package docs, magic strings |
| `make lint-nav-fix` | Same with suggested fixes for each issue |

## Security

| Target | Description |
|--------|-------------|
| `make govulncheck` | Check for Go vulnerabilities with govulncheck |
| `make check-security` | Run exarp-go security scan (no govulncheck) |

## Combined Targets

| Target | Description |
|--------|-------------|
| `make check` | `go-fmt` + `go-vet` + `golangci-lint-check` |
| `make check-fix` | `go-fmt` + `golangci-lint-fix` |
| `make check-all` | `check` + `govulncheck` |
| `make scorecard-fix` | `go-mod-tidy` + `fmt` + `lint-fix` (auto-fix all fixable scorecard issues) |

## CI Integration

The GitHub Actions workflow (`.github/workflows/go.yml`) runs:
- **lint** job: golangci-lint
- **vet** job: go vet + gofmt check
- **lint-extras** job: shellcheck + yamllint + ansible-lint
- **security** job: govulncheck
