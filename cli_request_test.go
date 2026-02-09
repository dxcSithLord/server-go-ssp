package ssp

import (
	"crypto/ed25519"
	"testing"
)

func TestParseSqrlQuery_Valid(t *testing.T) {
	query := "ver=1\r\ncmd=query\r\nopt=sqrlonly~hardlock\r\nidk=testkey123\r\n"
	params, err := ParseSqrlQuery(query)

	if err != nil {
		t.Fatalf("ParseSqrlQuery failed: %v", err)
	}

	if params["ver"] != "1" {
		t.Errorf("Expected ver=1, got %s", params["ver"])
	}
	if params["cmd"] != "query" {
		t.Errorf("Expected cmd=query, got %s", params["cmd"])
	}
	if params["opt"] != "sqrlonly~hardlock" {
		t.Errorf("Expected opt=sqrlonly~hardlock, got %s", params["opt"])
	}
	if params["idk"] != "testkey123" {
		t.Errorf("Expected idk=testkey123, got %s", params["idk"])
	}
}

func TestParseSqrlQuery_Empty(t *testing.T) {
	params, err := ParseSqrlQuery("")
	if err != nil {
		t.Fatalf("ParseSqrlQuery failed: %v", err)
	}
	if len(params) != 0 {
		t.Errorf("Expected empty params, got %v", params)
	}
}

func TestParseSqrlQuery_NoValue(t *testing.T) {
	query := "key\r\n"
	params, err := ParseSqrlQuery(query)

	if err != nil {
		t.Fatalf("ParseSqrlQuery failed: %v", err)
	}
	if params["key"] != "" {
		t.Errorf("Expected empty value, got %s", params["key"])
	}
}

func TestClientBodyFromParams_Valid(t *testing.T) {
	params := map[string]string{
		"ver":  "1",
		"cmd":  "ident",
		"opt":  "sqrlonly~hardlock",
		"idk":  "testidk",
		"suk":  "testsuk",
		"vuk":  "testvuk",
		"pidk": "testpidk",
		"btn":  "2",
	}

	cb, err := ClientBodyFromParams(params)
	if err != nil {
		t.Fatalf("ClientBodyFromParams failed: %v", err)
	}

	if cb.Version[0] != 1 {
		t.Errorf("Expected version 1, got %v", cb.Version)
	}
	if cb.Cmd != "ident" {
		t.Errorf("Expected cmd ident, got %s", cb.Cmd)
	}
	if !cb.Opt["sqrlonly"] {
		t.Error("Expected sqrlonly option to be true")
	}
	if !cb.Opt["hardlock"] {
		t.Error("Expected hardlock option to be true")
	}
	if cb.Idk != "testidk" {
		t.Errorf("Expected idk testidk, got %s", cb.Idk)
	}
	if cb.Suk != "testsuk" {
		t.Errorf("Expected suk testsuk, got %s", cb.Suk)
	}
	if cb.Vuk != "testvuk" {
		t.Errorf("Expected vuk testvuk, got %s", cb.Vuk)
	}
	if cb.Pidk != "testpidk" {
		t.Errorf("Expected pidk testpidk, got %s", cb.Pidk)
	}
	if cb.Btn != 2 {
		t.Errorf("Expected btn 2, got %d", cb.Btn)
	}
}

func TestClientBodyFromParams_InvalidVersion(t *testing.T) {
	params := map[string]string{
		"ver": "invalid",
		"cmd": "ident",
	}

	_, err := ClientBodyFromParams(params)
	if err == nil {
		t.Error("Expected error for invalid version")
	}
}

func TestClientBodyFromParams_NoBtn(t *testing.T) {
	params := map[string]string{
		"ver": "1",
		"cmd": "query",
	}

	cb, err := ClientBodyFromParams(params)
	if err != nil {
		t.Fatalf("ClientBodyFromParams failed: %v", err)
	}
	if cb.Btn != -1 {
		t.Errorf("Expected btn -1 (no value), got %d", cb.Btn)
	}
}

