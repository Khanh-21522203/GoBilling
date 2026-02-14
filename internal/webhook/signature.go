package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

func GenerateSignature(secret string, timestamp int64, payload []byte) string {
	message := fmt.Sprintf("%d.%s", timestamp, payload)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

func VerifySignature(secret string, signature string, timestamp int64, payload []byte) bool {
	now := time.Now().Unix()
	if now-timestamp > 300 || timestamp-now > 300 {
		return false
	}

	expected := GenerateSignature(secret, timestamp, payload)
	return hmac.Equal([]byte(signature), []byte(expected))
}
