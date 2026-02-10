package ssp

import (
	"crypto/ed25519"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// testAPI creates a fully configured SqrlSspAPI for testing
func testAPI(t *testing.T) (*SqrlSspAPI, *mockAuthenticator) {
	t.Helper()
	tree, err := NewRandomTree(8)
	if err != nil {
		t.Fatalf("NewRandomTree: %v", err)
	}
	hoard := NewMapHoard()
	authStore := NewMapAuthStore()
	auth := &mockAuthenticator{authURL: "https://example.com/auth"}
	api := NewSqrlSspAPI(tree, hoard, auth, authStore)
	api.HostOverride = "example.com"
	api.RootPath = "/sqrl"
	return api, auth
}

// signedCliRequest creates a properly signed CLI request body for testing.
// Returns the form-encoded body string and the ed25519 public key used.
func signedCliRequest(t *testing.T, cmd string, opts map[string]bool, serverResponse string) (string, ed25519.PublicKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("ed25519.GenerateKey: %v", err)
	}

	idk := Sqrl64.EncodeToString(pub)

	cb := &ClientBody{
		Version: []int{1},
		Cmd:     cmd,
		Opt:     opts,
		Idk:     idk,
	}
	clientEncoded := string(cb.Encode())

	server := serverResponse
	if server == "" {
		server = Sqrl64.EncodeToString([]byte("ver=1\r\nnut=testnut\r\ntif=0\r\nqry=/sqrl/cli.sqrl?nut=testnut\r\n"))
	}

	signingString := clientEncoded + server
	sig := ed25519.Sign(priv, []byte(signingString))
	ids := Sqrl64.EncodeToString(sig)

	form := url.Values{}
	form.Set("client", clientEncoded)
	form.Set("server", server)
	form.Set("ids", ids)
	return form.Encode(), pub
}

// signedCliRequestWithKeys creates a request with specific keys (for identity tracking).
func signedCliRequestWithKeys(t *testing.T, cmd string, opts map[string]bool, serverResponse string, pub ed25519.PublicKey, priv ed25519.PrivateKey) string {
	t.Helper()
	idk := Sqrl64.EncodeToString(pub)

	cb := &ClientBody{
		Version: []int{1},
		Cmd:     cmd,
		Opt:     opts,
		Idk:     idk,
	}
	clientEncoded := string(cb.Encode())

	server := serverResponse
	if server == "" {
		server = Sqrl64.EncodeToString([]byte("ver=1\r\nnut=testnut\r\ntif=0\r\nqry=/sqrl/cli.sqrl?nut=testnut\r\n"))
	}

	signingString := clientEncoded + server
	sig := ed25519.Sign(priv, []byte(signingString))
	ids := Sqrl64.EncodeToString(sig)

	form := url.Values{}
	form.Set("client", clientEncoded)
	form.Set("server", server)
	form.Set("ids", ids)
	return form.Encode()
}

