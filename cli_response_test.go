package ssp

import (
	"strings"
	"testing"
)

// TIF flag setter/getter tests
func TestCliResponse_TIFFlags(t *testing.T) {
	tests := []struct {
		name     string
		apply    func(*CliResponse) *CliResponse
		expected uint32
	}{
		{"WithIDMatch", func(cr *CliResponse) *CliResponse { return cr.WithIDMatch() }, TIFIDMatch},
		{"WithPreviousIDMatch", func(cr *CliResponse) *CliResponse { return cr.WithPreviousIDMatch() }, TIFPreviousIDMatch},
		{"WithIPMatch", func(cr *CliResponse) *CliResponse { return cr.WithIPMatch() }, TIFIPMatched},
		{"WithSQRLDisabled", func(cr *CliResponse) *CliResponse { return cr.WithSQRLDisabled() }, TIFSQRLDisabled},
		{"WithFunctionNotSupported", func(cr *CliResponse) *CliResponse { return cr.WithFunctionNotSupported() }, TIFFunctionNotSupported},
		{"WithTransientError", func(cr *CliResponse) *CliResponse { return cr.WithTransientError() }, TIFTransientError},
		{"WithClientFailure", func(cr *CliResponse) *CliResponse { return cr.WithClientFailure() }, TIFClientFailure},
		{"WithCommandFailed", func(cr *CliResponse) *CliResponse { return cr.WithCommandFailed() }, TIFCommandFailed},
		{"WithBadIDAssociation", func(cr *CliResponse) *CliResponse { return cr.WithBadIDAssociation() }, TIFBadIDAssociation},
		{"WithIdentitySuperseded", func(cr *CliResponse) *CliResponse { return cr.WithIdentitySuperseded() }, TIFIdentitySuperseded},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := NewCliResponse("testnut", "/test")
			result := tt.apply(cr)
			if result != cr {
				t.Error("Expected same object returned for chaining")
			}
			if cr.TIF&tt.expected == 0 {
				t.Errorf("Expected TIF flag %x to be set", tt.expected)
			}
		})
	}
}

func TestCliResponse_ClearIDMatch(t *testing.T) {
	cr := NewCliResponse("testnut", "/test")
	cr.WithIDMatch()
	if cr.TIF&TIFIDMatch == 0 {
		t.Fatal("IDMatch should be set")
	}
	cr.ClearIDMatch()
	if cr.TIF&TIFIDMatch != 0 {
		t.Error("IDMatch should be cleared")
	}
}

func TestCliResponse_ClearPreviousIDMatch(t *testing.T) {
	cr := NewCliResponse("testnut", "/test")
	cr.WithPreviousIDMatch()
	if cr.TIF&TIFPreviousIDMatch == 0 {
		t.Fatal("PreviousIDMatch should be set")
	}
	cr.ClearPreviousIDMatch()
	if cr.TIF&TIFPreviousIDMatch != 0 {
		t.Error("PreviousIDMatch should be cleared")
	}
}

func TestCliResponse_MultipleTIFFlags(t *testing.T) {
	cr := NewCliResponse("testnut", "/test")
	cr.WithIDMatch().WithIPMatch().WithClientFailure()

	if cr.TIF&TIFIDMatch == 0 {
		t.Error("Expected TIFIDMatch")
	}
	if cr.TIF&TIFIPMatched == 0 {
		t.Error("Expected TIFIPMatched")
	}
	if cr.TIF&TIFClientFailure == 0 {
		t.Error("Expected TIFClientFailure")
	}
	// Should not have flags we didn't set
	if cr.TIF&TIFSQRLDisabled != 0 {
		t.Error("Did not expect TIFSQRLDisabled")
	}
}

func TestCliResponse_ClearDoesNotAffectOthers(t *testing.T) {
	cr := NewCliResponse("testnut", "/test")
	cr.WithIDMatch().WithPreviousIDMatch().WithIPMatch()

	cr.ClearIDMatch()

	if cr.TIF&TIFIDMatch != 0 {
		t.Error("IDMatch should be cleared")
	}
	if cr.TIF&TIFPreviousIDMatch == 0 {
		t.Error("PreviousIDMatch should still be set")
	}
	if cr.TIF&TIFIPMatched == 0 {
		t.Error("IPMatched should still be set")
	}
}

