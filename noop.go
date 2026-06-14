package agentcore

type NoopMetricsCollector struct{}

func (NoopMetricsCollector) RecordLatency(name string, ms float64)         {}
func (NoopMetricsCollector) RecordTokenUsage(prompt, completion int)        {}
func (NoopMetricsCollector) IncrementCounter(name string, labels ...string) {}
func (NoopMetricsCollector) SetGauge(name string, value float64, labels ...string) {}

type NoopLogger struct{}
func (NoopLogger) Debug(msg string, keysAndValues ...any) {}
func (NoopLogger) Info(msg string, keysAndValues ...any)  {}
func (NoopLogger) Warn(msg string, keysAndValues ...any)  {}
func (NoopLogger) Error(msg string, keysAndValues ...any) {}
