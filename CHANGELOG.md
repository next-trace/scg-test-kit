# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