func TestCliResponse_EncodeAndParse(t *testing.T) {
	cr := NewCliResponse("testnut123", "/sqrl/cli.sqrl?nut=testnut123")
	cr.WithIDMatch().WithIPMatch()
	cr.URL = "https://example.com/redirect"
	cr.Sin = "sin-value"
	cr.Suk = "suk-value"
	cr.Can = "can-value"

	encoded := cr.Encode()
	if len(encoded) == 0 {
		t.Fatal("Expected non-empty encoded response")
	}

	parsed, err := ParseCliResponse(encoded)
	if err != nil {
		t.Fatalf("ParseCliResponse failed: %v", err)
	}

	if parsed.Nut != "testnut123" {
		t.Errorf("Expected nut testnut123, got %s", parsed.Nut)
	}
	if parsed.Qry != "/sqrl/cli.sqrl?nut=testnut123" {
		t.Errorf("Expected qry, got %s", parsed.Qry)
	}
	if parsed.TIF&TIFIDMatch == 0 {
		t.Error("Expected TIFIDMatch in parsed response")
	}
	if parsed.TIF&TIFIPMatched == 0 {
		t.Error("Expected TIFIPMatched in parsed response")
	}
	if parsed.URL != "https://example.com/redirect" {
		t.Errorf("Expected URL, got %s", parsed.URL)
	}
	if parsed.Sin != "sin-value" {
		t.Errorf("Expected sin, got %s", parsed.Sin)
	}
	if parsed.Suk != "suk-value" {
		t.Errorf("Expected suk, got %s", parsed.Suk)
	}
	if parsed.Can != "can-value" {
		t.Errorf("Expected can, got %s", parsed.Can)
	}
}

func TestCliResponse_EncodeWithAsk(t *testing.T) {
	cr := NewCliResponse("testnut", "/test")
	cr.Ask = &Ask{
		Message: "Do you want to proceed?",
		Button1: "Yes",
		URL1:    "https://example.com/yes",
		Button2: "No",
	}

	encoded := cr.Encode()
	if len(encoded) == 0 {
		t.Fatal("Expected non-empty encoded response")
	}

	parsed, err := ParseCliResponse(encoded)
	if err != nil {
		t.Fatalf("ParseCliResponse failed: %v", err)
	}

	if parsed.Ask == nil {
		t.Fatal("Expected Ask to be parsed")
	}
	if parsed.Ask.Message != "Do you want to proceed?" {
		t.Errorf("Expected Ask message, got %s", parsed.Ask.Message)
	}
	if parsed.Ask.Button1 != "Yes" {
		t.Errorf("Expected Button1 Yes, got %s", parsed.Ask.Button1)
	}
	if parsed.Ask.URL1 != "https://example.com/yes" {
		t.Errorf("Expected URL1, got %s", parsed.Ask.URL1)
	}
	if parsed.Ask.Button2 != "No" {
		t.Errorf("Expected Button2 No, got %s", parsed.Ask.Button2)
	}
}

func TestParseCliResponse_InvalidBase64(t *testing.T) {
	_, err := ParseCliResponse([]byte("!!!invalid!!!"))
	if err == nil {
		t.Error("Expected error for invalid base64")
	}
}

func TestParseCliResponse_InvalidTIF(t *testing.T) {
	// Create a response with invalid TIF
	raw := "ver=1\r\nnut=test\r\ntif=zzz\r\nqry=/test\r\n"
	encoded := Sqrl64.EncodeToString([]byte(raw))
	_, err := ParseCliResponse([]byte(encoded))
	if err == nil {
		t.Error("Expected error for invalid TIF")
	}
}

func TestAsk_Encode(t *testing.T) {
	ask := &Ask{
		Message: "Test message",
		Button1: "OK",
		URL1:    "https://example.com",
		Button2: "Cancel",
	}

	encoded := ask.Encode()
	if encoded == "" {
		t.Fatal("Expected non-empty encoded ask")
	}
	parts := strings.Split(encoded, "~")
	if len(parts) != 3 {
		t.Errorf("Expected 3 parts, got %d", len(parts))
	}
}

func TestAsk_EncodeOneButton(t *testing.T) {
	ask := &Ask{
		Message: "Confirm?",
		Button1: "Yes",
	}

	encoded := ask.Encode()
	parts := strings.Split(encoded, "~")
	if len(parts) != 2 {
		t.Errorf("Expected 2 parts for one button, got %d", len(parts))
	}
}

func TestAsk_EncodeNoButtons(t *testing.T) {
	ask := &Ask{
		Message: "Information only",
	}

	encoded := ask.Encode()
	parts := strings.Split(encoded, "~")
	if len(parts) != 1 {
		t.Errorf("Expected 1 part for no buttons, got %d", len(parts))
	}
}

