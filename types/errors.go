package types

import "errors"

var (
	ErrInvalidConfig        = errors.New("invalid configuration")
	ErrSessionNotFound      = errors.New("session not found")
	ErrProviderNotFound     = errors.New("LLM provider not found")
	ErrToolNotFound         = errors.New("tool not found")
	ErrSessionAlreadyEnded  = errors.New("session already ended")
	ErrToolDisabled         = errors.New("tool is disabled")
	ErrAgentClosed          = errors.New("agent core is closed")
	ErrToolExecution        = errors.New("tool execution failed")
	ErrCompressionFailed    = errors.New("compression failed")
	ErrNoAvailableProviders = errors.New("no available LLM providers")
)
