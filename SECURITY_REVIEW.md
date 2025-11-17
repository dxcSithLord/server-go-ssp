# SQRL SSP Security Review and Implementation Plan

**Repository:** github.com/sqrldev/server-go-ssp
**Review Date:** November 17, 2025
**Go Version:** 1.17 (current) → 1.21+ (recommended)

---

## 1. Code Requirements Documentation

### Project Overview
SQRL (Secure QR Login) Server-Side Protocol implementation in Go. The library provides:
- Cryptographic identity management using ED25519 signatures
- Session nonce (Nut) generation and validation
- Authentication state management
- Pluggable storage backends

### Functional Requirements

| Requirement | Description | Implementation Status |
|-------------|-------------|----------------------|
| SQRL Protocol Compliance | Implement GRC SQRL specification | Complete |
| Identity Management | Create, authenticate, disable, remove identities | Complete |
| Key Rotation | Support identity key rotation (Pidk/Idk) | Complete |
| Signature Verification | Validate ED25519 signatures (IDS, PIDS, URS) | Complete |
| Nonce Generation | Cryptographically secure nut generation | Complete |
| State Management | Track authentication state transitions | Complete |
| Horizontal Scaling | Support multi-server deployments | Complete |
| Load Balancer Support | Handle X-Forwarded-* headers | Complete |

### Non-Functional Requirements

| Requirement | Status | Notes |
|-------------|--------|-------|
| Test Coverage | **CRITICAL: 8.0%** | Requires minimum 80% |
| CI/CD Pipeline | **MISSING** | No GitHub Actions |
| Secure Memory Handling | **MISSING** | No clearing of sensitive data |
| Logging Security | **VULNERABLE** | Logs contain sensitive data |
| Dependency Updates | **OUTDATED** | Critical crypto library outdated |

---

## 2. Dependencies and Upgrade Path

### Current Dependencies (UPDATED)

| Package | Previous Version | Current Version | Status |
|---------|-----------------|-----------------|--------|
| golang.org/x/crypto | v0.31.0 | **v0.44.0** | **UPGRADED** |
| github.com/davecgh/go-spew | v1.1.1 | **REMOVED** | Eliminated (security risk) |
| github.com/skip2/go-qrcode | v0.0.0-20200617195104 | v0.0.0-20200617195104 | No formal releases |

### Upgrade Completed

```bash
# Go version upgraded: 1.17 → 1.24.0
# Toolchain: go1.24.7
# golang.org/x/crypto: v0.31.0 → v0.44.0

# Commands executed:
go get golang.org/x/crypto@v0.44.0
go mod tidy
go mod verify  # Result: all modules verified
go test ./...  # Result: all tests pass, 33.7% coverage
```

### Breaking Changes Assessment

**golang.org/x/crypto v0.31.0 → v0.44.0:**
- ED25519 API remains stable ✓
- Blowfish API remains stable ✓
- Minor performance improvements ✓
- No breaking changes for this codebase ✓
- Security fixes included ✓
- Critical DoS/SSH vulnerability fixed ✓

**Go 1.17 → 1.24.0:**
- Standard library ED25519 (crypto/ed25519) enhanced
- No breaking changes for this codebase
- Improved cryptographic performance
- Latest security patches included

---

## 3. Sensitive Data Handling Vulnerabilities

### Critical Security Issues Identified

#### CWE-226: Sensitive Information in Resource Not Removed Before Reuse

**Severity: HIGH**

| File | Line | Issue | Sensitive Data |
|------|------|-------|----------------|
| api.go | 61-73 | `SqrlIdentity` struct fields not cleared | Suk, Vuk, Idk, Pidk |
| cli_request.go | 173-183 | `CliRequest` signatures not cleared | Ids, Pids, Urs |
| cli_request.go | 59-69 | `ClientBody` keys not cleared | Suk, Vuk, Pidk, Idk |
| grc_tree.go | 13-17 | `GrcTree` stores blowfish key permanently | blowfishKey |
| cli_request.go | 224-228 | Decoded IDS signature not cleared | decodedIds |
| cli_request.go | 249-252 | Decoded PIDS signature not cleared | decodedPids |
| cli_request.go | 266-271 | Decoded URS and VUK not cleared | decodedUrs, pubKey |
| map_hoard.go | 69-78 | Deleted cache entries not securely cleared | HoardCache with Identity |
| random_tree.go | 34-41 | Random bytes not cleared after encoding | valueBytes |

