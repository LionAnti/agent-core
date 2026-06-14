package agentcore

func estimateMessagesTokenCount(msgs []Message) int {
	total := 0
	for _, m := range msgs {
		if m.TokenCount > 0 {
			total += m.TokenCount
		} else {
			total += len(m.Content)/4 + 1
		}
	}
	return total
}
