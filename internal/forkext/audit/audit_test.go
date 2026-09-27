package audit

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSanitizeMetadataDropsSecretsAndBoundsStrings(t *testing.T) {
	got := sanitizeMetadata(map[string]any{
		"password":      "never-store",
		"Authorization": "Bearer never-store",
		"safe":          "ok",
		"large":         strings.Repeat("x", 300),
	})
	if strings.Contains(got, "never-store") || strings.Contains(got, "large") || !strings.Contains(got, "safe") {
		t.Fatalf("sanitized metadata = %s", got)
	}
}

func TestHMACSignatureCanonical(t *testing.T) {
	got := hmacSignature("secret", "123", []byte(`{"event":"ok"}`))
	if got != "4043677fdef161fddbda3afb400407eb15cc90b168ed4fcbeee95f5c616b3c24" {
		t.Fatalf("signature = %s", got)
	}
}

func TestWebhookURLRejectsPrivateAndNonHTTPS(t *testing.T) {
	for _, raw := range []string{"http://example.com", "https://127.0.0.1", "https://10.0.0.1", "https://[::1]", "https://user:pass@example.com"} {
		if err := publicHTTPS(raw); err == nil {
			t.Fatalf("publicHTTPS(%q) accepted unsafe URL", raw)
		}
	}
}

func TestSecretRoundTripUses0600KeyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "master.key")
	t.Setenv("XUI_AUDIT_MASTER_KEY_FILE", path)
	stored, err := encryptSecret("webhook-secret")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(stored, encryptedPrefix) {
		t.Fatalf("ciphertext prefix = %q", stored)
	}
	plain, err := decryptSecret(stored)
	if err != nil || plain != "webhook-secret" {
		t.Fatalf("round trip = %q, %v", plain, err)
	}
	data, err := os.ReadFile(path)
	if err != nil || len(data) != 32 {
		t.Fatalf("key file = %d bytes, %v", len(data), err)
	}
	if encoded := base64.RawStdEncoding.EncodeToString(data); strings.Contains(stored, encoded) {
		t.Fatal("key material exposed")
	}
}
