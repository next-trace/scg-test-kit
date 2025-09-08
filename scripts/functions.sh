#!/usr/bin/env bash
# All logic lives here. ./scg sources this file.

# ===== Colors & logging =====
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[0;33m'; BLUE='\033[0;34m'; NC='\033[0m'
info() { echo -e "${GREEN}INFO:${NC} $*"; }
warn() { echo -e "${YELLOW}WARN:${NC} $*"; }
error() { echo -e "${RED}ERROR:${NC} $*" >&2; }
fail() { error "$*"; exit 1; }

# ===== CRITICAL: Ensure GOBIN is on PATH =====
# CI sets GOBIN to a workspace directory (cached).
# go install puts binaries there. If PATH doesn't include it, tools won't run.
fn_ensure_gobin_on_path() {
  if [[ -n "${GOBIN:-}" ]]; then
    case ":$PATH:" in
      *":${GOBIN}:"*) : ;;
      *) export PATH="${GOBIN}:${PATH}" ;;
    esac
  fi
}

fn_require_go() {
  local need="${GO_VERSION}"
  local have="$(go version | awk '{print $3}' | sed 's/go//')"
  [[ -n "$have" ]] || fail "Go not found"
  if ! [[ "$have" == "$need"* ]]; then
    fail "Go ${need} is required, but found ${have}. Please switch to the correct toolchain."
  fi
}

# ===== Go toolchain =====
fn_deps() {
  info "Downloading and tidying dependencies..."
  go mod download
  go mod tidy
  info "Dependencies updated."
}

# ===== Status Tracking =====
STATUS_FILE=".scg_status"

fn_set_status() {
  local check=$1
  local status=$2
  mkdir -p "$(dirname "$STATUS_FILE")"
  if [ ! -f "$STATUS_FILE" ]; then touch "$STATUS_FILE"; fi
  # Remove old status for this check
  grep -v "^${check}=" "$STATUS_FILE" > "${STATUS_FILE}.tmp" || true
  echo "${check}=${status}" >> "${STATUS_FILE}.tmp"
  mv "${STATUS_FILE}.tmp" "$STATUS_FILE"
}

fn_get_status() {
  local check=$1
  if [ -f "$STATUS_FILE" ]; then
    grep "^${check}=" "$STATUS_FILE" | cut -d'=' -f2
  else
    echo "unknown"
  fi
}

fn_reset_statuses() {
  rm -f "$STATUS_FILE"
}

fn_fmt() {
  local target="${1:-.}"
  fn_ensure_gobin_on_path
  info "Formatting Go files in $target using gofmt and goimports..."
  if ! command -v ${GOIMPORTS} &> /dev/null; then
    warn "${GOIMPORTS} not found, installing..."
    go install golang.org/x/tools/cmd/goimports@${GOIMPORTS_VERSION}
  fi
  
  if [[ -f "$target" ]]; then
    if [[ "$target" == *.go ]]; then
      gofmt -s -w "$target"
      ${GOIMPORTS} -w "$target"
    else
      fail "Error: $target is not a Go file."
    fi
  else
    gofmt -s -w "$target"
    ${GOIMPORTS} -w "$target"
  fi
  info "Code formatted."
}

fn_lint() {
  fn_ensure_gobin_on_path
  info "Running golangci-lint..."
  if ! command -v ${GOLANGCI_LINT} &> /dev/null; then
    warn "${GOLANGCI_LINT} not found, installing..."
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}
  fi
  # Ensure we are on correct golangci-lint version
  if ! ${GOLANGCI_LINT} version 2>/dev/null | grep -qE "^golangci-lint has version ${GOLANGCI_LINT_VERSION#v}([ \.]|$)"; then
    warn "Upgrading golangci-lint to ${GOLANGCI_LINT_VERSION}..."
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}
  fi
  if ${GOLANGCI_LINT} run --timeout=5m --concurrency=4; then
    fn_set_status "lint" "pass"
  else
    fn_set_status "lint" "fail"
    return 1
  fi
  info "Lint passed."
}

