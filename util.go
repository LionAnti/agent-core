package agentcore

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

func newID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

func now() int64 {
	return time.Now().UnixMilli()
}

func EstimateTokens(text string) int {
	if len(text) == 0 {
		return 0
	}
	return len(text)/4 + 1
}
