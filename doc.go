// Package agentcore provides a zero-dependency library for LLM-powered agent applications.
//
// It integrates five core capabilities:
//  1. LLM Provider Management - routing, weighted selection, failover
//  2. Context Engineering - compression (summary, mermaid, sliding window)
//  3. Tool Registry - registration, discovery, concurrent-safe cache
//  4. Rules Engine - 8-domain rule matching, regex, scoring
//  5. Dynamic Harness - intent classification, density estimation, offload decisions
//
// Users implement the interfaces (LLMClient, MemoryStore, ProviderStore, etc.)
// to connect their own infrastructure (PostgreSQL, Redis, etcd, ZK, etc.).
package agentcore
