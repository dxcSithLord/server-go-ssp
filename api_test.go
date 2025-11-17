package ssp

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewSqrlSspAPI_WithNilTree(t *testing.T) {
	hoard := NewMapHoard()
	authStore := NewMapAuthStore()
	auth := &mockAuthenticator{}

	api := NewSqrlSspAPI(nil, hoard, auth, authStore)

	if api == nil {
		t.Fatal("NewSqrlSspAPI returned nil")
	}
	if api.tree == nil {
		t.Error("Expected tree to be initialized")
	}
	if api.NutExpiration != 10*time.Minute {
		t.Errorf("Expected NutExpiration to be 10m, got %v", api.NutExpiration)
	}
}

func TestNewSqrlSspAPI_WithTree(t *testing.T) {
	tree, err := NewRandomTree(8)
	if err != nil {
		t.Fatalf("Failed to create RandomTree: %v", err)
	}
	hoard := NewMapHoard()
	authStore := NewMapAuthStore()
	auth := &mockAuthenticator{}

	api := NewSqrlSspAPI(tree, hoard, auth, authStore)

	if api == nil {
		t.Fatal("NewSqrlSspAPI returned nil")
	}
	if api.tree != tree {
		t.Error("Expected tree to be the provided tree")
	}
}

func TestSqrlSspAPI_NutExpirationSeconds(t *testing.T) {
	api := &SqrlSspAPI{
		NutExpiration: 5 * time.Minute,
	}

	seconds := api.NutExpirationSeconds()
	if seconds != 300 {
		t.Errorf("Expected 300 seconds, got %d", seconds)
	}
}

func TestSqrlSspAPI_Host_Override(t *testing.T) {
	api := &SqrlSspAPI{
		HostOverride: "override.example.com",
	}

	req := httptest.NewRequest("GET", "/", nil)
	req.Host = "original.example.com"
	req.Header.Set("X-Forwarded-Host", "forwarded.example.com")

	host := api.Host(req)
	if host != "override.example.com" {
		t.Errorf("Expected override.example.com, got %s", host)
	}
}

func TestSqrlSspAPI_Host_ForwardedHost(t *testing.T) {
	api := &SqrlSspAPI{}

	req := httptest.NewRequest("GET", "/", nil)
	req.Host = "original.example.com"
	req.Header.Set("X-Forwarded-Host", "forwarded.example.com")

	host := api.Host(req)
	if host != "forwarded.example.com" {
		t.Errorf("Expected forwarded.example.com, got %s", host)
	}
}

func TestSqrlSspAPI_Host_ForwardedServer(t *testing.T) {
	api := &SqrlSspAPI{}

	req := httptest.NewRequest("GET", "/", nil)
	req.Host = "original.example.com"
	req.Header.Set("X-Forwarded-Server", "server.example.com")

	host := api.Host(req)
	if host != "server.example.com" {
		t.Errorf("Expected server.example.com, got %s", host)
	}
}

func TestSqrlSspAPI_Host_RequestHost(t *testing.T) {
	api := &SqrlSspAPI{}

	req := httptest.NewRequest("GET", "/", nil)
	req.Host = "original.example.com"

	host := api.Host(req)
	if host != "original.example.com" {
		t.Errorf("Expected original.example.com, got %s", host)
	}
}

func TestSqrlSspAPI_RemoteIP_Forwarded(t *testing.T) {
	api := &SqrlSspAPI{}

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	req.Header.Set("X-Forwarded-For", "10.0.0.1")

	ip := api.RemoteIP(req)
	if ip != "10.0.0.1" {
		t.Errorf("Expected 10.0.0.1, got %s", ip)
	}
}

func TestSqrlSspAPI_RemoteIP_Direct(t *testing.T) {
	api := &SqrlSspAPI{}

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	ip := api.RemoteIP(req)
	if ip != "192.168.1.1:12345" {
		t.Errorf("Expected 192.168.1.1:12345, got %s", ip)
	}
}

func TestSqrlSspAPI_HTTPSRoot(t *testing.T) {
	api := &SqrlSspAPI{
		HostOverride: "example.com",
		RootPath:     "/sqrl",
	}

	req := httptest.NewRequest("GET", "/", nil)
	url := api.HTTPSRoot(req)

	if url.Scheme != "https" {
		t.Errorf("Expected scheme https, got %s", url.Scheme)
	}
	if url.Host != "example.com" {
		t.Errorf("Expected host example.com, got %s", url.Host)
	}
	if url.Path != "/sqrl" {
		t.Errorf("Expected path /sqrl, got %s", url.Path)
	}
}

