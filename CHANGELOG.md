# Changelog

## [0.1.0] - 2026-06-14

### Added
- Initial release of `agent-core` library
- Provider management: multi-LLM routing, weighted selection, failover
- Context engineering: summary/mermaid/sliding window compression
- Tool registry: registration, concurrent-safe caching, intent-based selection
- Rules engine: 8-domain regex matching, scoring, priority sorting
- Dynamic harness: intent classification, density estimation, offload decisions
- Memory pipeline: L1 extraction, L2 scenes, L3 persona

### Fixed
- Data race in DensityEstimator (shared state -> per-session HarnessState)
- Goroutine leak in pipeline stage execution
- Context leak in ensureTimeout
- Session map unbounded growth
- RulesEngine slice concurrent access (added RWMutex)
- Zero-value crash safety on all exported types
- regexp.MustCompile panic from user-supplied patterns

### Changed
- Refactored from flat package to 7 subpackages (types/rules/provider/compressor/registry/harness/memory)
- Types moved from string to typed constants (IntentType, RuleDomain, etc.)