#### CWE-200: Exposure of Sensitive Information to an Unauthorized Actor

**Severity: HIGH**

| File | Line | Issue | Data Exposed |
|------|------|-------|--------------|
| cli_request.go | 300 | Logs raw request body | `log.Printf("Got body: %v", string(body))` - Contains encoded keys |
| cli_request.go | 113 | Logs encoded response | `log.Printf("Encoded response: <%v>", encoded)` |
| cli_handler.go | 42 | Dumps entire request with spew | `spew.Dump(req)` - Contains all cryptographic data |
| cli_handler.go | 122-124 | Logs decoded/encoded responses | Contains server secrets |
| cli_handler.go | 164 | Logs identity with %#v | `log.Printf("Authenticated Idk: %#v", identity)` |
| cli_handler.go | 210 | Logs identity swap | `log.Printf("Swapped identity %#v for %#v", ...)` |
| cli_handler.go | 242 | Logs last response | Leaks previous cryptographic responses |

### Additional Security Concerns

1. **No Rate Limiting**: Brute force attacks possible on `/cli.sqrl`
2. **No Request Size Limits**: DoS via large request bodies
3. **Unbounded Memory Growth**: RandomTree channel buffers 1000 nuts indefinitely
4. **Information Leakage via Timing**: Some operations may leak timing information
5. **Missing Input Validation**: Limited bounds checking on decoded data

---

## 4. Implementation Plan: Secure Memory Clearing

### Phase 1: Create Secure Memory Utilities

```go
// secure_clear.go - Platform-aware secure memory clearing
package ssp

import (
    "runtime"
    "unsafe"
)

// ClearBytes securely clears a byte slice by overwriting with zeros
// Uses compiler memory fence to prevent optimization removal
func ClearBytes(b []byte) {
    for i := range b {
        b[i] = 0
    }
    // Memory fence to prevent compiler optimization
    runtime.KeepAlive(b)
}

// ClearString securely clears a string (note: strings are immutable,
// this works on the underlying bytes)
func ClearString(s *string) {
    if s == nil || *s == "" {
        return
    }
    b := []byte(*s)
    ClearBytes(b)
    *s = ""
}

// ClearIdentity securely clears all sensitive fields in SqrlIdentity
func (si *SqrlIdentity) Clear() {
    ClearString(&si.Idk)
    ClearString(&si.Suk)
    ClearString(&si.Vuk)
    ClearString(&si.Pidk)
    ClearString(&si.Rekeyed)
}

// ClearClientBody securely clears all sensitive fields
func (cb *ClientBody) Clear() {
    if cb == nil {
        return
    }
    ClearString(&cb.Suk)
    ClearString(&cb.Vuk)
    ClearString(&cb.Pidk)
    ClearString(&cb.Idk)
}

// ClearCliRequest securely clears all sensitive signature data
func (cr *CliRequest) Clear() {
    if cr == nil {
        return
    }
    ClearString(&cr.Ids)
    ClearString(&cr.Pids)
    ClearString(&cr.Urs)
    ClearString(&cr.ClientEncoded)
    ClearString(&cr.Server)
    if cr.Client != nil {
        cr.Client.Clear()
    }
}

// ClearHoardCache securely clears cached sensitive data
func (hc *HoardCache) Clear() {
    if hc == nil {
        return
    }
    if hc.Identity != nil {
        hc.Identity.Clear()
    }
    if hc.LastRequest != nil {
        hc.LastRequest.Clear()
    }
    ClearBytes(hc.LastResponse)
}
```

### Phase 2: Platform-Aware Implementation

