# ADR 0002: Generic Resource + Cleanup System

## Status
Accepted

## Context
The previous version of `scg-test-kit` had hardcoded capabilities for specific technologies like databases (PostgreSQL) and HTTP servers. This led to:
1. Vendor lock-in and tight coupling to specific infrastructure libraries.
2. Inability to easily support new types of resources (Kafka, Redis, etc.) without modifying the core library.
3. Bloated dependencies in the core module.

## Decision
We decided to implement a technology-agnostic "Harness" that supports:
1. Injecting arbitrary resources as `any` values.
2. Registering cleanup callbacks for each resource.
3. Using `testing.TB.Cleanup` for reliable orchestration (LIFO order).

Key design points:
- `WithResource(name string, value any, cleanup func() error)` is the canonical way to inject dependencies.
- `Resource[T](h, name)` provides type-safe retrieval.
- `scg-test-kit` does NOT provision infrastructure; services are responsible for starting their dependencies.

## Consequences
- **Pros**:
  - Zero infrastructure dependencies in the core library.
  - Future-proof: can support any current or future technology.
  - Consistent lifecycle management across all tests.
  - Compliance with the Open-Closed Principle (OCP).
- **Cons**:
  - Services need to provide their own provisioning logic (e.g., using `testcontainers-go` or local mocks).
  - Retrieval requires naming conventions for resources.
