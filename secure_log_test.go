package ssp

import (
	"testing"
)

func TestTruncateKey(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		maxLen   int
		expected string
	}{
		{"empty string", "", 8, "(empty)"},
		{"short key", "abc", 8, "abc"},
		{"exact length", "12345678", 8, "12345678"},
		{"long key", "1234567890abcdef", 8, "12345678"},
		{"very long key", "abcdefghijklmnopqrstuvwxyz", 10, "abcdefghij"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateKey(tt.key, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncateKey(%q, %d) = %q, want %q", tt.key, tt.maxLen, result, tt.expected)
			}
		})
	}
}

func TestMaskIP(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expected string
	}{
		{"empty IP", "", "(no-ip)"},
		{"short IP", "127.0.0.1", "127.0.0.1"},
		{"long IP", "192.168.100.200", "192.168..."},
		{"IPv6", "2001:0db8:85a3:0000", "2001:0db8..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskIP(tt.ip)
			if result != tt.expected {
				t.Errorf("maskIP(%q) = %q, want %q", tt.ip, result, tt.expected)
			}
		})
	}
}

func TestSafeLogRequest(t *testing.T) {
	// Test with nil request
	SafeLogRequest(nil)

	// Test with nil client
	req := &CliRequest{
		IPAddress: "192.168.1.1",
	}
	SafeLogRequest(req)

	// Test with valid request
	req = &CliRequest{
		Client: &ClientBody{
			Cmd: "ident",
			Idk: "abcdefghijklmnop",
		},
		IPAddress: "192.168.1.1",
	}
	SafeLogRequest(req)
}

func TestSafeLogIdentity(t *testing.T) {
	// Test with nil identity
	SafeLogIdentity(nil)

	// Test with valid identity
	identity := &SqrlIdentity{
		Idk:      "abcdefghijklmnop",
		Disabled: true,
		Rekeyed:  "newkey",
	}
	SafeLogIdentity(identity)

	// Test with identity without rekeyed
	identity = &SqrlIdentity{
		Idk:      "abcdefghijklmnop",
		Disabled: false,
	}
	SafeLogIdentity(identity)
}

func TestSafeLogResponse(t *testing.T) {
	// Test with nil response
	SafeLogResponse(nil)

	// Test with valid response
	resp := &CliResponse{
		Nut: "testnut12345",
		TIF: TIFIDMatch | TIFIPMatched,
	}
	SafeLogResponse(resp)
}

func TestSafeLogError(t *testing.T) {
	// Test with nil error
	SafeLogError("test_context", nil)

	// Test with valid error
	SafeLogError("test_context", ErrNotFound)
}

func TestSafeLogAuth(t *testing.T) {
	SafeLogAuth("login", "abcdefghijklmnop", true)
	SafeLogAuth("login", "abcdefghijklmnop", false)
	SafeLogAuth("logout", "short", true)
}

func BenchmarkTruncateKey(b *testing.B) {
	key := "abcdefghijklmnopqrstuvwxyz0123456789"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		truncateKey(key, 8)
	}
}

func BenchmarkMaskIP(b *testing.B) {
	ip := "192.168.100.200"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		maskIP(ip)
	}
}

func BenchmarkSafeLogRequest(b *testing.B) {
	req := &CliRequest{
		Client: &ClientBody{
			Cmd: "ident",
			Idk: "abcdefghijklmnop",
		},
		IPAddress: "192.168.1.1",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SafeLogRequest(req)
	}
}
