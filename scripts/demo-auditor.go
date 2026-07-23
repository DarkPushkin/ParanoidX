// Package main provides scripts functionality for the simplex-node server
package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func main() {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Pubkey:  %s\n", hex.EncodeToString(pub))
	fmt.Printf("Privkey: %s\n", hex.EncodeToString(priv))
}
