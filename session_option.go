package agentcore

import "github.com/agent-core/types"

type SessionOption func(*types.Session)

func WithModel(model string) SessionOption {
	return func(s *types.Session) { s.Model = model }
}
func WithProviderID(providerID string) SessionOption {
	return func(s *types.Session) { s.ProviderID = providerID }
}
func WithSessionID(id string) SessionOption {
	return func(s *types.Session) { s.ID = id }
}