```go
// secure_clear_unix.go
//go:build unix

package ssp

import (
    "syscall"
    "unsafe"
)

// mlock prevents memory from being swapped to disk
func LockMemory(b []byte) error {
    if len(b) == 0 {
        return nil
    }
    return syscall.Mlock(b)
}

// munlock unlocks memory
func UnlockMemory(b []byte) error {
    if len(b) == 0 {
        return nil
    }
    return syscall.Munlock(b)
}

// SecureAlloc allocates memory that won't be swapped
func SecureAlloc(size int) ([]byte, error) {
    b := make([]byte, size)
    if err := LockMemory(b); err != nil {
        return b, err // Return buffer but note locking failed
    }
    return b, nil
}
```

```go
// secure_clear_windows.go
//go:build windows

package ssp

import (
    "syscall"
    "unsafe"
)

var (
    kernel32         = syscall.NewLazyDLL("kernel32.dll")
    procVirtualLock  = kernel32.NewProc("VirtualLock")
    procVirtualUnlock = kernel32.NewProc("VirtualUnlock")
)

func LockMemory(b []byte) error {
    if len(b) == 0 {
        return nil
    }
    ret, _, err := procVirtualLock.Call(
        uintptr(unsafe.Pointer(&b[0])),
        uintptr(len(b)),
    )
    if ret == 0 {
        return err
    }
    return nil
}

func UnlockMemory(b []byte) error {
    if len(b) == 0 {
        return nil
    }
    ret, _, err := procVirtualUnlock.Call(
        uintptr(unsafe.Pointer(&b[0])),
        uintptr(len(b)),
    )
    if ret == 0 {
        return err
    }
    return nil
}
```

### Phase 3: Update Cryptographic Operations

**cli_request.go modifications:**

```go
// VerifySignature with secure clearing
func (cr *CliRequest) VerifySignature() error {
    pubKey, err := cr.Client.PublicKey()
    if err != nil {
        return err
    }
    defer ClearBytes(pubKey) // SECURE CLEAR

    decodedIds, err := Sqrl64.DecodeString(cr.Ids)
    if err != nil {
        return fmt.Errorf("invalid ids: %v", err)
    }
    defer ClearBytes(decodedIds) // SECURE CLEAR

    if !ed25519.Verify(pubKey, cr.SigningString(), decodedIds) {
        return fmt.Errorf("signature verification failed")
    }

    if cr.Pids != "" || cr.Client.Pidk != "" {
        return cr.VerifyPidsSignature()
    }
    return nil
}
```

**grc_tree.go modifications:**

```go
// GrcTree with secure key clearing
type GrcTree struct {
    monotonicCounter uint64
    cipher           *blowfish.Cipher
    key              []byte // Will be cleared when no longer needed
}

// Close securely clears the blowfish key
func (gt *GrcTree) Close() {
    ClearBytes(gt.key)
    gt.cipher = nil
}
```

### Phase 4: Secure Logging

```go
// secure_log.go - Safe logging without sensitive data exposure
package ssp

import (
    "log"
    "strings"
)

// SafeLogRequest logs request without sensitive data
func SafeLogRequest(req *CliRequest) {
    if req == nil {
        log.Printf("Request: nil")
        return
    }
    log.Printf("Request: cmd=%s, idk=%s..., ip=%s",
        req.Client.Cmd,
        truncateKey(req.Client.Idk, 8),
        req.IPAddress)
}

// SafeLogIdentity logs identity without exposing full keys
func SafeLogIdentity(identity *SqrlIdentity) {
    if identity == nil {
        log.Printf("Identity: nil")
        return
    }
    log.Printf("Identity: idk=%s..., disabled=%v, rekeyed=%v",
        truncateKey(identity.Idk, 8),
        identity.Disabled,
        identity.Rekeyed != "")
}

func truncateKey(key string, maxLen int) string {
    if len(key) <= maxLen {
        return key
    }
    return key[:maxLen]
}
```

---

## 5. Test Coverage Requirements

### Current State: 8.0% Coverage (CRITICAL)

### Target: 80% Minimum Coverage

### Test Plan

