# ADR 0001: Test Harness Composition

## Status
Accepted

## Context
The `scg-test-kit` library initially used an interface-based approach (`TestKit`) with a builder that returned a concrete implementation. While functional, it tended toward an inheritance-like pattern (even if via composition) where the user was expected to use a single "TestCase" or "IntegrationTest" object that had all possible components (DB, gRPC, Bus) as fixed fields.

As we scale, we want a more modular and extensible approach that avoids "kitchen sink" structs and follows Go's idiomatic composition-first philosophy.

## Decision
We will move to a `Harness` struct that uses **Composition** and **Capabilities**.

1. **Composition over Inheritance**:
   - We provide a single `Harness` struct.
   - It is NOT an interface; it is a concrete struct that holds optional capabilities.
   - It is initialized via functional options (`Option`).

2. **Capabilities (Plug-ins)**:
   - Capabilities are small, driver-agnostic interfaces (e.g., `Database`, `HttpServer`).
   - They are "plugged in" using options.
   - Discovery is safe: accessors return `(Capability, bool)` to avoid nil panics and clearly indicate if a capability is available.

3. **Presets over Hierarchies**:
   - Instead of different types for Unit, Integration, and Browser tests, we use convenience constructors (presets).
   - `NewUnitHarness`, `NewIntegrationHarness`, `NewBrowserHarness` are just functions that return the same `*Harness` type but with different default options.

4. **Resource Management**:
   - The `Harness` registers its own cleanup and its capabilities' cleanup with `testing.TB.Cleanup`.

## Consequences

### Positive
- **Extensible**: New capabilities can be added without changing the `Harness` core struct or breaking existing tests.
- **Discoverable**: The `(cap, ok)` pattern is idiomatic Go and prevents common nil-pointer bugs in tests.
- **Modular**: Heavy dependencies (like `testcontainers`) can be isolated in capability-specific providers, keeping the core harness light.

### Negative
- Slightly more verbose initialization if many options are needed (mitigated by presets).
- Requires a one-time migration from the old `TestKit` interface to the new `Harness` struct.

## Rules to prevent "kitchen sink" creep
- Each capability MUST have its own interface in `internal/contract`.
- The `Harness` struct MUST NOT have fields for specific drivers (e.g., no `Postgres` field; use `Database` capability).
- Accessors MUST follow the `(Capability, bool)` pattern.