func TestCli_MissingNut(t *testing.T) {
	api, _ := testAPI(t)

	req := httptest.NewRequest("POST", "/sqrl/cli.sqrl", nil)
	w := httptest.NewRecorder()
	api.Cli(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
	if w.Body.Len() == 0 {
		t.Error("Expected response body with error")
	}
}

func TestCli_InvalidRequestBody(t *testing.T) {
	api, _ := testAPI(t)

	// Save a nut so lookup succeeds
	nut, _ := api.tree.Nut()
	_ = api.hoard.Save(nut, &HoardCache{
		State:       "issued",
		RemoteIP:    "127.0.0.1",
		OriginalNut: nut,
		PagNut:      "pag123",
	}, api.NutExpiration)

	req := httptest.NewRequest("POST", fmt.Sprintf("/sqrl/cli.sqrl?nut=%s", nut),
		strings.NewReader("client=invalid&server=invalid&ids=invalid"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	api.Cli(w, req)

	// Should get a response with ClientFailure and CommandFailed flags
	if w.Body.Len() == 0 {
		t.Error("Expected non-empty response body")
	}

	resp, err := ParseCliResponse(w.Body.Bytes())
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if resp.TIF&TIFClientFailure == 0 {
		t.Error("Expected TIFClientFailure flag")
	}
	if resp.TIF&TIFCommandFailed == 0 {
		t.Error("Expected TIFCommandFailed flag")
	}
}

func TestCli_NutNotFound(t *testing.T) {
	api, _ := testAPI(t)

	body, _ := signedCliRequest(t, "query", nil, "")
	req := httptest.NewRequest("POST", "/sqrl/cli.sqrl?nut=nonexistent",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	api.Cli(w, req)

	resp, err := ParseCliResponse(w.Body.Bytes())
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if resp.TIF&TIFClientFailure == 0 {
		t.Error("Expected TIFClientFailure flag for missing nut")
	}
	if resp.TIF&TIFCommandFailed == 0 {
		t.Error("Expected TIFCommandFailed flag for missing nut")
	}
}

func TestCli_QueryNewIdentity(t *testing.T) {
	api, _ := testAPI(t)

	// Create and save a nut
	nut, _ := api.tree.Nut()
	_ = api.hoard.Save(nut, &HoardCache{
		State:       "issued",
		RemoteIP:    "127.0.0.1",
		OriginalNut: nut,
		PagNut:      "pag123",
	}, api.NutExpiration)

	body, _ := signedCliRequest(t, "query", nil, "")
	req := httptest.NewRequest("POST", fmt.Sprintf("/sqrl/cli.sqrl?nut=%s", nut),
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Forwarded-For", "127.0.0.1")
	w := httptest.NewRecorder()
	api.Cli(w, req)

	if w.Body.Len() == 0 {
		t.Fatal("Expected non-empty response body")
	}

	resp, err := ParseCliResponse(w.Body.Bytes())
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// New identity: no ID match expected
	if resp.TIF&TIFIDMatch != 0 {
		t.Error("Did not expect TIFIDMatch for new identity query")
	}
	// Should have IP match since same IP
	if resp.TIF&TIFIPMatched == 0 {
		t.Error("Expected TIFIPMatched for same IP")
	}
}

func TestCli_QueryKnownIdentity(t *testing.T) {
	api, _ := testAPI(t)

	// Generate keys and save identity
	pub, priv, _ := ed25519.GenerateKey(nil)
	idk := Sqrl64.EncodeToString(pub)
	_ = api.authStore.SaveIdentity(&SqrlIdentity{
		Idk: idk,
		Suk: "test-suk",
		Vuk: "test-vuk",
	})

	nut, _ := api.tree.Nut()
	_ = api.hoard.Save(nut, &HoardCache{
		State:       "issued",
		RemoteIP:    "127.0.0.1",
		OriginalNut: nut,
		PagNut:      "pag123",
	}, api.NutExpiration)

	body := signedCliRequestWithKeys(t, "query", nil, "", pub, priv)
	req := httptest.NewRequest("POST", fmt.Sprintf("/sqrl/cli.sqrl?nut=%s", nut),
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Forwarded-For", "127.0.0.1")
	w := httptest.NewRecorder()
	api.Cli(w, req)

	resp, err := ParseCliResponse(w.Body.Bytes())
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.TIF&TIFIDMatch == 0 {
		t.Error("Expected TIFIDMatch for known identity")
	}
}

func TestCli_IdentNewUser(t *testing.T) {
	api, auth := testAPI(t)

	nut, _ := api.tree.Nut()
	pagNut, _ := api.tree.Nut()
	_ = api.hoard.Save(nut, &HoardCache{
		State:       "issued",
		RemoteIP:    "127.0.0.1",
		OriginalNut: nut,
		PagNut:      pagNut,
	}, api.NutExpiration)
	// Also save pagnut for polling
	_ = api.hoard.Save(pagNut, &HoardCache{
		State:       "pending",
		RemoteIP:    "127.0.0.1",
		OriginalNut: nut,
		PagNut:      pagNut,
	}, api.NutExpiration)

	body, _ := signedCliRequest(t, "ident", map[string]bool{"cps": true}, "")
	req := httptest.NewRequest("POST", fmt.Sprintf("/sqrl/cli.sqrl?nut=%s", nut),
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Forwarded-For", "127.0.0.1")
	w := httptest.NewRecorder()
	api.Cli(w, req)

	resp, err := ParseCliResponse(w.Body.Bytes())
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.TIF&TIFIDMatch == 0 {
		t.Error("Expected TIFIDMatch for new ident")
	}
	if !auth.authenticated {
		t.Error("Expected Authenticator.AuthenticateIdentity to be called")
	}
	if resp.URL != "https://example.com/auth" {
		t.Errorf("Expected CPS URL, got %s", resp.URL)
	}
}

func TestCli_DisableCommand(t *testing.T) {
	api, _ := testAPI(t)

	pub, priv, _ := ed25519.GenerateKey(nil)
	idk := Sqrl64.EncodeToString(pub)
	_ = api.authStore.SaveIdentity(&SqrlIdentity{
		Idk:      idk,
		Suk:      "test-suk",
		Vuk:      "test-vuk",
		Disabled: false,
	})

	nut, _ := api.tree.Nut()
	_ = api.hoard.Save(nut, &HoardCache{
		State:       "issued",
		RemoteIP:    "127.0.0.1",
		OriginalNut: nut,
		PagNut:      "pag123",
	}, api.NutExpiration)

	body := signedCliRequestWithKeys(t, "disable", nil, "", pub, priv)
	req := httptest.NewRequest("POST", fmt.Sprintf("/sqrl/cli.sqrl?nut=%s", nut),
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Forwarded-For", "127.0.0.1")
	w := httptest.NewRecorder()
	api.Cli(w, req)

	resp, err := ParseCliResponse(w.Body.Bytes())
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.TIF&TIFSQRLDisabled == 0 {
		t.Error("Expected TIFSQRLDisabled after disable command")
	}

	// Verify identity is disabled in store
	identity, _ := api.authStore.FindIdentity(idk)
	if !identity.Disabled {
		t.Error("Expected identity to be disabled in store")
	}
}

func TestCli_IPMismatch_Rejected(t *testing.T) {
	api, _ := testAPI(t)

	nut, _ := api.tree.Nut()
	_ = api.hoard.Save(nut, &HoardCache{
		State:       "issued",
		RemoteIP:    "10.0.0.1",
		OriginalNut: nut,
		PagNut:      "pag123",
	}, api.NutExpiration)

	body, _ := signedCliRequest(t, "query", nil, "")
	req := httptest.NewRequest("POST", fmt.Sprintf("/sqrl/cli.sqrl?nut=%s", nut),
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Forwarded-For", "192.168.1.1") // Different IP
	w := httptest.NewRecorder()
	api.Cli(w, req)

	resp, err := ParseCliResponse(w.Body.Bytes())
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.TIF&TIFCommandFailed == 0 {
		t.Error("Expected TIFCommandFailed for IP mismatch")
	}
	if resp.TIF&TIFIPMatched != 0 {
		t.Error("Did not expect TIFIPMatched for IP mismatch")
	}
}

func TestCli_IPMismatch_NoIPTest(t *testing.T) {
	api, _ := testAPI(t)

	nut, _ := api.tree.Nut()
	_ = api.hoard.Save(nut, &HoardCache{
		State:       "issued",
		RemoteIP:    "10.0.0.1",
		OriginalNut: nut,
		PagNut:      "pag123",
	}, api.NutExpiration)

	body, _ := signedCliRequest(t, "query", map[string]bool{"noiptest": true}, "")
	req := httptest.NewRequest("POST", fmt.Sprintf("/sqrl/cli.sqrl?nut=%s", nut),
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Forwarded-For", "192.168.1.1") // Different IP but noiptest set
	w := httptest.NewRecorder()
	api.Cli(w, req)

	resp, err := ParseCliResponse(w.Body.Bytes())
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Should NOT fail on IP mismatch when noiptest is set
	if resp.TIF&TIFCommandFailed != 0 {
		t.Error("Did not expect TIFCommandFailed with noiptest option")
	}
}

func TestCli_UnsupportedCommand(t *testing.T) {
	api, _ := testAPI(t)

	nut, _ := api.tree.Nut()
	_ = api.hoard.Save(nut, &HoardCache{
		State:       "issued",
		RemoteIP:    "127.0.0.1",
		OriginalNut: nut,
		PagNut:      "pag123",
	}, api.NutExpiration)

	body, _ := signedCliRequest(t, "badcommand", nil, "")
	req := httptest.NewRequest("POST", fmt.Sprintf("/sqrl/cli.sqrl?nut=%s", nut),
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Forwarded-For", "127.0.0.1")
	w := httptest.NewRecorder()
	api.Cli(w, req)

	resp, err := ParseCliResponse(w.Body.Bytes())
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.TIF&TIFFunctionNotSupported == 0 {
		t.Error("Expected TIFFunctionNotSupported for unknown command")
	}
}

func TestCli_RekeyedIdentity(t *testing.T) {
	api, _ := testAPI(t)

	pub, priv, _ := ed25519.GenerateKey(nil)
	idk := Sqrl64.EncodeToString(pub)
	_ = api.authStore.SaveIdentity(&SqrlIdentity{
		Idk:     idk,
		Suk:     "test-suk",
		Vuk:     "test-vuk",
		Rekeyed: "new-idk-value", // This identity was rekeyed
	})

	nut, _ := api.tree.Nut()
	_ = api.hoard.Save(nut, &HoardCache{
		State:       "issued",
		RemoteIP:    "127.0.0.1",
		OriginalNut: nut,
		PagNut:      "pag123",
	}, api.NutExpiration)

	body := signedCliRequestWithKeys(t, "query", nil, "", pub, priv)
	req := httptest.NewRequest("POST", fmt.Sprintf("/sqrl/cli.sqrl?nut=%s", nut),
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Forwarded-For", "127.0.0.1")
	w := httptest.NewRecorder()
	api.Cli(w, req)

	resp, err := ParseCliResponse(w.Body.Bytes())
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.TIF&TIFIdentitySuperseded == 0 {
		t.Error("Expected TIFIdentitySuperseded for rekeyed identity")
	}
}

func TestCli_RekeyedIdentity_NonQueryFails(t *testing.T) {
	api, _ := testAPI(t)

	pub, priv, _ := ed25519.GenerateKey(nil)
	idk := Sqrl64.EncodeToString(pub)
	_ = api.authStore.SaveIdentity(&SqrlIdentity{
		Idk:     idk,
		Suk:     "test-suk",
		Vuk:     "test-vuk",
		Rekeyed: "new-idk-value",
	})

	nut, _ := api.tree.Nut()
	_ = api.hoard.Save(nut, &HoardCache{
		State:       "issued",
		RemoteIP:    "127.0.0.1",
		OriginalNut: nut,
		PagNut:      "pag123",
	}, api.NutExpiration)

	body := signedCliRequestWithKeys(t, "ident", nil, "", pub, priv)
	req := httptest.NewRequest("POST", fmt.Sprintf("/sqrl/cli.sqrl?nut=%s", nut),
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Forwarded-For", "127.0.0.1")
	w := httptest.NewRecorder()
	api.Cli(w, req)

	resp, err := ParseCliResponse(w.Body.Bytes())
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.TIF&TIFIdentitySuperseded == 0 {
		t.Error("Expected TIFIdentitySuperseded")
	}
	if resp.TIF&TIFCommandFailed == 0 {
		t.Error("Expected TIFCommandFailed for non-query on rekeyed identity")
	}
}

func TestCli_SukOption(t *testing.T) {
	api, _ := testAPI(t)

	pub, priv, _ := ed25519.GenerateKey(nil)
	idk := Sqrl64.EncodeToString(pub)
	_ = api.authStore.SaveIdentity(&SqrlIdentity{
		Idk: idk,
		Suk: "the-suk-value",
		Vuk: "test-vuk",
	})

	nut, _ := api.tree.Nut()
	_ = api.hoard.Save(nut, &HoardCache{
		State:       "issued",
		RemoteIP:    "127.0.0.1",
		OriginalNut: nut,
		PagNut:      "pag123",
	}, api.NutExpiration)

	body := signedCliRequestWithKeys(t, "query", map[string]bool{"suk": true}, "", pub, priv)
	req := httptest.NewRequest("POST", fmt.Sprintf("/sqrl/cli.sqrl?nut=%s", nut),
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Forwarded-For", "127.0.0.1")
	w := httptest.NewRecorder()
	api.Cli(w, req)

	resp, err := ParseCliResponse(w.Body.Bytes())
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Suk != "the-suk-value" {
		t.Errorf("Expected suk 'the-suk-value', got '%s'", resp.Suk)
	}
}

func TestCli_IdentDisabledAccount(t *testing.T) {
	api, _ := testAPI(t)

	pub, priv, _ := ed25519.GenerateKey(nil)
	idk := Sqrl64.EncodeToString(pub)
	_ = api.authStore.SaveIdentity(&SqrlIdentity{
		Idk:      idk,
		Suk:      "test-suk",
		Vuk:      "test-vuk",
		Disabled: true,
	})

	nut, _ := api.tree.Nut()
	_ = api.hoard.Save(nut, &HoardCache{
		State:       "issued",
		RemoteIP:    "127.0.0.1",
		OriginalNut: nut,
		PagNut:      "pag123",
	}, api.NutExpiration)

	body := signedCliRequestWithKeys(t, "ident", nil, "", pub, priv)
	req := httptest.NewRequest("POST", fmt.Sprintf("/sqrl/cli.sqrl?nut=%s", nut),
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Forwarded-For", "127.0.0.1")
	w := httptest.NewRecorder()
	api.Cli(w, req)

	resp, err := ParseCliResponse(w.Body.Bytes())
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.TIF&TIFSQRLDisabled == 0 {
		t.Error("Expected TIFSQRLDisabled for disabled account")
	}
	if resp.TIF&TIFCommandFailed == 0 {
		t.Error("Expected TIFCommandFailed for ident on disabled account")
	}
}

// Test api.go identity management functions
func TestSwapIdentities(t *testing.T) {
	api, _ := testAPI(t)

	prev := &SqrlIdentity{Idk: "old-idk", Suk: "old-suk", Vuk: "old-vuk"}
	newIdent := &SqrlIdentity{Idk: "new-idk", Suk: "new-suk", Vuk: "new-vuk"}

	_ = api.authStore.SaveIdentity(prev)
	err := api.swapIdentities(prev, newIdent)
	if err != nil {
		t.Fatalf("swapIdentities failed: %v", err)
	}

	if prev.Rekeyed != "new-idk" {
		t.Errorf("Expected Rekeyed to be new-idk, got %s", prev.Rekeyed)
	}

	// Verify old identity was saved with rekeyed flag
	found, _ := api.authStore.FindIdentity("old-idk")
	if found.Rekeyed != "new-idk" {
		t.Errorf("Expected saved identity to have Rekeyed=new-idk, got %s", found.Rekeyed)
	}
}

func TestSwapIdentities_Error(t *testing.T) {
	api, auth := testAPI(t)
	auth.swapError = fmt.Errorf("swap error")

	prev := &SqrlIdentity{Idk: "old-idk"}
	newIdent := &SqrlIdentity{Idk: "new-idk"}

	err := api.swapIdentities(prev, newIdent)
	if err == nil {
		t.Error("Expected error from swapIdentities")
	}
}

func TestRemoveIdentity(t *testing.T) {
	api, _ := testAPI(t)

	identity := &SqrlIdentity{Idk: "remove-me", Suk: "suk", Vuk: "vuk"}
	_ = api.authStore.SaveIdentity(identity)

	err := api.removeIdentity(identity)
	if err != nil {
		t.Fatalf("removeIdentity failed: %v", err)
	}

	_, err = api.authStore.FindIdentity("remove-me")
	if err != ErrNotFound {
		t.Errorf("Expected identity to be deleted, got %v", err)
	}
}

func TestRemoveIdentity_Error(t *testing.T) {
	api, auth := testAPI(t)
	auth.removeError = fmt.Errorf("remove error")

	identity := &SqrlIdentity{Idk: "remove-me"}
	err := api.removeIdentity(identity)
	if err == nil {
		t.Error("Expected error from removeIdentity")
	}
}

func TestAuthenticateIdentity(t *testing.T) {
	api, auth := testAPI(t)
	auth.authURL = "https://example.com/dashboard"

	identity := &SqrlIdentity{Idk: "auth-idk", Suk: "suk", Vuk: "vuk"}
	redirect, err := api.authenticateIdentity(identity, -1)
	if err != nil {
		t.Fatalf("authenticateIdentity failed: %v", err)
	}
	if redirect != "https://example.com/dashboard" {
		t.Errorf("Expected redirect URL, got %s", redirect)
	}
	if !auth.authenticated {
		t.Error("Expected authenticator to be called")
	}
}

func TestWriteResponse(t *testing.T) {
	api, _ := testAPI(t)

	nut, _ := api.tree.Nut()
	response := NewCliResponse(nut, api.qry(nut))
	response.HoardCache = &HoardCache{
		State:       "issued",
		RemoteIP:    "127.0.0.1",
		OriginalNut: nut,
		PagNut:      "pag123",
	}

	req := &CliRequest{
		Client: &ClientBody{
			Version: []int{1},
			Cmd:     "query",
			Idk:     "testidk",
		},
	}

	w := httptest.NewRecorder()
	api.writeResponse(req, response, w)

	if w.Body.Len() == 0 {
		t.Error("Expected non-empty response body")
	}

	// Verify the nut was saved in hoard
	_, err := api.hoard.Get(nut)
	if err != nil {
		t.Errorf("Expected nut to be saved in hoard after writeResponse: %v", err)
	}
}

func TestSetSuk_WithKnownIdentity(t *testing.T) {
	api, _ := testAPI(t)
	_ = api

	req := &CliRequest{
		Client: &ClientBody{
			Cmd: "query",
			Opt: map[string]bool{"suk": true},
		},
	}
	response := NewCliResponse("testnut", "/sqrl/cli.sqrl?nut=testnut")
	identity := &SqrlIdentity{Suk: "identity-suk-value"}

	api.setSuk(req, response, identity)

	if response.Suk != "identity-suk-value" {
		t.Errorf("Expected Suk from identity, got %s", response.Suk)
	}
}

func TestSetSuk_IdentWithNoIdentity(t *testing.T) {
	api, _ := testAPI(t)

	req := &CliRequest{
		Client: &ClientBody{
			Cmd: "ident",
			Opt: map[string]bool{"suk": true},
			Suk: "request-suk-value",
		},
	}
	response := NewCliResponse("testnut", "/sqrl/cli.sqrl?nut=testnut")

	api.setSuk(req, response, nil)

	if response.Suk != "request-suk-value" {
		t.Errorf("Expected Suk from request, got %s", response.Suk)
	}
}

func TestSetSuk_NoOption(t *testing.T) {
	api, _ := testAPI(t)

	req := &CliRequest{
		Client: &ClientBody{
			Cmd: "query",
			Opt: map[string]bool{},
		},
	}
	response := NewCliResponse("testnut", "/sqrl/cli.sqrl?nut=testnut")
	identity := &SqrlIdentity{Suk: "identity-suk-value"}

	api.setSuk(req, response, identity)

	if response.Suk != "" {
		t.Errorf("Expected empty Suk when option not set, got %s", response.Suk)
	}
}

func TestCheckPreviousIdentity_Found(t *testing.T) {
	api, _ := testAPI(t)

	_ = api.authStore.SaveIdentity(&SqrlIdentity{
		Idk: "previous-idk",
		Suk: "prev-suk",
		Vuk: "prev-vuk",
	})

	req := &CliRequest{
		Client: &ClientBody{
			Pidk: "previous-idk",
			Opt:  map[string]bool{},
		},
	}
	response := NewCliResponse("testnut", "/sqrl/cli.sqrl?nut=testnut")

	prev, err := api.checkPreviousIdentity(req, response)
	if err != nil {
		t.Fatalf("checkPreviousIdentity failed: %v", err)
	}
	if prev == nil {
		t.Fatal("Expected previous identity to be found")
	}
	if response.TIF&TIFPreviousIDMatch == 0 {
		t.Error("Expected TIFPreviousIDMatch flag")
	}
}

func TestCheckPreviousIdentity_NotFound(t *testing.T) {
	api, _ := testAPI(t)

	req := &CliRequest{
		Client: &ClientBody{
			Pidk: "nonexistent-pidk",
		},
	}
	response := NewCliResponse("testnut", "/sqrl/cli.sqrl?nut=testnut")

	prev, err := api.checkPreviousIdentity(req, response)
	if err != nil {
		t.Fatalf("checkPreviousIdentity failed: %v", err)
	}
	if prev != nil {
		t.Error("Expected no previous identity")
	}
}

func TestCheckPreviousIdentity_NoPidk(t *testing.T) {
	api, _ := testAPI(t)

	req := &CliRequest{
		Client: &ClientBody{
			Pidk: "",
		},
	}
	response := NewCliResponse("testnut", "/sqrl/cli.sqrl?nut=testnut")

	prev, err := api.checkPreviousIdentity(req, response)
	if err != nil {
		t.Fatalf("checkPreviousIdentity failed: %v", err)
	}
	if prev != nil {
		t.Error("Expected no previous identity when pidk is empty")
	}
}

func TestCheckPreviousSwap_Success(t *testing.T) {
	api, _ := testAPI(t)

	prev := &SqrlIdentity{Idk: "prev-idk", Suk: "suk", Vuk: "vuk"}
	_ = api.authStore.SaveIdentity(prev)

	newIdent := &SqrlIdentity{Idk: "new-idk", Suk: "suk2", Vuk: "vuk2"}
	response := NewCliResponse("testnut", "/sqrl/cli.sqrl?nut=testnut")
	response.TIF |= TIFPreviousIDMatch

	err := api.checkPreviousSwap(prev, newIdent, response)
	if err != nil {
		t.Fatalf("checkPreviousSwap failed: %v", err)
	}

	// PreviousIDMatch should be cleared after swap
	if response.TIF&TIFPreviousIDMatch != 0 {
		t.Error("Expected TIFPreviousIDMatch to be cleared after swap")
	}
}

func TestCheckPreviousSwap_NoPrevious(t *testing.T) {
	api, _ := testAPI(t)

	newIdent := &SqrlIdentity{Idk: "new-idk"}
	response := NewCliResponse("testnut", "/sqrl/cli.sqrl?nut=testnut")

	err := api.checkPreviousSwap(nil, newIdent, response)
	if err != nil {
		t.Fatalf("checkPreviousSwap should succeed with nil previous: %v", err)
	}
}

func TestFinishCliResponse_AuthNilIdentity(t *testing.T) {
	api, _ := testAPI(t)

	req := &CliRequest{
		Client: &ClientBody{
			Cmd: "ident",
		},
	}
	response := NewCliResponse("testnut", "/sqrl/cli.sqrl?nut=testnut")
	hoardCache := &HoardCache{
		State:       "issued",
		RemoteIP:    "127.0.0.1",
		OriginalNut: "orignut",
		PagNut:      "pagnut",
	}

	api.finishCliResponse(req, response, nil, hoardCache)

	if response.TIF&TIFClientFailure == 0 {
		t.Error("Expected TIFClientFailure for nil identity on auth command")
	}
	if response.TIF&TIFCommandFailed == 0 {
		t.Error("Expected TIFCommandFailed for nil identity on auth command")
	}
}

func TestFinishCliResponse_NonCPS(t *testing.T) {
	api, _ := testAPI(t)

	identity := &SqrlIdentity{Idk: "auth-idk", Suk: "suk", Vuk: "vuk"}
	_ = api.authStore.SaveIdentity(identity)

	req := &CliRequest{
		Client: &ClientBody{
			Cmd: "ident",
			Opt: map[string]bool{},
		},
	}
	response := NewCliResponse("testnut", "/sqrl/cli.sqrl?nut=testnut")
	hoardCache := &HoardCache{
		State:       "issued",
		RemoteIP:    "127.0.0.1",
		OriginalNut: "orignut",
		PagNut:      "pagnut",
	}

	api.finishCliResponse(req, response, identity, hoardCache)

	// Non-CPS: should save to pagnut for polling
	pagCache, err := api.hoard.Get(Nut("pagnut"))
	if err != nil {
		t.Fatalf("Expected pagnut to be saved: %v", err)
	}
	if pagCache.State != "authenticated" {
		t.Errorf("Expected pagnut state 'authenticated', got '%s'", pagCache.State)
	}
}

func TestRequestValidations_IPMatch(t *testing.T) {
	api, _ := testAPI(t)

	hoardCache := &HoardCache{
		RemoteIP: "127.0.0.1",
	}
	req := &CliRequest{
		Client: &ClientBody{
			Cmd: "query",
			Opt: map[string]bool{},
		},
	}
	response := NewCliResponse("testnut", "/sqrl/cli.sqrl?nut=testnut")

	httpReq := httptest.NewRequest("POST", "/cli.sqrl", nil)
	httpReq.Header.Set("X-Forwarded-For", "127.0.0.1")

	err := api.requestValidations(hoardCache, req, httpReq, response)
	if err != nil {
		t.Fatalf("requestValidations failed: %v", err)
	}
	if response.TIF&TIFIPMatched == 0 {
		t.Error("Expected TIFIPMatched")
	}
}

func TestRequestValidations_IDMismatch(t *testing.T) {
	api, _ := testAPI(t)

	hoardCache := &HoardCache{
		RemoteIP: "127.0.0.1",
		LastRequest: &CliRequest{
			Client: &ClientBody{
				Idk: "original-idk",
			},
		},
	}
	req := &CliRequest{
		Client: &ClientBody{
			Cmd: "query",
			Opt: map[string]bool{},
			Idk: "different-idk",
		},
	}
	response := NewCliResponse("testnut", "/sqrl/cli.sqrl?nut=testnut")

	httpReq := httptest.NewRequest("POST", "/cli.sqrl", nil)
	httpReq.Header.Set("X-Forwarded-For", "127.0.0.1")

	err := api.requestValidations(hoardCache, req, httpReq, response)
	if err == nil {
		t.Error("Expected error for ID mismatch")
	}
	if response.TIF&TIFBadIDAssociation == 0 {
		t.Error("Expected TIFBadIDAssociation")
	}
}

func TestKnownIdentity_Query(t *testing.T) {
	api, _ := testAPI(t)

	identity := &SqrlIdentity{
		Idk: "known-idk",
		Suk: "suk",
		Vuk: "vuk",
	}

	req := &CliRequest{
		Client: &ClientBody{
			Cmd: "query",
			Opt: map[string]bool{},
			Btn: -1,
		},
	}
	response := NewCliResponse("testnut", "/sqrl/cli.sqrl?nut=testnut")

	err := api.knownIdentity(req, response, identity)
	if err != nil {
		t.Fatalf("knownIdentity failed: %v", err)
	}
	if response.TIF&TIFIDMatch == 0 {
		t.Error("Expected TIFIDMatch for known identity")
	}
}

// Benchmark the full CLI flow
func BenchmarkCli_Query(b *testing.B) {
	tree, _ := NewRandomTree(8)
	hoard := NewMapHoard()
	authStore := NewMapAuthStore()
	auth := &mockAuthenticator{authURL: "https://example.com/auth"}
	api := NewSqrlSspAPI(tree, hoard, auth, authStore)
	api.HostOverride = "example.com"
	api.RootPath = "/sqrl"

	pub, priv, _ := ed25519.GenerateKey(nil)
	idk := Sqrl64.EncodeToString(pub)
	cb := &ClientBody{Version: []int{1}, Cmd: "query", Idk: idk}
	clientEncoded := string(cb.Encode())
	server := Sqrl64.EncodeToString([]byte("ver=1\r\nnut=x\r\ntif=0\r\nqry=/sqrl/cli.sqrl?nut=x\r\n"))
	sig := ed25519.Sign(priv, []byte(clientEncoded+server))
	ids := Sqrl64.EncodeToString(sig)
	form := url.Values{}
	form.Set("client", clientEncoded)
	form.Set("server", server)
	form.Set("ids", ids)
	body := form.Encode()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		nut, _ := api.tree.Nut()
		_ = api.hoard.Save(nut, &HoardCache{
			State: "issued", RemoteIP: "127.0.0.1",
			OriginalNut: nut, PagNut: "pag",
		}, 10*time.Minute)

		req := httptest.NewRequest("POST", fmt.Sprintf("/sqrl/cli.sqrl?nut=%s", nut),
			strings.NewReader(body))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("X-Forwarded-For", "127.0.0.1")
		w := httptest.NewRecorder()
		api.Cli(w, req)
	}
}