func TestSqrlSspAPI_qry(t *testing.T) {
	api := &SqrlSspAPI{
		RootPath: "/auth",
	}

	qry := api.qry("testnut123")
	expected := "/auth/cli.sqrl?nut=testnut123"
	if qry != expected {
		t.Errorf("Expected %s, got %s", expected, qry)
	}
}

func TestSqrlIdentity_Fields(t *testing.T) {
	identity := &SqrlIdentity{
		Idk:      "test-idk",
		Suk:      "test-suk",
		Vuk:      "test-vuk",
		Pidk:     "test-pidk",
		SQRLOnly: true,
		Hardlock: true,
		Disabled: false,
		Rekeyed:  "",
		Btn:      1,
	}

	if identity.Idk != "test-idk" {
		t.Errorf("Expected Idk to be test-idk, got %s", identity.Idk)
	}
	if !identity.SQRLOnly {
		t.Error("Expected SQRLOnly to be true")
	}
	if !identity.Hardlock {
		t.Error("Expected Hardlock to be true")
	}
	if identity.Disabled {
		t.Error("Expected Disabled to be false")
	}
}

func TestHoardCache_Fields(t *testing.T) {
	cache := &HoardCache{
		State:       "authenticated",
		RemoteIP:    "192.168.1.1",
		OriginalNut: "orig-nut",
		PagNut:      "pag-nut",
	}

	if cache.State != "authenticated" {
		t.Errorf("Expected State to be authenticated, got %s", cache.State)
	}
	if cache.RemoteIP != "192.168.1.1" {
		t.Errorf("Expected RemoteIP to be 192.168.1.1, got %s", cache.RemoteIP)
	}
	if cache.OriginalNut != "orig-nut" {
		t.Errorf("Expected OriginalNut to be orig-nut, got %s", cache.OriginalNut)
	}
}

// Mock authenticator for testing
type mockAuthenticator struct {
	authURL       string
	swapError     error
	removeError   error
	askResponse   *Ask
	authenticated bool
}

func (m *mockAuthenticator) AuthenticateIdentity(identity *SqrlIdentity) string {
	m.authenticated = true
	return m.authURL
}

func (m *mockAuthenticator) SwapIdentities(previousIdentity, newIdentity *SqrlIdentity) error {
	return m.swapError
}

func (m *mockAuthenticator) RemoveIdentity(identity *SqrlIdentity) error {
	return m.removeError
}

func (m *mockAuthenticator) AskResponse(identity *SqrlIdentity) *Ask {
	return m.askResponse
}

func TestSqrl64Encoding(t *testing.T) {
	// Test that Sqrl64 is properly configured
	testData := []byte("test data for sqrl64 encoding")
	encoded := Sqrl64.EncodeToString(testData)

	decoded, err := Sqrl64.DecodeString(encoded)
	if err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}

	if string(decoded) != string(testData) {
		t.Errorf("Expected %s, got %s", string(testData), string(decoded))
	}
}

func TestNutType(t *testing.T) {
	nut := Nut("test-nut-value")
	if string(nut) != "test-nut-value" {
		t.Errorf("Expected test-nut-value, got %s", string(nut))
	}
}

func TestErrNotFound(t *testing.T) {
	if ErrNotFound == nil {
		t.Error("ErrNotFound should not be nil")
	}
	if ErrNotFound.Error() != "Not Found" {
		t.Errorf("Expected 'Not Found', got %s", ErrNotFound.Error())
	}
}

func TestSqrlScheme(t *testing.T) {
	if SqrlScheme != "sqrl" {
		t.Errorf("Expected sqrl, got %s", SqrlScheme)
	}
}

func BenchmarkHost(b *testing.B) {
	api := &SqrlSspAPI{
		HostOverride: "example.com",
	}
	req := httptest.NewRequest("GET", "/", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		api.Host(req)
	}
}

func BenchmarkRemoteIP(b *testing.B) {
	api := &SqrlSspAPI{}
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		api.RemoteIP(req)
	}
}