func TestClientBody_Encode(t *testing.T) {
	cb := &ClientBody{
		Version: []int{1},
		Cmd:     "ident",
		Opt:     map[string]bool{"sqrlonly": true},
		Idk:     "testidk",
		Suk:     "testsuk",
		Vuk:     "testvuk",
		Pidk:    "testpidk",
	}

	encoded := cb.Encode()
	if len(encoded) == 0 {
		t.Error("Expected non-empty encoded result")
	}

	// Verify it can be decoded
	decoded, err := Sqrl64.DecodeString(string(encoded))
	if err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}

	decodedStr := string(decoded)
	if len(decodedStr) == 0 {
		t.Error("Expected non-empty decoded string")
	}
}

func TestClientBody_EncodeNil(t *testing.T) {
	var cb *ClientBody
	encoded := cb.Encode()
	if len(encoded) != 0 {
		t.Errorf("Expected empty result for nil ClientBody, got %s", string(encoded))
	}
}

func TestCliRequest_Identity(t *testing.T) {
	cr := &CliRequest{
		Client: &ClientBody{
			Idk:  "test-idk",
			Suk:  "test-suk",
			Vuk:  "test-vuk",
			Pidk: "test-pidk",
			Opt:  map[string]bool{"sqrlonly": true, "hardlock": true},
			Btn:  1,
		},
	}

	identity := cr.Identity()

	if identity.Idk != "test-idk" {
		t.Errorf("Expected Idk test-idk, got %s", identity.Idk)
	}
	if identity.Suk != "test-suk" {
		t.Errorf("Expected Suk test-suk, got %s", identity.Suk)
	}
	if identity.Vuk != "test-vuk" {
		t.Errorf("Expected Vuk test-vuk, got %s", identity.Vuk)
	}
	if identity.Pidk != "test-pidk" {
		t.Errorf("Expected Pidk test-pidk, got %s", identity.Pidk)
	}
	if !identity.SQRLOnly {
		t.Error("Expected SQRLOnly to be true")
	}
	if !identity.Hardlock {
		t.Error("Expected Hardlock to be true")
	}
	if identity.Btn != 1 {
		t.Errorf("Expected Btn 1, got %d", identity.Btn)
	}
}

func TestCliRequest_SigningString(t *testing.T) {
	cr := &CliRequest{
		ClientEncoded: "client123",
		Server:        "server456",
	}

	signingString := cr.SigningString()
	expected := "client123server456"
	if string(signingString) != expected {
		t.Errorf("Expected %s, got %s", expected, string(signingString))
	}
}

func TestCliRequest_UpdateIdentity(t *testing.T) {
	identity := &SqrlIdentity{
		SQRLOnly: false,
		Hardlock: false,
	}

	cr := &CliRequest{
		Client: &ClientBody{
			Opt: map[string]bool{"sqrlonly": true, "hardlock": true},
		},
	}

	changed := cr.UpdateIdentity(identity)

	if changed {
		t.Error("Expected changed to be false (identity was modified)")
	}
	if !identity.SQRLOnly {
		t.Error("Expected SQRLOnly to be updated to true")
	}
	if !identity.Hardlock {
		t.Error("Expected Hardlock to be updated to true")
	}
}

func TestCliRequest_UpdateIdentity_NoChange(t *testing.T) {
	identity := &SqrlIdentity{
		SQRLOnly: true,
		Hardlock: true,
	}

	cr := &CliRequest{
		Client: &ClientBody{
			Opt: map[string]bool{"sqrlonly": true, "hardlock": true},
		},
	}

	changed := cr.UpdateIdentity(identity)

	if !changed {
		t.Error("Expected changed to be true (identity was not modified)")
	}
}

func TestCliRequest_IsAuthCommand_Ident(t *testing.T) {
	cr := &CliRequest{
		Client: &ClientBody{
			Cmd: "ident",
		},
	}

	if !cr.IsAuthCommand() {
		t.Error("Expected ident to be an auth command")
	}
}

func TestCliRequest_IsAuthCommand_Enable(t *testing.T) {
	cr := &CliRequest{
		Client: &ClientBody{
			Cmd: "enable",
		},
	}

	if !cr.IsAuthCommand() {
		t.Error("Expected enable to be an auth command")
	}
}

func TestCliRequest_IsAuthCommand_Query(t *testing.T) {
	cr := &CliRequest{
		Client: &ClientBody{
			Cmd: "query",
		},
	}

	if cr.IsAuthCommand() {
		t.Error("Expected query to not be an auth command")
	}
}

