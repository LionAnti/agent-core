package agentcore

import "errors"

var (
	// Config validation
	ErrInvalidConfig = errors.New("invalid configuration")

	// Resource not found
	ErrSessionNotFound    = errors.New("session not found")
	ErrProviderNotFound   = errors.New("LLM provider not found")
	ErrToolNotFound       = errors.New("tool not found")

	// State errors
	ErrSessionAlreadyEnded  = errors.New("session already ended")
	ErrToolDisabled         = errors.New("tool is disabled")
	ErrAgentClosed          = errors.New("agent core is closed")

	// Execution errors
	ErrToolExecution     = errors.New("tool execution failed")
	ErrCompressionFailed = errors.New("compression failed")
	ErrNoMessages        = errors.New("no messages to process")

	// Pipeline
	ErrStagePanic        = errors.New("pipeline stage panicked")
	ErrNoAvailableProviders = errors.New("no available LLM providers")
)