func TestParseAsk_Full(t *testing.T) {
	ask := &Ask{
		Message: "Test?",
		Button1: "OK",
		URL1:    "https://example.com/ok",
		Button2: "Cancel",
		URL2:    "https://example.com/cancel",
	}

	encoded := ask.Encode()
	parsed := ParseAsk(encoded)

	if parsed.Message != "Test?" {
		t.Errorf("Expected message 'Test?', got '%s'", parsed.Message)
	}
	if parsed.Button1 != "OK" {
		t.Errorf("Expected Button1 'OK', got '%s'", parsed.Button1)
	}
	if parsed.URL1 != "https://example.com/ok" {
		t.Errorf("Expected URL1, got '%s'", parsed.URL1)
	}
	if parsed.Button2 != "Cancel" {
		t.Errorf("Expected Button2 'Cancel', got '%s'", parsed.Button2)
	}
}

func TestEncodeButton(t *testing.T) {
	result := encodeButton("Accept", "https://example.com")
	if result == "" {
		t.Error("Expected non-empty encoded button")
	}
	// Decode and verify
	decoded, err := Sqrl64.DecodeString(result)
	if err != nil {
		t.Fatalf("Failed to decode button: %v", err)
	}
	if string(decoded) != "Accept;https://example.com" {
		t.Errorf("Expected 'Accept;https://example.com', got '%s'", string(decoded))
	}
}

func TestEncodeButton_NoURL(t *testing.T) {
	result := encodeButton("OK", "")
	if result == "" {
		t.Error("Expected non-empty encoded button")
	}
	decoded, _ := Sqrl64.DecodeString(result)
	if string(decoded) != "OK" {
		t.Errorf("Expected 'OK', got '%s'", string(decoded))
	}
}

func TestEncodeButton_Empty(t *testing.T) {
	result := encodeButton("", "")
	if result != "" {
		t.Errorf("Expected empty string for empty button, got '%s'", result)
	}
}

func TestRemoveSemi(t *testing.T) {
	tests := []struct {
		input, expected string
	}{
		{"no semicolons", "no semicolons"},
		{"semi;colon", "semicolon"},
		{"a;b;c", "abc"},
		{"", ""},
	}
	for _, tt := range tests {
		result := removeSemi(tt.input)
		if result != tt.expected {
			t.Errorf("removeSemi(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestTIFDesc(t *testing.T) {
	// Verify all TIF flags have descriptions
	flags := []uint32{
		TIFIDMatch, TIFPreviousIDMatch, TIFIPMatched, TIFSQRLDisabled,
		TIFFunctionNotSupported, TIFTransientError, TIFCommandFailed,
		TIFClientFailure, TIFBadIDAssociation, TIFIdentitySuperseded,
	}
	for _, flag := range flags {
		if _, ok := TIFDesc[flag]; !ok {
			t.Errorf("Missing TIFDesc for flag %x", flag)
		}
	}
}

func BenchmarkCliResponseEncode(b *testing.B) {
	cr := NewCliResponse("testnut", "/test")
	cr.WithIDMatch().WithIPMatch()
	cr.URL = "https://example.com"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cr.Encode()
	}
}

func BenchmarkParseCliResponse(b *testing.B) {
	cr := NewCliResponse("testnut", "/test")
	cr.WithIDMatch().WithIPMatch()
	encoded := cr.Encode()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseCliResponse(encoded)
	}
}

func TestAskButtonParse(t *testing.T) {
	b, u := splitButton("button;https://example.com")
	if b != "button" {
		t.Errorf("didn't get button value: %v", b)
	}
	if u != "https://example.com" {
		t.Errorf("Failed url: %v", u)
	}
}

func TestAskButtonNoURL(t *testing.T) {
	b, u := splitButton("button")
	if b != "button" {
		t.Errorf("didn't get button value: %v", b)
	}
	if u != "" {
		t.Errorf("Failed url: %v", u)
	}
}

func TestAskButtonNoURLWithSemi(t *testing.T) {
	b, u := splitButton("button;")
	if b != "button" {
		t.Errorf("didn't get button value: %v", b)
	}
	if u != "" {
		t.Errorf("Failed url: %v", u)
	}
}

func TestAskNoButtonWithURL(t *testing.T) {
	b, u := splitButton(";https://example.com")
	if b != "" {
		t.Errorf("didn't get button value: %v", b)
	}
	if u != "https://example.com" {
		t.Errorf("Failed url: %v", u)
	}
}