fn_lint_fix() {
  fn_ensure_gobin_on_path
  info "Running golangci-lint and fixing issues..."
  if ! command -v ${GOLANGCI_LINT} &> /dev/null; then
    warn "${GOLANGCI_LINT} not found, installing..."
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}
  fi
  # Ensure we are on correct golangci-lint version
  if ! ${GOLANGCI_LINT} version 2>/dev/null | grep -qE "^golangci-lint has version ${GOLANGCI_LINT_VERSION#v}([ \.]|$)"; then
    warn "Upgrading golangci-lint to ${GOLANGCI_LINT_VERSION}..."
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}
  fi
  if ${GOLANGCI_LINT} run --fix --timeout=5m --concurrency=4; then
    fn_set_status "lint" "pass"
  else
    fn_set_status "lint" "fail"
    return 1
  fi
  info "Linting and fixing completed."
}

fn_build() {
  info "Building code..."
  if go build -v ./...; then
    fn_set_status "build" "pass"
  else
    fn_set_status "build" "fail"
    return 1
  fi
  info "Build successful!"
}

fn_test() {
  local target=( "$@" )
  if [[ ${#target[@]} -eq 0 ]]; then
    target=( "./..." )
  fi
  info "Running tests for ${target[*]}..."
  if go test -race -v -parallel 4 -coverprofile=coverage.txt -covermode=atomic "${target[@]}"; then
    fn_set_status "test" "pass"
  else
    fn_set_status "test" "fail"
    return 1
  fi
  info "Tests passed."
}

fn_check_coverage() {
  local threshold=${1:-90}
  if [[ -f "coverage.txt" ]]; then
    local total_line=$(go tool cover -func=coverage.txt | tail -n 1)
    info "Coverage summary: $total_line"
    local pct=$(echo "$total_line" | awk '{print $3}' | tr -d '%')
    # Use printf to handle float to int conversion safely in bash
    local pct_int=$(printf "%.0f" "$pct" 2>/dev/null || echo "0")
    if [[ "$pct_int" -lt "$threshold" ]]; then
      fn_set_status "coverage" "fail"
      fail "Coverage $pct% is below required threshold ${threshold}%"
    fi
    fn_set_status "coverage" "pass"
    info "Coverage threshold check passed ($pct% >= ${threshold}%)."
  else
    warn "coverage.txt not found, skipping threshold check."
  fi
}

fn_bench() {
  local target="${1:-./...}"
  info "Running benchmarks for $target..."
  go test -bench=. -benchmem "$target"
  info "Benchmarks completed."
}

fn_security() {
  fn_ensure_gobin_on_path
  info "Running security checks..."
  if ! command -v ${GOVULNCHECK} &> /dev/null; then
    warn "${GOVULNCHECK} not found, installing..."
    go install golang.org/x/vuln/cmd/govulncheck@${GOVULNCHECK_VERSION}
  fi
  if ! command -v ${GOSEC} &> /dev/null; then
    warn "${GOSEC} not found, installing..."
    go install github.com/securego/gosec/v2/cmd/gosec@${GOSEC_VERSION}
  fi
  info "Running govulncheck..."
  # Skip example directory as it might have different dependencies or build tags
  local PKGS=$(go list ./... | grep -Ev '/(example|cmd)($|/)' || true)
  local sec_fail=false
  if [[ -n "${PKGS}" ]]; then
    if ! ${GOVULNCHECK} -mode=source ${PKGS}; then
      sec_fail=true
    fi
  fi
  info "Running gosec..."
  if ! ${GOSEC} -quiet -exclude-generated -exclude-dir=.git -exclude-dir=.github -exclude-dir=example -exclude-dir=cmd ./...; then
    sec_fail=true
  fi
  
  if $sec_fail; then
    fn_set_status "security" "fail"
    return 1
  else
    fn_set_status "security" "pass"
  fi
  info "Security checks passed."
}

fn_clean() {
  info "Cleaning build and test cache..."
  go clean -cache -testcache -modcache
  if [[ -f "coverage.out" ]]; then rm -f coverage.out; fi
  if [[ -f "coverage.html" ]]; then rm -f coverage.html; fi
  if [[ -f "coverage.txt" ]]; then rm -f coverage.txt; fi
  info "Cache cleaned successfully."
}

fn_coverage() {
  info "Generating coverage report..."
  go test -coverprofile=coverage.out ./...
  go tool cover -html=coverage.out -o coverage.html
  info "Coverage report generated: coverage.html"
  if command -v xdg-open &> /dev/null; then
    xdg-open coverage.html
  elif command -v open &> /dev/null; then
    open coverage.html
  else
    warn "Please open coverage.html in your browser to view the report"
  fi
}

fn_docs() {
  info "Generating documentation..."
  go doc -all ./... > docs.txt
  info "Documentation generated: docs.txt"
}

# ===== CI bundle =====
fn_ci() {
  info "Running CI checks locally..."
  fn_reset_statuses
  fn_require_go

  # Use a subshell or a pattern that allows capturing failure without exiting the whole script
  # if set -e is active in the caller.
  local failed=0

  fn_guardrails || failed=1
  
  if (( failed == 0 )); then
    fn_build || failed=1
  fi

  if (( failed == 0 )); then
    # Determine packages to include in coverage (exclude example and cmd by default)
    local -a pkgs
    mapfile -t pkgs < <(go list ./... | grep -Ev '/(example|cmd)($|/)' || true)
    if ((${#pkgs[@]} == 0)); then
      pkgs=(./...)
    fi

    # Join all packages into a single argument so fn_test (which uses only $1) sees them all
    # NOTE: We pass each package as a separate argument to support multiple packages correctly
    fn_test "${pkgs[@]}" || failed=1
  fi

  if (( failed == 0 )); then
    fn_check_coverage 80 || failed=1
  fi

  if (( failed == 0 )); then
    fn_lint || failed=1
  fi

  if (( failed == 0 )); then
    fn_security || failed=1
  fi

  if (( failed > 0 )); then
    error "❌ CI checks FAILED. See summary below."
    fn_summary
    return 1
  fi

  info "✅ All CI checks passed successfully!"
  fn_summary
}

# ===== Doctor checks =====
fn_guardrails() {
  info "Running Guardrails..."
  if bash scripts/guardrails.sh; then
    fn_set_status "guardrails" "pass"
    return 0
  else
    fn_set_status "guardrails" "fail"
    return 1
  fi
}

fn_doctor() {
  info "Running health checks..."
  fn_require_go
  info "✓ Go ${GO_VERSION} found"
  fn_guardrails
  info "✅ Doctor checks complete."
}

fn_summary() {
  local build_st=$(fn_get_status "build")
  local test_st=$(fn_get_status "test")
  local lint_st=$(fn_get_status "lint")
  local sec_st=$(fn_get_status "security")
  local guard_st=$(fn_get_status "guardrails")
  local cov_st=$(fn_get_status "coverage")

  status_icon() {
    case "$1" in
      "pass") echo "✅" ;;
      "fail") echo "❌" ;;
      *)      echo "⚪" ;;
    esac
  }

  echo "## 🛡️ SCG Quality Gate Summary"
  echo ""
  echo "> **SupplyChainGuard Standard Test Kit Library**"
  echo ""
  echo "### 📊 Quality Check Results"
  echo "| Check | Status | Description |"
  echo "| :--- | :---: | :--- |"
  echo "| 🛡️ **Guardrails** | $(status_icon "$guard_st") | Repo hygiene and purity |"
  echo "| 🏗️ **Build** | $(status_icon "$build_st") | Library compiles successfully |"
  echo "| 🧪 **Tests** | $(status_icon "$test_st") | All tests passed (Race Detection: ON) |"
  echo "| 📈 **Coverage** | $(status_icon "$cov_st") | Code coverage threshold met |"
  echo "| 🧹 **Lint** | $(status_icon "$lint_st") | Code style and quality verified |"
  echo "| 🔒 **Security** | $(status_icon "$sec_st") | Vulnerability and static analysis clean |"
  echo ""
  echo "### 🛠️ Environment Details"
  echo "- **Go Version:** \`$(go version | awk '{print $3}')\`"
  echo "- **Runner OS:** \`${RUNNER_OS:-Linux}\`"
  echo "- **Workflow:** \`${GITHUB_WORKFLOW:-scg-test-kit-ci}\`"
  echo "- **Commit:** \`${GITHUB_SHA:-local}\`"
  echo ""
  echo "---"
  echo "Generated by **scg-test-kit** Quality Gate at $(date -u '+%Y-%m-%d %H:%M:%S') UTC"
}

fn_install_tools() {
  fn_ensure_gobin_on_path
  info "Installing required tools..."
  go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}
  go install golang.org/x/vuln/cmd/govulncheck@${GOVULNCHECK_VERSION}
  go install github.com/securego/gosec/v2/cmd/gosec@${GOSEC_VERSION}
  go install golang.org/x/tools/cmd/goimports@${GOIMPORTS_VERSION}
  info "Tools installed."
}

# ===== Export functions =====
export -f info warn error fail fn_ensure_gobin_on_path