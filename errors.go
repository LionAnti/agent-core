package agentcore

import "errors"

var (
    ErrSessionNotFound      = errors.New("session not found")
    ErrSessionAlreadyEnded  = errors.New("session already ended")
    ErrProviderNotFound     = errors.New("LLM provider not found")
    ErrNoAvailableProviders = errors.New("no available LLM providers")
    ErrToolNotFound         = errors.New("tool not found")
    ErrToolDisabled         = errors.New("tool is disabled")
    ErrToolExecution        = errors.New("tool execution failed")
    ErrNoMessages           = errors.New("no messages to process")
    ErrCompressionFailed    = errors.New("compression failed")
    ErrInvalidConfig        = errors.New("invalid configuration")
)
