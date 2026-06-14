# agent-core

**Zero-dependency Go library** for building LLM-powered agent applications. Provides five core capabilities as a reusable package:

| Module | Description |
|--------|-------------|
| **Provider Management** | Multi-LLM routing, weighted selection, failover |
| **Context Engineering** | Adaptive compression (summary/mermaid/sliding window) |
| **Tool Registry** | Registration, concurrent-safe caching, intent-based selection |
| **Rules Engine** | 8-domain regex matching, scoring, priority sorting |
| **Dynamic Harness** | Intent classification, density estimation, offload decisions, utility tracking |

## Architecture

Users implement 8 interfaces to connect their infrastructure — no storage or transport dependencies are bundled.

```
┌─────────────────────────────────────────────────────────┐
│                    agent-core                            │
│  ┌────────┐ ┌──────────┐ ┌──────┐ ┌──────┐ ┌────────┐ │
│  │Provider│ │Compressor│ │Registry│ │Rules │ │Harness │ │
│  │Manager │ │  Engine  │ │  +Cache│ │Engine│ │Pipeline│ │
│  └───┬────┘ └────┬─────┘ └───┬──┘ └──┬───┘ └───┬────┘ │
│      │           │           │       │         │       │
└──────┼───────────┼───────────┼───────┼─────────┼───────┘
       │           │           │       │         │
  ProviderStore   LLMClient  RegistryStore RuleStore  ...
  (interface)    (interface) (interface)  (interface)
```

## Usage

```go
package main

import (
    "context"
    "fmt"
    "github.com/agent-core"
)

// Implement the LLMClient interface
type myLLM struct {}
func (m *myLLM) Chat(ctx context.Context, req *agentcore.ChatRequest) (*agentcore.ChatResponse, error) {
    // Wrap your LLM SDK here (OpenAI, Claude, DeepSeek, etc.)
    return &agentcore.ChatResponse{
        Content: "Hello from LLM!",
        Usage:   agentcore.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
    }, nil
}

func main() {
    core, err := agentcore.New(agentcore.Config{
        LLMClient:     &myLLM{},
        ProviderStore: &myProviderStore{},  // your implementation
        RegistryStore: &myRegistryStore{},  // your implementation
        MemoryStore:   &myMemoryStore{},    // your implementation
        Logger:        agentcore.NoopLogger{},
    })
    if err != nil {
        panic(err)
    }
    defer core.Close(context.Background())

    sess := core.NewSession("tenant-001", "user-001",
        agentcore.WithModel("deepseek-chat"))
    result, err := sess.Send(context.Background(), "Hello!", nil)
    if err != nil {
        panic(err)
    }
    fmt.Println(result.Response.Content)
}
```

## Pipeline

`Session.Send()` executes a 9-stage pipeline with full panic isolation and graceful degradation:

1. **Pre-rules** — Safety checks, routing hints
2. **Intent Classification** — Rule-based + LLM hybrid
3. **Tool Selection** — Intent-aware tool matching
4. **Memory Recall** — L1/L2/L3 memory retrieval (continues on failure)
5. **Density Estimation** — 6-signal context density analysis
6. **Offload Decision** — 50% mild / 85% aggressive thresholds
7. **Compression** — Summary, mermaid, or sliding window
8. **LLM Call** — Provider-routed request
9. **Post-rules** — Scoring and classification

## Interfaces

| Interface | Required | Purpose |
|-----------|----------|---------|
| `LLMClient` | Yes | Wrap any LLM SDK |
| `ProviderStore` | Yes | Persist LLM provider configs |
| `RegistryStore` | Yes | KV store for tool registration |
| `MemoryStore` | Yes | Persist L1-L3 memory |
| `L0Store` | No | Persist raw conversation records |
| `RuleStore` | No | Persist user-defined rules |
| `VectorStore` | No | Semantic memory search |
| `Logger` | No | Structured logging (default: noop) |
| `MetricsCollector` | No | Performance metrics (default: noop) |
