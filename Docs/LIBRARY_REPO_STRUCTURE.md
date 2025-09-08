# SCG Library Repository Structure Standard

This document defines the mandatory structure for all SCG libraries.

## 1. Root Directory Layout
- `Docs/`: Mandatory documentation folder (Capitalized).
- `examples/`: Usage examples (Go tests or code).
- `CHANGELOG.md`: Mandatory history of changes.
- `LICENSE`: Mandatory license file.
- `README.md`: Entry point documentation.
- `go.mod`: Go module definition.

## 2. Documentation Standards
- All architectural decisions must be recorded in `Docs/adr/`.
- ADRs must be numbered sequentially: `0001-title.md`.
- Public API must be documented in `Docs/public_api.md`.

## 3. Library Boundaries
- Libraries must not provision infrastructure.
- Lifecycle management must be handled via `testing.TB.Cleanup`.
