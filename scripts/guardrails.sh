#!/usr/bin/env bash
# Guardrails script to prevent committing unwanted files and patterns
# This script MUST fail if any forbidden artifacts are detected

set -euo pipefail

# Ensure we are in the project root
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

echo "=== Guardrails Check ==="

FAILED=false

# Check for tracked junk files
echo "Checking for tracked junk files..."
JUNK_FILES=$(git ls-files | grep -E '(\.idea/|coverage\.|\.db$|\.test$|^bin/|^dist/|^build/|\.bak$)' || true)
if [[ -n "${JUNK_FILES}" ]]; then
  echo "❌ FAIL: Forbidden files are tracked in git:"
  echo "${JUNK_FILES}"
  echo ""
  echo "These files should be in .gitignore and removed from git."
  echo "To fix: git rm <file> or git rm -r <directory>"
  FAILED=true
else
  echo "✓ No junk files tracked in git"
fi

# Check for common destructive patterns in scripts (basic check)
echo ""
echo "Checking for potentially destructive commands in scripts..."
DESTRUCTIVE_PATTERNS=$(grep -rn --include="*.sh" --include="*.yml" --include="*.yaml" -E 'rm\s+-rf\s+/|sudo\s+rm\s+-rf' . || true)
if [[ -n "${DESTRUCTIVE_PATTERNS}" ]]; then
  echo "⚠️  WARNING: Found potentially destructive patterns:"
  echo "${DESTRUCTIVE_PATTERNS}"
  echo ""
  echo "Review these carefully before committing."
  # Note: This is a warning, not a failure - some scripts may legitimately need rm -rf
fi

# Check for accidentally committed .env files (other than .env.test, .env.example, etc.)
echo ""
echo "Checking for accidentally committed .env files..."
ENV_FILES=$(git ls-files | grep -E '^\.env$|/\.env$' || true)
if [[ -n "${ENV_FILES}" ]]; then
  echo "❌ FAIL: .env file is tracked in git:"
  echo "${ENV_FILES}"
  echo ""
  echo ".env files often contain secrets and should never be committed."
  echo "Use .env.example or .env.test for templates."
  FAILED=true
else
  echo "✓ No .env files tracked"
fi

# Check for common binary artifacts
echo ""
echo "Checking for binary artifacts..."
BINARY_ARTIFACTS=$(git ls-files | grep -E '\.(exe|dll|so|dylib|a)$' || true)
if [[ -n "${BINARY_ARTIFACTS}" ]]; then
  echo "⚠️  WARNING: Binary files tracked in git:"
  echo "${BINARY_ARTIFACTS}"
  echo ""
  echo "Binary files should generally not be committed to a Go library repo."
  # This is a warning for now, as some repos may intentionally track test fixtures
fi

# Check for forbidden CLI framework imports in internal/ packages
echo ""
echo "Checking internal/ package purity (no CLI framework dependencies)..."
FORBIDDEN_PATTERNS='github.com/spf13/cobra\|github.com/spf13/viper\|gopkg.in/yaml\|github.com/joho/godotenv'
FORBIDDEN_INTERNAL_IMPORTS=$(find internal -name "*.go" -type f -exec grep -l "${FORBIDDEN_PATTERNS}" {} + 2>/dev/null || true)
if [[ -n "${FORBIDDEN_INTERNAL_IMPORTS}" ]]; then
  echo "❌ FAIL: internal/ packages import CLI frameworks:"
  echo "${FORBIDDEN_INTERNAL_IMPORTS}"
  echo ""
  echo "The internal/ packages MUST remain pure (no cobra/viper/yaml/godotenv)."
  FAILED=true
else
  echo "✓ internal/ packages are pure (no CLI framework imports)"
fi

# Public API Guard
echo ""
echo "Checking Public API Guard..."
MODULE_PATH="$(go list -m)"
# For scg-test-kit, we allow the root package and postgres package to import from internal.
ALLOWLIST=(
  "${MODULE_PATH}"
  "${MODULE_PATH}/postgres"
)
INTERNAL_PREFIX="${MODULE_PATH}/internal/"

pkgs="$(go list ./... 2>/dev/null || true)"
violations=0
while IFS= read -r pkg; do
  imports="$(go list -f '{{ join .Imports "\n" }}' "$pkg" 2>/dev/null || true)"
  [[ -z "${imports}" ]] && continue
  while IFS= read -r imp; do
    if [[ "${imp}" == "${MODULE_PATH}"* ]]; then
      # Rule 1: Allowlisted packages (the public ones) are authorized to use internals
      if [[ "${pkg}" == "${MODULE_PATH}" || "${pkg}" == "${MODULE_PATH}/postgres" ]]; then
         if [[ "${imp}" == "${INTERNAL_PREFIX}"* ]]; then
            continue
         fi
      fi
      
      # Rule 2: Internal -> Internal
      [[ "${pkg}" == "${INTERNAL_PREFIX}"* && "${imp}" == "${INTERNAL_PREFIX}"* ]] && continue
      
      # Rule 4: Global Restriction - No other public package can import internal
      if [[ "${pkg}" != "${INTERNAL_PREFIX}"* && "${pkg}" != "${MODULE_PATH}" && "${pkg}" != "${MODULE_PATH}/postgres" ]]; then
        if [[ "${imp}" == "${INTERNAL_PREFIX}"* ]]; then
          echo "❌ Illegal internal import: Public package '${pkg}' is not allowed to import '${imp}'"
          echo "   Only allowed public packages are authorized to use internal utilities."
          violations=$((violations + 1))
          continue
        fi
      fi
      
      # Rule 5: CLI tools -> Public (Must only use public API)
      if [[ "${pkg}" == "${MODULE_PATH}/cmd/"* && "${imp}" == "${INTERNAL_PREFIX}"* ]]; then
        echo "❌ Illegal internal import in CLI tool: '${pkg}' importing '${imp}'"
        echo "   CLI packages under 'cmd/' must only depend on public APIs."
        violations=$((violations + 1))
        continue
      fi
    fi
  done <<< "${imports}"
done <<< "${pkgs}"

if [[ "${violations}" -gt 0 ]]; then
  echo "❌ Public API Guard FAILED: ${violations} violation(s)."
  FAILED=true
else
  echo "✓ Public API Guard OK"
fi

echo ""
if $FAILED; then
  echo "❌ Guardrails check FAILED"
  echo "Please fix the issues above and try again."
  exit 1
else
  echo "✅ All guardrails checks passed"
  exit 0
fi
