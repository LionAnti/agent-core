
// Package agentcore provides a zero-dependency library for LLM-powered agent applications.
//
// It integrates five core capabilities:
//  1. LLM Provider Management - routing, weighted selection, failover
//  2. Context Engineering - compression (summary, mermaid, sliding window)
//  3. Tool Registry - registration, discovery, caching, selection
//  4. Rules Engine - intent classification, safety, memory recall rules
//  5. Dynamic Harness - intent classification, density estimation, offload decisions, utility tracking
//
// Users implement the interfaces (LLMClient, MemoryStore, ProviderStore, etc.)
// to connect their own infrastructure (PostgreSQL, Redis, etcd, ZK, etc.).
//
// Usage:
//
//   core, err := agentcore.New(agentcore.Config{
//       LLMClient:     myOpenAIWrapper,
//       ProviderStore: myProviderDB,
//       MemoryStore:   myPgStore,
//       Logger:        myLogger,
//   })
//   sess := core.NewSession("tenant-001", "user-001")
//   result, err := sess.Send(ctx, "Hello!", nil)
package agentcore
