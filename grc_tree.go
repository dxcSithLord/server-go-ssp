package ssp

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"fmt"
	"sync/atomic"
)

// GrcTree Creates a 64-bit nut based on the GRC spec using a monotonic counter
// and AES cipher (upgraded from deprecated blowfish)
type GrcTree struct {
	monotonicCounter uint64
	cipher           cipher.Block
	key              []byte
}

// NewGrcTree takes an initial counter value (in the case of reboot) and
// an AES key (16, 24, or 32 bytes for AES-128, AES-192, or AES-256)
// This replaces the deprecated blowfish cipher with AES as recommended.
func NewGrcTree(counterInit uint64, aesKey []byte) (*GrcTree, error) {
	// Validate key length for AES
	keyLen := len(aesKey)
	if keyLen != 16 && keyLen != 24 && keyLen != 32 {
		return nil, fmt.Errorf("invalid AES key size: must be 16, 24, or 32 bytes, got %d", keyLen)
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("couldn't initialize AES cipher: %v", err)
	}
	return &GrcTree{
		monotonicCounter: counterInit,
		cipher:           block,
		key:              aesKey,
	}, nil
}

// Nut Create a nut based on the GRC spec.
// Uses AES encryption on a monotonic counter to generate unique, unpredictable tokens.
func (gt *GrcTree) Nut() (Nut, error) {
	nextValue := atomic.AddUint64(&gt.monotonicCounter, 1)

	// Create 16-byte block for AES (pad counter with zeros)
	plaintext := make([]byte, aes.BlockSize)
	binary.LittleEndian.PutUint64(plaintext[:8], nextValue)
	defer ClearBytes(plaintext) // Securely clear plaintext

	encrypted := make([]byte, aes.BlockSize)
	gt.cipher.Encrypt(encrypted, plaintext)
	defer ClearBytes(encrypted) // Securely clear encrypted bytes after encoding

	return Nut(Sqrl64.EncodeToString(encrypted)), nil
}

// Close securely clears the AES key material.
// Should be called when the GrcTree is no longer needed.
func (gt *GrcTree) Close() {
	if gt.key != nil {
		ClearBytes(gt.key)
		gt.key = nil
	}
	gt.cipher = nil
}
