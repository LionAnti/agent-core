package agentcore

import "github.com/LionAnti/agent-core/types"

type noopMetricsCollector struct{}

func (noopMetricsCollector) RecordLatency(name string, ms float64)         {}
func (noopMetricsCollector) RecordTokenUsage(prompt, completion int)        {}
func (noopMetricsCollector) IncrementCounter(name string, labels ...string) {}
func (noopMetricsCollector) SetGauge(name string, value float64, labels ...string) {}

type noopLogger struct{}

func (noopLogger) Debug(msg string, keysAndValues ...any) {}
func (noopLogger) Info(msg string, keysAndValues ...any)  {}
func (noopLogger) Warn(msg string, keysAndValues ...any)  {}
func (noopLogger) Error(msg string, keysAndValues ...any) {}

var _ types.Logger = (*noopLogger)(nil)
var _ types.MetricsCollector = (*noopMetricsCollector)(nil)