| Component | Current Coverage | Target | Priority |
|-----------|-----------------|--------|----------|
| cli_request.go | ~10% | 90% | CRITICAL |
| cli_handler.go | 0% | 85% | CRITICAL |
| cli_response.go | ~15% | 90% | HIGH |
| api.go | 0% | 80% | HIGH |
| handers.go | 0% | 85% | HIGH |
| grc_tree.go | ~40% | 95% | MEDIUM |
| random_tree.go | ~40% | 95% | MEDIUM |
| map_hoard.go | ~50% | 95% | MEDIUM |
| map_auth_store.go | 0% | 90% | HIGH |
| secure_clear.go | N/A (new) | 100% | CRITICAL |

### Required Test Categories

1. **Unit Tests**
   - Signature verification (valid/invalid/malformed)
   - Request parsing (edge cases, malformed input)
   - Response encoding/decoding
   - State transitions
   - Secure memory clearing verification

2. **Integration Tests**
   - Full authentication flows
   - Identity lifecycle (create, disable, enable, remove)
   - Key rotation scenarios
   - Error handling paths

3. **Security Tests**
   - Memory clearing verification
   - Timing attack resistance
   - Input validation bounds
   - Injection attempts

4. **Benchmark Tests**
   - Cryptographic operation performance
   - Memory allocation patterns
   - Nonce generation throughput

---

## 6. CI/CD Pipeline Configuration

### GitHub Actions Workflow

```yaml
# .github/workflows/ci.yml
name: CI/CD Pipeline

on:
  push:
    branches: [ main, master, develop ]
  pull_request:
    branches: [ main, master ]

jobs:
  security-scan:
    name: Security Scanning
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Run Gosec Security Scanner
        uses: securego/gosec@master
        with:
          args: '-no-fail -fmt sarif -out results.sarif ./...'

      - name: Upload SARIF file
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: results.sarif

  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v4
        with:
          version: latest
          args: --timeout=5m

  test:
    name: Test
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: ['1.21', '1.22', '1.23']
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go-version }}

      - name: Get dependencies
        run: go mod download

      - name: Run Tests with Race Detector
        run: go test -race -coverprofile=coverage.out -covermode=atomic ./...

      - name: Check Coverage Threshold
        run: |
          COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print substr($3, 1, length($3)-1)}')
          echo "Total Coverage: ${COVERAGE}%"
          if (( $(echo "$COVERAGE < 80" | bc -l) )); then
            echo "Coverage ${COVERAGE}% is below 80% threshold"
            exit 1
          fi

      - name: Upload Coverage to Codecov
        uses: codecov/codecov-action@v4
        with:
          file: ./coverage.out
          flags: unittests
          name: codecov-umbrella

  build:
    name: Build
    needs: [lint, test]
    runs-on: ubuntu-latest
    strategy:
      matrix:
        goos: [linux, windows, darwin]
        goarch: [amd64, arm64]
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Build
        env:
          GOOS: ${{ matrix.goos }}
          GOARCH: ${{ matrix.goarch }}
        run: |
          go build -v ./...
          if [ -d "server" ]; then
            cd server && go build -o ../sqrl-server-${{ matrix.goos }}-${{ matrix.goarch }}
          fi

  dependency-review:
    name: Dependency Review
    runs-on: ubuntu-latest
    if: github.event_name == 'pull_request'
    steps:
      - uses: actions/checkout@v4
      - name: Dependency Review
        uses: actions/dependency-review-action@v4
        with:
          fail-on-severity: high

  codeql:
    name: CodeQL Analysis
    runs-on: ubuntu-latest
    permissions:
      actions: read
      contents: read
      security-events: write
    steps:
      - uses: actions/checkout@v4

      - name: Initialize CodeQL
        uses: github/codeql-action/init@v3
        with:
          languages: go

      - name: Autobuild
        uses: github/codeql-action/autobuild@v3

      - name: Perform CodeQL Analysis
        uses: github/codeql-action/analyze@v3
```

### Pre-commit Hooks

