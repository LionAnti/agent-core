# agent-core

**Zero-dependency Go library** for building LLM-powered agent applications.

[![Go](https://github.com/LionAnti/agent-core/actions/workflows/ci.yml/badge.svg)](https://github.com/LionAnti/agent-core/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/LionAnti/agent-core)](https://pkg.go.dev/github.com/LionAnti/agent-core)

## Modules

| Package | Import | Description |
|---------|--------|-------------|
| `agentcore` | `github.com/LionAnti/agent-core` | Orchestrator: New(), Config, Session |
| `types` | `.../types` | Shared types, 16 user interfaces, errors |
| `rules` | `.../rules` | Regex rule engine (Match, AddRule, ListRules) |
| `rules` | `.../rules` | Regex rule engine (Match, AddRule, ListRules) |
| `provider` | `.../provider` | LLM provider routing + weighted selection |
| `provider` | `.../provider` | LLM provider routing + weighted selection |
| `compressor` | `.../compressor` | Context compression (summary/mermaid/sliding) |
| `registry` | `.../registry` | Tool registration + intent-based selection |
| `harness` | `.../harness` | Dynamic pipeline (classifier, density, offload, utility) |
| `harness` | `.../harness` | Dynamic pipeline (classifier, density, offload, utility) |
| `memory` | `.../memory` | Memory pipeline (L1 extraction + recall) |

> The package is at `github.com/LionAnti/agent-core`, NOT `github.com/agent-core`.

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "github.com/LionAnti/agent-core"
    "github.com/LionAnti/agent-core/types"
)

type myLLM struct{}
func (m *myLLM) Chat(ctx context.Context, req *types.ChatRequest) (*types.ChatResponse, error) {
    return &types.ChatResponse{Content: "Hello!"}, nil
}

// ... implement ProviderStore, RegistryStore, MemoryStore (see examples/)

func main() {
    core, _ := agentcore.New(agentcore.Config{
        LLMClient:     &myLLM{},
        ProviderStore: &myProviderStore{},
        RegistryStore: &myRegistryStore{},
        MemoryStore:   &myMemoryStore{},
    })
    defer core.Close(context.Background())

    sess, _ := core.NewSession("tenant-1", "user-1",
        agentcore.WithModel("gpt-4"))
    defer sess.Close()

    result, _ := sess.Send(context.Background(), "Hello!", nil)
    fmt.Println(result.Response.Content)
}
```

See [examples/basic](examples/basic/) and [examples/advanced](examples/advanced/) for full runnable examples.

## Pipeline

```
Send() → Pre-rules → Intent Classification → Tool Selection →
Memory Recall → Density Estimation → Offload Decision →
Compression (if needed) → LLM Call → Post-rules
```

## Interfaces (16 total)

| Interface | Methods | Required | Storage Example |
|-----------|---------|----------|----------------|
| `LLMClient` | 1 | Yes | OpenAI/Claude SDK wrapper |
| `ProviderStore` | 5 | Yes | PostgreSQL/etcd providers table |
| `RegistryStore` | 4 | Yes | etcd/Redis/ZK KV |
| `MemoryStore` | 15 | Yes | PostgreSQL/tidb |
| `L0Store` | 2 | No | PostgreSQL |
| `RuleStore` | 3 | No | etcd/YAML file |
| `VectorStore` | 3 | No | Milvus/pgvector |
| `Logger` | 4 | No | zap/logrus (default: noop) |
| `MetricsCollector` | 4 | No | Prometheus (default: noop) |
| `StreamLLMClient` | 2 | No | Extends LLMClient with streaming |

## Architecture

```
agentcore (orchestrator)
  ├── types/    — shared data types + interfaces
  ├── rules/    — regex-based rule matching
  ├── provider/ — LLM provider routing
  ├── compressor/ — context compression
  ├── registry/ — tool registration + selection
  ├── harness/  — dynamic harness pipeline
  └── memory/   — memory extraction + recall
```

## Testing

```bash
go test ./...              # all tests
go test -race -count=1 ./... # race detector
go test -bench=. -benchmem ./... # benchmarks
```

## License

MIT
