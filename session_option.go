package agentcore

type SessionOption func(*Session)

func WithModel(model string) SessionOption {
    return func(s *Session) { s.Model = model }
}
func WithProviderID(providerID string) SessionOption {
    return func(s *Session) { s.ProviderID = providerID }
}
func WithSessionID(id string) SessionOption {
    return func(s *Session) { s.ID = id }
}
