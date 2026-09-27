package audit

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func hmacSignature(secret, timestamp string, body []byte) string {
	h := hmac.New(sha256.New, []byte(secret))
	_, _ = h.Write([]byte(timestamp))
	_, _ = h.Write([]byte("."))
	_, _ = h.Write(body)
	return hex.EncodeToString(h.Sum(nil))
}
