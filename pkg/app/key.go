package app

import (
	"crypto/ed25519"
	_ "embed"
	"encoding/hex"
	"os"
)

//go:embed key.pub
var publicKey []byte

func PublicKey() []byte {
	return publicKey
}

func GenerateKey() error {
	pub, pri, err := ed25519.GenerateKey(nil)
	if err != nil {
		return err
	}
	_ = os.WriteFile("key.pub", pub, 0600)
	_ = os.WriteFile("key.txt", []byte(hex.EncodeToString(pri)), 0600)
	return nil
}
