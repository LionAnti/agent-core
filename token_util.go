package agentcore

func estimateMessagesTokenCount(msgs []Message) int {
    total := 0
    for _, m := range msgs {
        if m.TokenCount > 0 {
            total += m.TokenCount
        } else {
            total += len(m.Content) / 4
        }
    }
    return total
}

func derefMemories(in []*L1Memory) []L1Memory {
	out := make([]L1Memory, len(in))
	for i, m := range in {
		if m != nil {
			out[i] = *m
		}
	}
	return out
}
