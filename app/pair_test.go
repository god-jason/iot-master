package app

import (
	"crypto/ed25519"
	_ "embed"
	"encoding/hex"
	"os"
	"testing"
)

func TestGenerateKeyPair(t *testing.T) {
	pub, pri, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("pub:", hex.EncodeToString(pub))
	t.Log("pri:", hex.EncodeToString(pri))

	_ = os.WriteFile("public.key", []byte(hex.EncodeToString(pub)), 0600)
	_ = os.WriteFile("private.key", []byte(hex.EncodeToString(pri)), 0600)
}
