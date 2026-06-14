package agentcore

import "errors"

var (
	// Config validation errors.
	ErrInvalidConfig = errors.New("invalid configuration")

	// Resource not found errors.
	ErrSessionNotFound  = errors.New("session not found")
	ErrProviderNotFound = errors.New("LLM provider not found")
	ErrToolNotFound     = errors.New("tool not found")

	// State errors.
	ErrSessionAlreadyEnded  = errors.New("session already ended")
	ErrToolDisabled         = errors.New("tool is disabled")
	ErrAgentClosed          = errors.New("agent core is closed")

	// Execution errors.
	ErrToolExecution         = errors.New("tool execution failed")
	ErrCompressionFailed     = errors.New("compression failed")
	ErrNoAvailableProviders  = errors.New("no available LLM providers")
)