func TestCliRequest_ValidateLastResponse_Match(t *testing.T) {
	cr := &CliRequest{
		Server: "test-server-response",
	}

	lastResponse := []byte("test-server-response")
	if !cr.ValidateLastResponse(lastResponse) {
		t.Error("Expected response to match")
	}
}

func TestCliRequest_ValidateLastResponse_NoMatch(t *testing.T) {
	cr := &CliRequest{
		Server: "test-server-response",
	}

	lastResponse := []byte("different-response")
	if cr.ValidateLastResponse(lastResponse) {
		t.Error("Expected response to not match")
	}
}

func TestCliRequest_Encode(t *testing.T) {
	cr := &CliRequest{
		ClientEncoded: "encoded-client",
		Server:        "server-data",
		Ids:           "signature-ids",
		Pids:          "signature-pids",
		Urs:           "signature-urs",
	}

	encoded := cr.Encode()
	if encoded == "" {
		t.Error("Expected non-empty encoded result")
	}
}

func TestCliRequest_Encode_WithClient(t *testing.T) {
	cr := &CliRequest{
		Client: &ClientBody{
			Version: []int{1},
			Cmd:     "query",
			Idk:     "testidk",
		},
		Server: "server-data",
		Ids:    "signature-ids",
	}

	encoded := cr.Encode()
	if encoded == "" {
		t.Error("Expected non-empty encoded result")
	}
}

func BenchmarkParseSqrlQuery(b *testing.B) {
	query := "ver=1\r\ncmd=query\r\nopt=sqrlonly~hardlock\r\nidk=testkey123\r\nsuk=testsuk\r\nvuk=testvuk\r\n"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseSqrlQuery(query)
	}
}

func BenchmarkClientBodyFromParams(b *testing.B) {
	params := map[string]string{
		"ver":  "1",
		"cmd":  "ident",
		"opt":  "sqrlonly~hardlock",
		"idk":  "testidk",
		"suk":  "testsuk",
		"vuk":  "testvuk",
		"pidk": "testpidk",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ClientBodyFromParams(params)
	}
}

