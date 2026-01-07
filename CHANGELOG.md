# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.2] - 2024-05-22

### Added
- `testkit.New(t)` as the primary entrypoint for creating a Harness.
- `h.RegisterCleanup(func())` for registering custom cleanup functions.
- `h.Close()` as an alias for `h.Cleanup()`.
- Automatic registration of `h.Cleanup()` with `t.Cleanup()`.

### Changed
- `Harness` now automatically registers its cleanup with `testing.TB`, removing the need for manual `h.Cleanup()` calls in tests.
- `NewHarness`, `NewUnitHarness`, and `NewIntegrationHarness` now delegate to `New` and benefit from automatic cleanup. They are preserved as semantic aliases.

## [Unreleased]

### Changed
- **BREAKING**: `Get()` and `Post()` functions no longer return `*http.Response` to avoid returning a response with a closed body. These functions now only decode the response into the provided target parameter.

### Added
- Initial CHANGELOG.md file to track project changes

## [0.1.0] - Initial Release

### Added
- Core `Harness` for managing test resource lifecycles
- Generic resource injection with `WithResource()`
- HTTP test server integration with `NewBrowserHarness()`
- Helper functions `Get()` and `Post()` for HTTP testing
- JSON encoding/decoding utilities
- LIFO cleanup management using `testing.TB.Cleanup`
- Comprehensive documentation and examples
