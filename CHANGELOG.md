# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

## [1.1.0] - 2026-05-07

### Added

- New `@Security` annotation for specifying security requirements
- Support for preserving existing comments when generating annotations
- Added function name as documentation title in generated swagger annotations
- New `--path` flag to specify target file path for generated swagger annotations

### Changed

- Changed config filename from `goswaggen.yaml` to `.goswaggen.yaml`

### Fixed

- Returning nothing when the response is generic type
- Showing `-` instead of showing nothing for `-` tag response in swagger documentation
- Returning nothing when the field is `time.Time` type
- Missing `@Accept` for `c.FormValue` and `c.FormFile` parameters

## [1.0.0] - 2025-07-15

### Added

- Initial release
- Basic swagger generation from Go source files
