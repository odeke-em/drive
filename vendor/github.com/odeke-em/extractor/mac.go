package extractor

import (
	"crypto/hmac"
	"crypto/sha256"
	"os"
)

type KeySet struct {
	PublicKey  string
	PrivateKey string
}

type EnvKey struct {
	PubKeyAlias  string
	PrivKeyAlias string
}

var hashAlgo = sha256.New

func KeySetFromEnv(e *EnvKey) *KeySet {
	pubKey := os.Getenv(e.PubKeyAlias)
	privKey := os.Getenv(e.PrivKeyAlias)

	return &KeySet{
		PublicKey:  pubKey,
		PrivateKey: privKey,
	}
}

func (ks *KeySet) Sign(message []byte) []byte {
	mac := hmac.New(hashAlgo, []byte(ks.PrivateKey+ks.PublicKey))
	mac.Write(message)
	return mac.Sum(nil)
}

func (ks *KeySet) Match(message, messageMAC []byte) bool {
	expectedMAC := ks.Sign(message)
	return hmac.Equal(expectedMAC, messageMAC)
}
