package ssp

import "testing"

func TestGrcStaticGenerate(t *testing.T) {
	// Use a valid 16-byte AES key (AES-128)
	aesKey := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	tree, err := NewGrcTree(10, aesKey)
	if err != nil {
		t.Fatalf("Error creating tree: %v", err)
	}
	defer tree.Close()

	nut, err := tree.Nut()
	if err != nil {
		t.Fatalf("Error creating nut: %v", err)
	}

	// AES block size is 16 bytes, base64url encoded = 22 characters
	if len(nut) != 22 {
		t.Fatalf("Nut length is: %d expected: 22", len(nut))
	}

	// Expected value for counter=11 with the given AES key
	expectedNut := "oHEbbCFEu0nMdJORt8kAyw"
	if nut != Nut(expectedNut) {
		t.Fatalf("Expected nut value %v but got %v", expectedNut, nut)
	}
}

func TestGrcUniqueGenerate(t *testing.T) {
	// Use a valid 16-byte AES key (AES-128)
	aesKey := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	tree, err := NewGrcTree(9, aesKey)
	if err != nil {
		t.Fatalf("Error creating tree: %v", err)
	}
	defer tree.Close()

	values := make(map[Nut]struct{}, 10)

	for i := 0; i < 10; i++ {
		nut, err := tree.Nut()
		if err != nil {
			t.Fatalf("Error creating nut: %v", err)
		}
		if _, ok := values[nut]; ok {
			t.Fatalf("Found duplicate %v", nut)
		}
		values[nut] = struct{}{}
	}
}

func TestGrcInvalidKeySize(t *testing.T) {
	// Test that invalid key sizes are rejected
	invalidKeys := [][]byte{
		{1, 2, 3, 4},          // Too short
		{1, 2, 3, 4, 5},       // Invalid size (5 bytes)
		make([]byte, 15),      // Too short for AES-128
		make([]byte, 17),      // Invalid size between AES-128 and AES-192
		make([]byte, 33),      // Too long
	}

	for i, key := range invalidKeys {
		_, err := NewGrcTree(0, key)
		if err == nil {
			t.Errorf("Test case %d: Expected error for key size %d, but got none", i, len(key))
		}
	}
}

func TestGrcValidKeySizes(t *testing.T) {
	// Test all valid AES key sizes
	validKeys := [][]byte{
		make([]byte, 16), // AES-128
		make([]byte, 24), // AES-192
		make([]byte, 32), // AES-256
	}

	for i, key := range validKeys {
		tree, err := NewGrcTree(0, key)
		if err != nil {
			t.Errorf("Test case %d: Unexpected error for key size %d: %v", i, len(key), err)
		}
		if tree != nil {
			tree.Close()
		}
	}
}

func TestGrcClose(t *testing.T) {
	aesKey := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	tree, err := NewGrcTree(0, aesKey)
	if err != nil {
		t.Fatalf("Error creating tree: %v", err)
	}

	tree.Close()

	// After close, key should be nil
	if tree.key != nil {
		t.Error("Key should be nil after Close()")
	}
	if tree.cipher != nil {
		t.Error("Cipher should be nil after Close()")
	}
}
