package whatsapp

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func sign(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestValidSignature(t *testing.T) {
	secret := "app-secret"
	body := []byte(`{"entry":[]}`)

	if !validSignature(secret, sign(secret, body), body) {
		t.Error("correct signature rejected")
	}
	if validSignature(secret, sign("wrong-secret", body), body) {
		t.Error("wrong signature accepted")
	}
	if validSignature(secret, "sha256=not-hex", body) {
		t.Error("malformed signature accepted")
	}
	if validSignature(secret, "", body) {
		t.Error("missing signature accepted")
	}
	// Empty configured secret disables verification (local dev).
	if !validSignature("", "", body) {
		t.Error("empty secret should skip verification")
	}
}
