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
		{"key with newline", "abc\ndef", 8, "abc def"},
		{"key with carriage return", "abc\rdef", 8, "abc def"},
		{"key with tab", "abc\tdef", 8, "abc def"},
		{"key with null", "abc\x00def", 8, "abc def"},
		{"key with escape", "abc\x1bdef", 8, "abc def"},
		{"key with control chars long", "abc\ndef\rghi\tjkl", 8, "abc def "},
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

func TestSanitizeControlChars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"no control chars", "abcdef", "abcdef"},
		{"newline", "abc\ndef", "abc def"},
		{"carriage return", "abc\rdef", "abc def"},
		{"tab", "abc\tdef", "abc def"},
		{"null byte", "abc\x00def", "abc def"},
		{"escape char", "abc\x1bdef", "abc def"},
		{"multiple controls", "a\nb\rc\td", "a b c d"},
		{"control at start", "\nabc", " abc"},
		{"control at end", "abc\n", "abc "},
		{"only controls", "\n\r\t", "   "},
		{"DEL character", "abc\x7fdef", "abc def"},
		{"bell character", "abc\x07def", "abc def"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeControlChars(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeControlChars(%q) = %q, want %q", tt.input, result, tt.expected)
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
		{"short IP", "127.0.0.1", "127.0.*.*"},
		{"long IP", "192.168.100.200", "192.168.*.*"},
		{"IPv6", "2001:0db8:85a3:0000", "2001:***"},
		{"IPv4 with port", "192.168.1.1:8080", "192.168.1.1:***"},
		{"short string", "abc", "abc"},
		{"IP with newline", "192.168.1.1\n", "192.168.*.*"},
		{"IP with carriage return", "192.168.1.1\r", "192.168.*.*"},
		{"IP with tab", "192.168.1.1\t", "192.168.*.*"},
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
