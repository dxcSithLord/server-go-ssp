package ssp

import (
	"testing"
)

func TestClearBytes(t *testing.T) {
	// Test clearing a byte slice
	data := []byte("sensitive-key-data-12345")
	original := make([]byte, len(data))
	copy(original, data)

	ClearBytes(data)

	// Verify all bytes are zero
	for i, b := range data {
		if b != 0 {
			t.Errorf("ClearBytes failed: byte at index %d is %d, expected 0", i, b)
		}
	}

	// Verify original was not empty
	hasNonZero := false
	for _, b := range original {
		if b != 0 {
			hasNonZero = true
			break
		}
	}
	if !hasNonZero {
		t.Error("Test setup error: original data was all zeros")
	}
}

func TestClearBytesEmpty(t *testing.T) {
	// Should not panic on empty slice
	var empty []byte
	ClearBytes(empty)
}

func TestClearBytesNil(t *testing.T) {
	// Should not panic on nil slice
	var nilSlice []byte = nil
	ClearBytes(nilSlice)
}

func TestClearString(t *testing.T) {
	sensitive := "my-secret-password-123"
	ClearString(&sensitive)

	if sensitive != "" {
		t.Errorf("ClearString failed: string is '%s', expected empty", sensitive)
	}
}

func TestClearStringEmpty(t *testing.T) {
	empty := ""
	ClearString(&empty)

	if empty != "" {
		t.Errorf("ClearString failed: string is '%s', expected empty", empty)
	}
}

func TestClearStringNil(t *testing.T) {
	// Should not panic on nil pointer
	ClearString(nil)
}

func TestSqrlIdentityClear(t *testing.T) {
	identity := &SqrlIdentity{
		Idk:      "test-idk-key-12345",
		Suk:      "test-suk-key-67890",
		Vuk:      "test-vuk-key-abcde",
		Pidk:     "test-pidk-key-fghij",
		Rekeyed:  "test-rekeyed-value",
		SQRLOnly: true,
		Hardlock: true,
		Disabled: true,
		Btn:      2,
	}

	identity.Clear()

	if identity.Idk != "" {
		t.Errorf("Clear failed: Idk is '%s', expected empty", identity.Idk)
	}
	if identity.Suk != "" {
		t.Errorf("Clear failed: Suk is '%s', expected empty", identity.Suk)
	}
	if identity.Vuk != "" {
		t.Errorf("Clear failed: Vuk is '%s', expected empty", identity.Vuk)
	}
	if identity.Pidk != "" {
		t.Errorf("Clear failed: Pidk is '%s', expected empty", identity.Pidk)
	}
	if identity.Rekeyed != "" {
		t.Errorf("Clear failed: Rekeyed is '%s', expected empty", identity.Rekeyed)
	}
	if identity.SQRLOnly {
		t.Error("Clear failed: SQRLOnly is true, expected false")
	}
	if identity.Hardlock {
		t.Error("Clear failed: Hardlock is true, expected false")
	}
	if identity.Disabled {
		t.Error("Clear failed: Disabled is true, expected false")
	}
	if identity.Btn != 0 {
		t.Errorf("Clear failed: Btn is %d, expected 0", identity.Btn)
	}
}

func TestSqrlIdentityClearNil(t *testing.T) {
	// Should not panic on nil
	var identity *SqrlIdentity
	identity.Clear()
}

func TestClientBodyClear(t *testing.T) {
	cb := &ClientBody{
		Version: []int{1, 2, 3},
		Cmd:     "ident",
		Opt:     map[string]bool{"sqrlonly": true, "hardlock": true},
		Suk:     "test-suk",
		Vuk:     "test-vuk",
		Pidk:    "test-pidk",
		Idk:     "test-idk",
		Btn:     1,
	}

	cb.Clear()

	if cb.Suk != "" {
		t.Errorf("Clear failed: Suk is '%s', expected empty", cb.Suk)
	}
	if cb.Vuk != "" {
		t.Errorf("Clear failed: Vuk is '%s', expected empty", cb.Vuk)
	}
	if cb.Pidk != "" {
		t.Errorf("Clear failed: Pidk is '%s', expected empty", cb.Pidk)
	}
	if cb.Idk != "" {
		t.Errorf("Clear failed: Idk is '%s', expected empty", cb.Idk)
	}
	if cb.Version != nil {
		t.Errorf("Clear failed: Version is %v, expected nil", cb.Version)
	}
	if cb.Cmd != "" {
		t.Errorf("Clear failed: Cmd is '%s', expected empty", cb.Cmd)
	}
	if cb.Opt != nil {
		t.Errorf("Clear failed: Opt is %v, expected nil", cb.Opt)
	}
	if cb.Btn != 0 {
		t.Errorf("Clear failed: Btn is %d, expected 0", cb.Btn)
	}
}

