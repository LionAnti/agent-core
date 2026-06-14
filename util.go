package agentcore

import (
	"crypto/rand"
	"encoding/hex"
	"time"
	"github.com/LionAnti/agent-core/types"
)

func newID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return time.Now().Format("150405.000")
	}
	return hex.EncodeToString(b)
}

func now() int64 {
	return time.Now().UnixMilli()
}

func estimateTokens(text string) int {
	if len(text) == 0 { return 0 }
	return len(text)/4 + 1
}

func estimateMessagesTokenCount(msgs []types.Message) int {
	total := 0
	for _, m := range msgs {
		if m.TokenCount > 0 { total += m.TokenCount } else { total += estimateTokens(m.Content) }
	}
	return total
}