func BenchmarkCliRequestIdentity(b *testing.B) {
	cr := &CliRequest{
		Client: &ClientBody{
			Idk:  "test-idk",
			Suk:  "test-suk",
			Vuk:  "test-vuk",
			Pidk: "test-pidk",
			Opt:  map[string]bool{"sqrlonly": true, "hardlock": true},
			Btn:  1,
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cr.Identity()
	}
}

func BenchmarkClientBodyEncode(b *testing.B) {
	cb := &ClientBody{
		Version: []int{1},
		Cmd:     "ident",
		Opt:     map[string]bool{"sqrlonly": true},
		Idk:     "testidk",
		Suk:     "testsuk",
		Vuk:     "testvuk",
		Pidk:    "testpidk",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb.Encode()
	}
}

// Signature verification tests

func TestVerifySignature_Valid(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	idk := Sqrl64.EncodeToString(pub)
	cb := &ClientBody{
		Version: []int{1},
		Cmd:     "query",
		Idk:     idk,
	}
	clientEncoded := string(cb.Encode())
	server := Sqrl64.EncodeToString([]byte("ver=1\r\nnut=test\r\ntif=0\r\n"))

	signingString := clientEncoded + server
	sig := ed25519.Sign(priv, []byte(signingString))
	ids := Sqrl64.EncodeToString(sig)

	cr := &CliRequest{
		Client:        cb,
		ClientEncoded: clientEncoded,
		Server:        server,
		Ids:           ids,
	}

	err = cr.VerifySignature()
	if err != nil {
		t.Errorf("Expected valid signature, got error: %v", err)
	}
}

func TestVerifySignature_Invalid(t *testing.T) {
	pub, _, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	// Sign with a different key
	_, wrongPriv, _ := ed25519.GenerateKey(nil)

	idk := Sqrl64.EncodeToString(pub)
	cb := &ClientBody{
		Version: []int{1},
		Cmd:     "query",
		Idk:     idk,
	}
	clientEncoded := string(cb.Encode())
	server := "server-data"

	signingString := clientEncoded + server
	sig := ed25519.Sign(wrongPriv, []byte(signingString))
	ids := Sqrl64.EncodeToString(sig)

	cr := &CliRequest{
		Client:        cb,
		ClientEncoded: clientEncoded,
		Server:        server,
		Ids:           ids,
	}

	err = cr.VerifySignature()
	if err == nil {
		t.Error("Expected signature verification to fail with wrong key")
	}
}

func TestVerifySignature_InvalidIds(t *testing.T) {
	pub, _, _ := ed25519.GenerateKey(nil)
	idk := Sqrl64.EncodeToString(pub)

	cr := &CliRequest{
		Client: &ClientBody{
			Version: []int{1},
			Cmd:     "query",
			Idk:     idk,
		},
		ClientEncoded: "client",
		Server:        "server",
		Ids:           "!!!not-base64!!!",
	}

	err := cr.VerifySignature()
	if err == nil {
		t.Error("Expected error for invalid ids encoding")
	}
}

func TestVerifySignature_InvalidIdk(t *testing.T) {
	cr := &CliRequest{
		Client: &ClientBody{
			Version: []int{1},
			Cmd:     "query",
			Idk:     "too-short", // Not a valid 32-byte key
		},
		ClientEncoded: "client",
		Server:        "server",
		Ids:           Sqrl64.EncodeToString([]byte("fake-sig-64-bytes-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")),
	}

	err := cr.VerifySignature()
	if err == nil {
		t.Error("Expected error for invalid idk")
	}
}

func TestVerifySignature_WithPidk(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(nil)
	pidkPub, pidkPriv, _ := ed25519.GenerateKey(nil)

	idk := Sqrl64.EncodeToString(pub)
	pidk := Sqrl64.EncodeToString(pidkPub)

	cb := &ClientBody{
		Version: []int{1},
		Cmd:     "ident",
		Idk:     idk,
		Pidk:    pidk,
	}
	clientEncoded := string(cb.Encode())
	server := "server-data"

	signingString := clientEncoded + server
	idsSig := ed25519.Sign(priv, []byte(signingString))
	pidsSig := ed25519.Sign(pidkPriv, []byte(signingString))

	cr := &CliRequest{
		Client:        cb,
		ClientEncoded: clientEncoded,
		Server:        server,
		Ids:           Sqrl64.EncodeToString(idsSig),
		Pids:          Sqrl64.EncodeToString(pidsSig),
	}

	err := cr.VerifySignature()
	if err != nil {
		t.Errorf("Expected valid signature with pidk, got: %v", err)
	}
}

func TestVerifyPidsSignature_InvalidPidk(t *testing.T) {
	cr := &CliRequest{
		Client: &ClientBody{
			Pidk: "too-short",
		},
		ClientEncoded: "client",
		Server:        "server",
		Pids:          Sqrl64.EncodeToString([]byte("fake")),
	}

	err := cr.VerifyPidsSignature()
	if err == nil {
		t.Error("Expected error for invalid pidk")
	}
}

func TestVerifyUrs_Valid(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(nil)
	vuk := Sqrl64.EncodeToString(pub)

	cr := &CliRequest{
		ClientEncoded: "client-data",
		Server:        "server-data",
	}
	signingString := cr.SigningString()
	ursSig := ed25519.Sign(priv, signingString)

	cr.Urs = Sqrl64.EncodeToString(ursSig)

	err := cr.VerifyUrs(vuk)
	if err != nil {
		t.Errorf("Expected valid urs, got: %v", err)
	}
}

func TestVerifyUrs_InvalidSig(t *testing.T) {
	pub, _, _ := ed25519.GenerateKey(nil)
	_, wrongPriv, _ := ed25519.GenerateKey(nil)
	vuk := Sqrl64.EncodeToString(pub)

	cr := &CliRequest{
		ClientEncoded: "client-data",
		Server:        "server-data",
	}
	signingString := cr.SigningString()
	ursSig := ed25519.Sign(wrongPriv, signingString) // Wrong key

	cr.Urs = Sqrl64.EncodeToString(ursSig)

	err := cr.VerifyUrs(vuk)
	if err == nil {
		t.Error("Expected urs verification to fail with wrong key")
	}
}

func TestVerifyUrs_EmptyVuk(t *testing.T) {
	cr := &CliRequest{
		ClientEncoded: "client",
		Server:        "server",
		Urs:           "some-urs",
	}

	err := cr.VerifyUrs("")
	if err == nil {
		t.Error("Expected error for empty vuk")
	}
}

func TestVerifyUrs_EmptyUrs(t *testing.T) {
	cr := &CliRequest{
		ClientEncoded: "client",
		Server:        "server",
		Urs:           "",
	}

	err := cr.VerifyUrs("some-vuk")
	if err == nil {
		t.Error("Expected error for empty urs")
	}
}

func TestVerifyUrs_InvalidVuk(t *testing.T) {
	cr := &CliRequest{
		ClientEncoded: "client",
		Server:        "server",
		Urs:           Sqrl64.EncodeToString([]byte("fake-sig")),
	}

	// Valid base64 but wrong size
	err := cr.VerifyUrs(Sqrl64.EncodeToString([]byte("short")))
	if err == nil {
		t.Error("Expected error for invalid vuk size")
	}
}

func TestPublicKey_Valid(t *testing.T) {
	pub, _, _ := ed25519.GenerateKey(nil)
	idk := Sqrl64.EncodeToString(pub)

	cb := &ClientBody{Idk: idk}
	key, err := cb.PublicKey()
	if err != nil {
		t.Fatalf("PublicKey failed: %v", err)
	}
	if len(key) != ed25519.PublicKeySize {
		t.Errorf("Expected %d bytes, got %d", ed25519.PublicKeySize, len(key))
	}
}

func TestPublicKey_InvalidSize(t *testing.T) {
	cb := &ClientBody{Idk: Sqrl64.EncodeToString([]byte("short"))}
	_, err := cb.PublicKey()
	if err == nil {
		t.Error("Expected error for invalid key size")
	}
}

func TestPublicKey_InvalidEncoding(t *testing.T) {
	cb := &ClientBody{Idk: "!!!invalid!!!"}
	_, err := cb.PublicKey()
	if err == nil {
		t.Error("Expected error for invalid encoding")
	}
}

func TestPidkPublicKey_Valid(t *testing.T) {
	pub, _, _ := ed25519.GenerateKey(nil)
	pidk := Sqrl64.EncodeToString(pub)

	cb := &ClientBody{Pidk: pidk}
	key, err := cb.PidkPublicKey()
	if err != nil {
		t.Fatalf("PidkPublicKey failed: %v", err)
	}
	if len(key) != ed25519.PublicKeySize {
		t.Errorf("Expected %d bytes, got %d", ed25519.PublicKeySize, len(key))
	}
}

func TestPidkPublicKey_InvalidSize(t *testing.T) {
	cb := &ClientBody{Pidk: Sqrl64.EncodeToString([]byte("short"))}
	_, err := cb.PidkPublicKey()
	if err == nil {
		t.Error("Expected error for invalid pidk size")
	}
}

func TestSigningString_BuildsFromClientEncode(t *testing.T) {
	cb := &ClientBody{
		Version: []int{1},
		Cmd:     "query",
		Idk:     "testidk",
	}
	cr := &CliRequest{
		Client: cb,
		Server: "server-data",
	}

	result := cr.SigningString()
	// Should have set ClientEncoded
	if cr.ClientEncoded == "" {
		t.Error("Expected ClientEncoded to be populated")
	}
	expected := cr.ClientEncoded + "server-data"
	if string(result) != expected {
		t.Errorf("Expected %s, got %s", expected, string(result))
	}
}

func BenchmarkVerifySignature(b *testing.B) {
	pub, priv, _ := ed25519.GenerateKey(nil)
	idk := Sqrl64.EncodeToString(pub)
	cb := &ClientBody{Version: []int{1}, Cmd: "query", Idk: idk}
	clientEncoded := string(cb.Encode())
	server := "server-data"
	sig := ed25519.Sign(priv, []byte(clientEncoded+server))
	ids := Sqrl64.EncodeToString(sig)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cr := &CliRequest{
			Client:        cb,
			ClientEncoded: clientEncoded,
			Server:        server,
			Ids:           ids,
		}
		_ = cr.VerifySignature()
	}
}