func TestClientBodyClearNil(t *testing.T) {
	var cb *ClientBody
	cb.Clear()
}

func TestCliRequestClear(t *testing.T) {
	cr := &CliRequest{
		Client: &ClientBody{
			Idk: "test-idk",
			Suk: "test-suk",
		},
		ClientEncoded: "encoded-data",
		Server:        "server-data",
		Ids:           "signature-ids",
		Pids:          "signature-pids",
		Urs:           "signature-urs",
		IPAddress:     "192.168.1.1",
	}

	cr.Clear()

	if cr.Ids != "" {
		t.Errorf("Clear failed: Ids is '%s', expected empty", cr.Ids)
	}
	if cr.Pids != "" {
		t.Errorf("Clear failed: Pids is '%s', expected empty", cr.Pids)
	}
	if cr.Urs != "" {
		t.Errorf("Clear failed: Urs is '%s', expected empty", cr.Urs)
	}
	if cr.ClientEncoded != "" {
		t.Errorf("Clear failed: ClientEncoded is '%s', expected empty", cr.ClientEncoded)
	}
	if cr.Server != "" {
		t.Errorf("Clear failed: Server is '%s', expected empty", cr.Server)
	}
	if cr.IPAddress != "" {
		t.Errorf("Clear failed: IPAddress is '%s', expected empty", cr.IPAddress)
	}
	if cr.Client.Idk != "" {
		t.Errorf("Clear failed: Client.Idk is '%s', expected empty", cr.Client.Idk)
	}
}

func TestCliRequestClearNil(t *testing.T) {
	var cr *CliRequest
	cr.Clear()
}

func TestHoardCacheClear(t *testing.T) {
	hc := &HoardCache{
		State:       "authenticated",
		RemoteIP:    "192.168.1.1",
		OriginalNut: "original-nut-value",
		PagNut:      "pag-nut-value",
		LastRequest: &CliRequest{
			Ids: "test-ids",
		},
		Identity: &SqrlIdentity{
			Idk: "test-idk",
		},
		LastResponse: []byte("response-data"),
	}

	hc.Clear()

	if hc.State != "" {
		t.Errorf("Clear failed: State is '%s', expected empty", hc.State)
	}
	if hc.RemoteIP != "" {
		t.Errorf("Clear failed: RemoteIP is '%s', expected empty", hc.RemoteIP)
	}
	if hc.OriginalNut != "" {
		t.Errorf("Clear failed: OriginalNut is '%s', expected empty", hc.OriginalNut)
	}
	if hc.PagNut != "" {
		t.Errorf("Clear failed: PagNut is '%s', expected empty", hc.PagNut)
	}
	if hc.LastRequest.Ids != "" {
		t.Errorf("Clear failed: LastRequest.Ids is '%s', expected empty", hc.LastRequest.Ids)
	}
	if hc.Identity.Idk != "" {
		t.Errorf("Clear failed: Identity.Idk is '%s', expected empty", hc.Identity.Idk)
	}
	for i, b := range hc.LastResponse {
		if b != 0 {
			t.Errorf("Clear failed: LastResponse[%d] is %d, expected 0", i, b)
		}
	}
}

func TestHoardCacheClearNil(t *testing.T) {
	var hc *HoardCache
	hc.Clear()
}

func TestClearBytesSecure(t *testing.T) {
	data := []byte("highly-sensitive-data-that-must-be-cleared")
	ClearBytesSecure(data)

	for i, b := range data {
		if b != 0 {
			t.Errorf("ClearBytesSecure failed: byte at index %d is %d, expected 0", i, b)
		}
	}
}

func BenchmarkClearBytes(b *testing.B) {
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ClearBytes(data)
	}
}

func BenchmarkClearBytesSecure(b *testing.B) {
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ClearBytesSecure(data)
	}
}

func BenchmarkSqrlIdentityClear(b *testing.B) {
	for i := 0; i < b.N; i++ {
		identity := &SqrlIdentity{
			Idk:      "test-idk-key-12345",
			Suk:      "test-suk-key-67890",
			Vuk:      "test-vuk-key-abcde",
			Pidk:     "test-pidk-key-fghij",
			Rekeyed:  "test-rekeyed-value",
			SQRLOnly: true,
			Hardlock: true,
			Disabled: true,
			Btn:      2,
		}
		identity.Clear()
	}
}
