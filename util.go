package agentcore

import (
    "crypto/rand"
    "encoding/hex"
    "time"
)

func newID() string {
    b := make([]byte, 16)
    rand.Read(b)
    return hex.EncodeToString(b)
}

func now() int64 {
    return time.Now().UnixMilli()
}

func EstimateTokens(text string) int {
    return len(text) / 4
}