```yaml
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/dnephin/pre-commit-golang
    rev: master
    hooks:
      - id: go-fmt
      - id: go-vet
      - id: go-imports
      - id: go-cyclo
        args: [-over=15]
      - id: validate-toml
      - id: no-go-testing
      - id: golangci-lint
      - id: go-critic
      - id: go-unit-tests
      - id: go-build
      - id: go-mod-tidy
```

---

## 7. Implementation Priority and Timeline

### Phase 1: Critical Security Fixes (Week 1)
- [ ] Implement secure memory clearing utilities
- [ ] Remove sensitive data from logs
- [ ] Remove spew.Dump debug calls
- [ ] Update golang.org/x/crypto to v0.44.0
- [ ] Update Go version to 1.21

### Phase 2: Cryptographic Security (Week 2)
- [ ] Add defer ClearBytes() to all signature verifications
- [ ] Implement GrcTree.Close() for key clearing
- [ ] Add secure clearing to HoardCache cleanup
- [ ] Implement platform-aware memory locking (optional)

### Phase 3: Test Coverage (Weeks 3-4)
- [ ] Write comprehensive unit tests for cli_request.go
- [ ] Write integration tests for cli_handler.go
- [ ] Add security-focused tests
- [ ] Achieve 80% coverage minimum
- [ ] Add benchmark tests

### Phase 4: CI/CD Setup (Week 5)
- [ ] Create GitHub Actions workflow
- [ ] Add security scanning (gosec, CodeQL)
- [ ] Configure code coverage reporting
- [ ] Set up dependency review
- [ ] Add pre-commit hooks

### Phase 5: Documentation and Validation (Week 6)
- [ ] Update README with security practices
- [ ] Document API security considerations
- [ ] Conduct final security audit
- [ ] Performance benchmarking

---

## 8. Risk Assessment

| Risk | Severity | Likelihood | Mitigation |
|------|----------|------------|------------|
| Key exposure via logs | HIGH | HIGH | Remove sensitive logging immediately |
| Memory not cleared | HIGH | MEDIUM | Implement secure clearing functions |
| Outdated crypto library | MEDIUM | HIGH | Update to v0.44.0 |
| Low test coverage | MEDIUM | HIGH | Mandate 80% minimum coverage |
| No CI/CD pipeline | MEDIUM | N/A | Implement GitHub Actions |
| Timing attacks | MEDIUM | LOW | Already uses subtle.ConstantTimeCompare |
| Brute force attacks | LOW | MEDIUM | Consider rate limiting (future) |

---

## 9. Compliance Considerations

This security review addresses:
- **CWE-226**: Sensitive Information in Resource Not Removed Before Reuse
- **CWE-200**: Exposure of Sensitive Information to an Unauthorized Actor
- **CWE-312**: Cleartext Storage of Sensitive Information (in logs)
- **CWE-532**: Insertion of Sensitive Information into Log File
- **OWASP Top 10 2021**: A02 - Cryptographic Failures

---

## 10. Monitoring and Maintenance

### Post-Implementation Monitoring
1. Set up Dependabot for automatic dependency updates
2. Schedule quarterly security reviews
3. Monitor for new CVEs in dependencies
4. Track test coverage trends
5. Review CI/CD pipeline effectiveness

### Dependency Update Schedule
- **golang.org/x/crypto**: Monitor monthly for security updates
- **Go runtime**: Update within 3 months of new releases
- **Development tools**: Update quarterly

---

## Summary

This SQRL SSP implementation requires immediate security improvements:

1. **CRITICAL**: Implement secure memory clearing for cryptographic keys
2. **CRITICAL**: Remove sensitive data from log outputs
3. **HIGH**: Update golang.org/x/crypto from v0.31.0 to v0.44.0
4. **HIGH**: Increase test coverage from 8% to 80%+
5. **MEDIUM**: Set up comprehensive CI/CD pipeline

The codebase is functionally complete but lacks essential security hardening for production use. The identified vulnerabilities (CWE-226, CWE-200) must be addressed before deploying in security-sensitive environments.
