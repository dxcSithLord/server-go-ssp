# Staged Dependency Upgrade Plan
**Project:** github.com/dxcSithLord/server-go-ssp
**Current State:** Go 1.25.4, Minimal Dependencies
**Created:** November 18, 2025
**Status:** Planning Phase

---

## Executive Summary

This document outlines a comprehensive, staged approach to upgrading dependencies and integrating new components into the SQRL SSP codebase. The plan is structured to allow incremental testing and validation at each stage, ensuring system stability and functional test continuity.

### Current State
- **Go Version:** 1.25.4 (Latest stable) ✅
- **External Dependencies:** 1 (github.com/skip2/go-qrcode)
- **Dependency Status:** Unmaintained QR library (5+ years old)
- **Deployment:** Single-server with in-memory storage
- **Test Coverage:** 29.4% (Target: 80%)

### Proposed Enhancements
1. **Replace unmaintained QR code library** (Security & Maintenance)
2. **Add distributed storage with etcd** (Horizontal Scaling)
3. **Optional: Integrate with Gogs** (Git hosting)
4. **Optional: Integrate with Authentik** (OAuth2/OIDC provider)

---

## Upgrade Philosophy

### Staged Approach Benefits
- ✅ **Incremental Risk:** Each stage is tested independently
- ✅ **Rollback Safety:** Can revert any stage without affecting others
- ✅ **Test Validation:** Functional tests pass before next stage
- ✅ **Branch Isolation:** Each stage has dedicated git branch
- ✅ **Production Ready:** Can deploy after any successful stage

### Progression Criteria
Each stage must meet these criteria before advancing:
1. ✅ All unit tests pass (go test ./...)
2. ✅ Coverage maintained or improved (≥29.4%, targeting 80%)
3. ✅ No new security vulnerabilities (gosec, CodeQL)
4. ✅ Functional API tests pass
5. ✅ No performance regression
6. ✅ Code review approved
7. ✅ Documentation updated

---

## Stage 0: Baseline & Preparation

**Branch:** `main` (current state)
**Duration:** Complete
**Status:** ✅ Complete

### Accomplished
- [x] Go 1.25.4 upgrade complete
- [x] Blowfish → AES migration complete
- [x] golang.org/x/crypto dependency removed
- [x] Security review completed
- [x] CI/CD pipeline established (7 jobs)
- [x] Test coverage improved (8% → 29.4%)

### Current Dependency Inventory

```bash
# Dependency List (go list -m all)
github.com/dxcSithLord/server-go-ssp
github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e
```

**Security Assessment:**
- Go 1.25.4: ✅ Latest stable, all CVEs patched
- go-qrcode: ⚠️ Unmaintained (last update June 2020)

### Baseline Metrics

| Metric | Current Value | Target |
|--------|--------------|--------|
| Test Coverage | 29.4% | 80% |
| Security Vulnerabilities | 0 | 0 |
| Build Time | ~15s | <30s |
| Binary Size (linux/amd64) | ~12MB | <20MB |
| QPS (cli.sqrl) | ~500/s | ≥500/s |

---

## Stage 1: QR Code Library Replacement

**Branch:** `claude/dependency-upgrade-stage1-qrcode`
**Duration:** 1 week
**Priority:** 🔴 **HIGH** (Security & Maintenance)
**Dependencies:** None

### Objective
Replace unmaintained `github.com/skip2/go-qrcode` with actively maintained `github.com/yeqown/go-qrcode` v2.3.1

### Rationale
- Current library unmaintained for 5+ years
- No security updates or bug fixes
- SECURITY_REVIEW.md flags potential historical CVEs
- Modern alternative provides better features and maintenance

### Dependency Changes

```diff
# go.mod
-require github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e
+require github.com/yeqown/go-qrcode/v2 v2.3.1
+require github.com/yeqown/go-qrcode/writer/standard v1.1.0
```

### Breaking Changes
**API Differences:**
```go
// OLD (skip2/go-qrcode)
png, err := qrcode.Encode(value, qrcode.Medium, -5)

// NEW (yeqown/go-qrcode)
qrc, err := qrcode.New(value)
w := standard.NewWithWriter(buf, standard.WithQRWidth(21))
err = qrc.Save(w)
```

**Impact:** Low - isolated to `/png.sqrl` endpoint

### Files Modified
- `/home/user/server-go-ssp/go.mod` - Update dependency
- `/home/user/server-go-ssp/go.sum` - Update checksums
- `/home/user/server-go-ssp/handers.go:121` - Update QR generation code
- `/home/user/server-go-ssp/SECURITY_REVIEW.md` - Update dependency notes

### Implementation Steps

#### 1.1 Research & Validation (Day 1)
```bash
# Create feature branch
git checkout -b claude/dependency-upgrade-stage1-qrcode

# Test new library in isolation
go get github.com/yeqown/go-qrcode/v2@v2.3.1
go get github.com/yeqown/go-qrcode/writer/standard@v1.1.0

# Create test program
cat > test_qrcode.go <<'EOF'
package main

import (
    "bytes"
    "fmt"
    "os"

    "github.com/yeqown/go-qrcode/v2"
    "github.com/yeqown/go-qrcode/writer/standard"
)

func main() {
    testURL := "sqrl://example.com/cli.sqrl?nut=abc123"

    qrc, err := qrcode.New(testURL)
    if err != nil {
        panic(err)
    }

    var buf bytes.Buffer
    w := standard.NewWithWriter(&buf, standard.WithQRWidth(21))

    if err = qrc.Save(w); err != nil {
        panic(err)
    }

    fmt.Printf("Generated QR code: %d bytes\n", buf.Len())
    os.WriteFile("test_qr.png", buf.Bytes(), 0644)
}
EOF

go run test_qrcode.go
# Verify test_qr.png is valid
```

#### 1.2 Code Migration (Days 2-3)
Update `handers.go`:
```go
// File: handers.go
package ssp

import (
    "bytes"
    "fmt"
    "net/http"

    "github.com/yeqown/go-qrcode/v2"
    "github.com/yeqown/go-qrcode/writer/standard"
)

func (api *SqrlSspAPI) Png(w http.ResponseWriter, r *http.Request) {
    // ... existing nut generation code ...

    // Generate SQRL URL
    sqrlURL := fmt.Sprintf("%s/cli.sqrl?nut=%s", api.SiteURL(r), nutValue)

    // NEW: Generate QR code with yeqown/go-qrcode
    qrc, err := qrcode.New(sqrlURL)
    if err != nil {
        SafeLog(fmt.Sprintf("Failed to create QR code: %v", err))
        http.Error(w, "Internal error", http.StatusInternalServerError)
        return
    }

    var buf bytes.Buffer
    qrWriter := standard.NewWithWriter(&buf,
        standard.WithQRWidth(21),        // ~256px default
        standard.WithBorderWidth(4),     // Quiet zone
    )

    if err = qrc.Save(qrWriter); err != nil {
        SafeLog(fmt.Sprintf("Failed to encode QR code: %v", err))
        http.Error(w, "Internal error", http.StatusInternalServerError)
        return
    }

    // Set headers
    w.Header().Set("Content-Type", "image/png")
    w.Header().Set("Content-Length", fmt.Sprintf("%d", buf.Len()))

    if nutIsNew {
        w.Header().Set("Sqrl-Nut", nutValue.String())
        w.Header().Set("Sqrl-Pag", pagValue.String())
        w.Header().Set("Sqrl-Exp", fmt.Sprintf("%d", int(api.NutExpiration.Seconds())))
    }

    w.Write(buf.Bytes())
}
```

#### 1.3 Testing (Days 4-5)

**Unit Tests:**
```bash
# Run all existing tests
go test -v -race -coverprofile=coverage.out ./...

# Check coverage (must be ≥29.4%)
go tool cover -func=coverage.out | grep total

# Specific test for PNG endpoint
curl -v http://localhost:8000/png.sqrl -o test.png
file test.png  # Should show: PNG image data
```

**Integration Tests:**
```bash
# Start test server
go run server/main.go -p 8080 &
SERVER_PID=$!

# Test QR generation
curl http://localhost:8080/png.sqrl -o stage1_qr.png
curl http://localhost:8080/png.sqrl?nut=$(curl -s http://localhost:8080/nut.sqrl | grep -oP 'nut=\K[^&]+') -o stage1_qr_with_nut.png

# Verify both are valid PNG files
file stage1_qr.png stage1_qr_with_nut.png

# Test with SQRL client
# Manual verification: Scan with SQRL app

kill $SERVER_PID
```

**Performance Benchmark:**
```bash
# Benchmark QR generation performance
cat > png_benchmark_test.go <<'EOF'
package ssp

import (
    "bytes"
    "testing"

    "github.com/yeqown/go-qrcode/v2"
    "github.com/yeqown/go-qrcode/writer/standard"
)

func BenchmarkQRCodeGeneration(b *testing.B) {
    testURL := "sqrl://example.com/cli.sqrl?nut=abc123def456"

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        qrc, _ := qrcode.New(testURL)
        var buf bytes.Buffer
        w := standard.NewWithWriter(&buf, standard.WithQRWidth(21))
        qrc.Save(w)
    }
}
EOF

go test -bench=BenchmarkQRCodeGeneration -benchmem
# Compare with baseline (if available)
```

#### 1.4 Security Validation (Day 6)
```bash
# Run security scanners
gosec ./...

# Check for new vulnerabilities
govulncheck ./...

# Run CodeQL analysis
# (via GitHub Actions or locally)

# Verify no new issues introduced
```

#### 1.5 Documentation & PR (Day 7)
```bash
# Update documentation
vim SECURITY_REVIEW.md  # Update dependency table
vim README.md           # Update dependency notes

# Commit changes
git add .
git commit -m "Replace skip2/go-qrcode with yeqown/go-qrcode v2.3.1

- Migrate from unmaintained skip2/go-qrcode (last update 2020)
- Update to actively maintained yeqown/go-qrcode v2.3.1
- No API changes for end users (internal implementation only)
- All tests pass with maintained coverage
- No new security vulnerabilities

Fixes: Security maintenance concern from SECURITY_REVIEW.md
Closes: #[issue-number]"

# Push and create PR
git push -u origin claude/dependency-upgrade-stage1-qrcode
gh pr create --title "Stage 1: Replace QR code library" --body "..."
```

### Success Criteria
- [ ] All tests pass (`go test ./...`)
- [ ] Coverage ≥29.4% (maintained or improved)
- [ ] No security vulnerabilities
- [ ] QR codes scannable by SQRL clients
- [ ] Performance maintained (≥500 QPS)
- [ ] Documentation updated
- [ ] Code review approved

### Rollback Plan
```bash
# If issues discovered, revert PR
git checkout main
git branch -D claude/dependency-upgrade-stage1-qrcode

# OR revert specific commit
git revert <commit-hash>
```

### Risk Assessment
- **Risk Level:** 🟢 **LOW**
- **Impact Area:** Limited to `/png.sqrl` endpoint only
- **Failure Mode:** QR codes fail to generate (HTTP 500)
- **User Impact:** Users cannot scan QR codes (can still use manual entry)
- **Mitigation:** Comprehensive testing before merge

---

## Stage 2: Test Coverage Enhancement

**Branch:** `claude/dependency-upgrade-stage2-tests`
**Duration:** 2-3 weeks
**Priority:** 🔴 **HIGH** (Security Requirement)
**Dependencies:** Stage 1 complete

### Objective
Increase test coverage from 29.4% to 80% minimum to meet CI/CD requirements

### Rationale
- Required for production deployment
- Enforced by CI/CD pipeline
- Improves code quality and maintainability
- Enables safer refactoring in future stages

### Target Coverage by Component

| Component | Current | Target | Priority |
|-----------|---------|--------|----------|
| cli_request.go | ~10% | 90% | 🔴 CRITICAL |
| cli_handler.go | 0% | 85% | 🔴 CRITICAL |
| cli_response.go | ~15% | 90% | 🔴 HIGH |
| api.go | 0% | 80% | 🔴 HIGH |
| handers.go | 0% | 85% | 🔴 HIGH |
| grc_tree.go | ~40% | 95% | 🟡 MEDIUM |
| random_tree.go | ~40% | 95% | 🟡 MEDIUM |
| map_hoard.go | ~50% | 95% | 🟡 MEDIUM |
| map_auth_store.go | 0% | 90% | 🔴 HIGH |
| secure_clear.go | NEW | 100% | 🔴 CRITICAL |
| secure_log.go | Partial | 95% | 🔴 HIGH |

### Implementation Plan

#### Week 1: Critical Path Coverage (cli_handler, cli_request)
Create comprehensive tests for authentication flows

#### Week 2: API & Response Coverage
Test all endpoints and response encoding

#### Week 3: Storage & Utilities
Complete coverage for storage backends and utilities

### Success Criteria
- [ ] Overall coverage ≥80%
- [ ] All critical paths tested
- [ ] Edge cases covered
- [ ] Security tests added
- [ ] CI/CD enforces threshold

### Dependencies
None - can proceed after Stage 1

---

## Stage 3: Distributed Storage (etcd Integration)

**Branch:** `claude/dependency-upgrade-stage3-etcd`
**Duration:** 2-3 weeks
**Priority:** 🟡 **MEDIUM** (Production Scaling)
**Dependencies:** Stage 2 complete (high test coverage required)

### Objective
Enable horizontal scaling by implementing distributed storage with etcd v3.6.6

### Rationale
- Current `MapHoard`/`MapAuthStore` not production-ready (documented limitation)
- Horizontal scaling requires distributed state
- etcd provides strong consistency for authentication state
- Enables multi-datacenter deployments

### Dependency Changes

```diff
# go.mod
+require go.etcd.io/etcd/client/v3 v3.6.6
```

### Architecture Changes

**Before (Single Server):**
```
┌─────────────────┐
│  SQRL SSP       │
│  ┌────────┐     │
│  │ Hoard  │     │  In-memory
│  └────────┘     │  (map)
│  ┌────────┐     │
│  │AuthStore│    │  In-memory
│  └────────┘     │  (map)
└─────────────────┘
```

**After (Multi-Server with etcd):**
```
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│ SQRL SSP #1  │  │ SQRL SSP #2  │  │ SQRL SSP #3  │
│  ┌────────┐  │  │  ┌────────┐  │  │  ┌────────┐  │
│  │Etcd    │  │  │  │Etcd    │  │  │  │Etcd    │  │
│  │Hoard   │──┼──┼──│Hoard   │──┼──┼──│Hoard   │  │
│  └────────┘  │  │  └────────┘  │  │  └────────┘  │
└──────┬───────┘  └──────┬───────┘  └──────┬───────┘
       │                 │                 │
       └────────┬────────┴────────┬────────┘
                │                 │
           ┌────▼─────────────────▼────┐
           │   etcd Cluster (3 nodes)  │
           │  - Nut Storage (TTL)      │
           │  - Identity Storage       │
           │  - Configuration          │
           └──────────────────────────┘
```

### Files Created
- `/home/user/server-go-ssp/etcd_hoard.go` - Distributed nut storage
- `/home/user/server-go-ssp/etcd_auth_store.go` - Distributed identity storage
- `/home/user/server-go-ssp/etcd_hoard_test.go` - Unit tests
- `/home/user/server-go-ssp/etcd_auth_store_test.go` - Unit tests
- `/home/user/server-go-ssp/integration_test.go` - Multi-server tests

### Files Modified
- `/home/user/server-go-ssp/server/main.go` - Add etcd configuration flags
- `/home/user/server-go-ssp/api.go` - Support etcd-backed stores
- `/home/user/server-go-ssp/README.md` - Document etcd deployment
- `/home/user/server-go-ssp/go.mod` - Add etcd dependency

### Implementation Steps

#### 3.1 etcd Client Integration (Week 1)
```go
// File: etcd_hoard.go
package ssp

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdHoard struct {
    client *clientv3.Client
    prefix string
}

func NewEtcdHoard(endpoints []string) (*EtcdHoard, error) {
    client, err := clientv3.New(clientv3.Config{
        Endpoints:   endpoints,
        DialTimeout: 5 * time.Second,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to connect to etcd: %w", err)
    }

    return &EtcdHoard{
        client: client,
        prefix: "/sqrl/nuts/",
    }, nil
}

func (h *EtcdHoard) Save(nut Nut, cache *HoardCache, expiration time.Duration) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    key := h.prefix + string(nut)
    value, err := json.Marshal(cache)
    if err != nil {
        return fmt.Errorf("failed to marshal cache: %w", err)
    }

    // Create lease for automatic expiration
    lease, err := h.client.Grant(ctx, int64(expiration.Seconds()))
    if err != nil {
        return fmt.Errorf("failed to create lease: %w", err)
    }

    _, err = h.client.Put(ctx, key, string(value), clientv3.WithLease(lease.ID))
    if err != nil {
        return fmt.Errorf("failed to save to etcd: %w", err)
    }

    return nil
}

func (h *EtcdHoard) Get(nut Nut) (*HoardCache, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    key := h.prefix + string(nut)
    resp, err := h.client.Get(ctx, key)
    if err != nil {
        return nil, fmt.Errorf("failed to get from etcd: %w", err)
    }

    if len(resp.Kvs) == 0 {
        return nil, fmt.Errorf("nut not found: %s", nut)
    }

    var cache HoardCache
    if err := json.Unmarshal(resp.Kvs[0].Value, &cache); err != nil {
        return nil, fmt.Errorf("failed to unmarshal cache: %w", err)
    }

    return &cache, nil
}

func (h *EtcdHoard) GetAndDelete(nut Nut) (*HoardCache, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    key := h.prefix + string(nut)

    // Atomic get-and-delete transaction
    txn := h.client.Txn(ctx)
    txn.If(clientv3.Compare(clientv3.Version(key), ">", 0)).
        Then(clientv3.OpGet(key), clientv3.OpDelete(key))

    resp, err := txn.Commit()
    if err != nil {
        return nil, fmt.Errorf("failed to get-and-delete: %w", err)
    }

    if !resp.Succeeded {
        return nil, fmt.Errorf("nut not found: %s", nut)
    }

    getResp := resp.Responses[0].GetResponseRange()
    if len(getResp.Kvs) == 0 {
        return nil, fmt.Errorf("nut not found: %s", nut)
    }

    var cache HoardCache
    if err := json.Unmarshal(getResp.Kvs[0].Value, &cache); err != nil {
        return nil, fmt.Errorf("failed to unmarshal cache: %w", err)
    }

    // Secure clearing before return
    defer cache.Clear()

    return &cache, nil
}

func (h *EtcdHoard) Close() error {
    return h.client.Close()
}
```

#### 3.2 Server Configuration (Week 1)
Update `server/main.go`:
```go
var (
    // Existing flags
    keyFile      string
    certFile     string
    hostOverride string
    rootPath     string
    port         int

    // NEW: etcd flags
    useEtcd       bool
    etcdEndpoints string
)

func main() {
    // Existing flag parsing
    flag.StringVar(&keyFile, "key", "", "key.pem file for TLS")
    flag.StringVar(&certFile, "cert", "", "cert.pem file for TLS")
    flag.StringVar(&hostOverride, "h", "", "hostname used in creating URLs")
    flag.StringVar(&rootPath, "path", "", "path used as the root for the SQRL handlers")
    flag.IntVar(&port, "p", 8000, "port to listen on")

    // NEW: etcd configuration
    flag.BoolVar(&useEtcd, "etcd", false, "use etcd for distributed storage")
    flag.StringVar(&etcdEndpoints, "etcd-endpoints", "localhost:2379",
        "comma-separated list of etcd endpoints")

    flag.Parse()

    // Initialize storage backends
    var hoard ssp.Hoard
    var authStore ssp.AuthStore

    if useEtcd {
        endpoints := strings.Split(etcdEndpoints, ",")

        etcdHoard, err := ssp.NewEtcdHoard(endpoints)
        if err != nil {
            log.Fatalf("Failed to initialize etcd hoard: %v", err)
        }
        defer etcdHoard.Close()
        hoard = etcdHoard

        etcdAuthStore, err := ssp.NewEtcdAuthStore(endpoints)
        if err != nil {
            log.Fatalf("Failed to initialize etcd auth store: %v", err)
        }
        defer etcdAuthStore.Close()
        authStore = etcdAuthStore

        log.Printf("Using etcd storage: %v", endpoints)
    } else {
        // Default: in-memory storage (NOT for production)
        hoard = ssp.NewMapHoard()
        authStore = ssp.NewMapAuthStore()

        log.Printf("WARNING: Using in-memory storage (not suitable for production)")
    }

    // Rest of initialization...
}
```

#### 3.3 Testing (Week 2)

**Unit Tests:**
```bash
# Test with etcd in Docker
docker run -d --name etcd-test \
    -p 2379:2379 \
    -e ALLOW_NONE_AUTHENTICATION=yes \
    quay.io/coreos/etcd:v3.6.6 \
    etcd --advertise-client-urls http://0.0.0.0:2379 \
          --listen-client-urls http://0.0.0.0:2379

# Run tests
go test -v ./... -run TestEtcdHoard

# Cleanup
docker stop etcd-test && docker rm etcd-test
```

**Integration Tests:**
```bash
# Test multi-server scenario
# Terminal 1: Start etcd
docker-compose up etcd

# Terminal 2: Start SSP server 1
go run server/main.go -p 8001 -etcd -etcd-endpoints localhost:2379

# Terminal 3: Start SSP server 2
go run server/main.go -p 8002 -etcd -etcd-endpoints localhost:2379

# Terminal 4: Test shared state
# Generate nut on server 1
NUT=$(curl -s http://localhost:8001/nut.sqrl | grep -oP 'nut=\K[^&]+')

# Use nut on server 2 (should work!)
curl http://localhost:8002/png.sqrl?nut=$NUT -o test.png

# Verify PNG generated successfully
file test.png
```

**Performance Tests:**
```bash
# Benchmark etcd vs map performance
go test -bench=BenchmarkHoard -benchmem

# Load test with distributed setup
# Use apache bench or vegeta
echo "GET http://localhost:8001/nut.sqrl" | vegeta attack -duration=30s | vegeta report
```

#### 3.4 Documentation (Week 3)
Create deployment guide:
```markdown
# Deploying SQRL SSP with etcd

## etcd Cluster Setup

### Development (Single Node)
docker run -d --name sqrl-etcd \
    -p 2379:2379 \
    quay.io/coreos/etcd:v3.6.6 \
    etcd --advertise-client-urls http://0.0.0.0:2379 \
          --listen-client-urls http://0.0.0.0:2379

### Production (3-Node Cluster)
# See etcd deployment documentation
# Minimum: 3 nodes for quorum
# Recommended: 5 nodes for high availability

## SQRL SSP Configuration

./sqrl_server \
    -p 8000 \
    -etcd \
    -etcd-endpoints "etcd1:2379,etcd2:2379,etcd3:2379" \
    -cert cert.pem \
    -key key.pem

## Monitoring
# Watch etcd metrics
curl http://localhost:2379/metrics

# Check cluster health
etcdctl endpoint health --endpoints=localhost:2379
```

### Success Criteria
- [ ] EtcdHoard implements Hoard interface
- [ ] EtcdAuthStore implements AuthStore interface
- [ ] All tests pass with etcd backend
- [ ] Multi-server tests pass (shared state)
- [ ] Performance acceptable (≥400 QPS with etcd)
- [ ] Coverage maintained at ≥80%
- [ ] Deployment documentation complete

### Rollback Plan
```bash
# Revert to map-based storage
git revert <etcd-commit>

# OR continue using MapHoard (default if -etcd not specified)
./sqrl_server -p 8000  # Uses in-memory storage
```

### Risk Assessment
- **Risk Level:** 🟡 **MEDIUM**
- **Impact Area:** Storage layer (major architectural change)
- **Failure Mode:** etcd unavailable → authentication fails
- **User Impact:** Service outage if etcd cluster fails
- **Mitigation:**
  - Comprehensive testing
  - etcd cluster monitoring
  - Fallback to map storage during development
  - High availability etcd deployment (3-5 nodes)

---

## Stage 4: OAuth2/OIDC Integration (Authentik)

**Branch:** `claude/dependency-upgrade-stage4-authentik`
**Duration:** 3-4 weeks
**Priority:** 🟢 **LOW** (Optional Enhancement)
**Dependencies:** Stage 3 complete (requires distributed architecture)

### Decision Point: Is Authentik Needed?

**✅ Proceed if:**
- Need OAuth2/OIDC for third-party app integration
- Want centralized identity management across multiple services
- Require SAML, LDAP, or RADIUS support
- Have existing Authentik deployment

**❌ Skip if:**
- SQRL is sufficient for your use case
- No need for OAuth2/OIDC tokens
- Want to minimize infrastructure complexity
- Small deployment (<100 users)

### Objective
Bridge SQRL authentication to OAuth2/OIDC tokens via Authentik

### Architecture

**SQRL → Authentik Bridge:**
```
┌─────────┐    SQRL Auth     ┌──────────────┐
│  User   │─────────────────>│  SQRL SSP    │
└─────────┘                  └──────┬───────┘
                                    │ Create/Update User
                                    │ via Authentik API
                                    ▼
                             ┌──────────────┐
                             │  Authentik   │
                             │   IdP        │
                             └──────┬───────┘
                                    │ Issue OAuth2
                                    │ Access Token
                                    ▼
                             ┌──────────────┐
                             │  Application │
                             │  (consumes   │
                             │  OAuth2)     │
                             └──────────────┘
```

### Dependency Changes

**Note:** Authentik is a separate service (Python/Django), not a Go dependency.

**Integration requires:**
- Authentik API client for Go
- OAuth2 library for Go

```diff
# go.mod
+require golang.org/x/oauth2 v0.15.0
+require github.com/go-resty/resty/v2 v2.11.0  # For Authentik API calls
```

### Implementation Plan

#### Week 1: Authentik Deployment
- Deploy Authentik 2025.8+ (Docker or Kubernetes)
- Configure OAuth2/OIDC provider
- Create API token for SSP integration

#### Week 2: SQRL → Authentik Bridge
- Implement Authentik API client
- Create users in Authentik on SQRL authentication
- Generate OAuth2 authorization codes

#### Week 3: Testing
- Test full authentication flow
- Verify OAuth2 token issuance
- Test with sample OAuth2 application

#### Week 4: Documentation
- Integration guide
- Deployment instructions
- Security best practices

### Success Criteria
- [ ] Authentik deployed and configured
- [ ] SQRL authentication creates/updates Authentik users
- [ ] OAuth2 tokens issued successfully
- [ ] Sample app authenticates via OAuth2
- [ ] All tests pass
- [ ] Documentation complete

### Risk Assessment
- **Risk Level:** 🟡 **MEDIUM-HIGH**
- **Complexity:** High (two complex systems)
- **Maintenance:** Ongoing (Authentik upgrades)
- **Recommendation:** Only proceed if OAuth2/OIDC is required

---

## Stage 5: Git Hosting Integration (Gogs/Gitea)

**Branch:** `claude/dependency-upgrade-stage5-git`
**Duration:** 2-3 weeks
**Priority:** 🟢 **LOW** (Optional Enhancement)
**Dependencies:** Stage 4 complete (requires OAuth2/OIDC)

### Decision Point: Is Git Hosting Needed?

**✅ Proceed if:**
- Need self-hosted Git repository management
- Want integrated Git + SQRL authentication
- Have use case for passwordless Git access

**❌ Skip if:**
- Already have Git hosting (GitHub, GitLab, Bitbucket)
- No need for self-hosted Git
- SQRL authentication for Git not required

### Recommendation: Use Gitea Instead of Gogs

**Gitea Advantages:**
- Native OAuth2/OIDC support (Gogs lacks this)
- Active development and maintenance
- Better feature set
- Easier integration with modern auth systems

### Architecture

```
┌─────────┐   SQRL Auth    ┌──────────┐   OAuth2    ┌──────────┐
│  User   │───────────────>│ SQRL SSP │────────────>│Authentik │
└─────────┘                └──────────┘             └─────┬────┘
                                                          │
                                                          │ OAuth2
                                                          │ Token
                                                          ▼
                                                   ┌──────────────┐
                                                   │    Gitea     │
                                                   │ (Git Hosting)│
                                                   └──────────────┘
```

### Implementation Approach
1. Deploy Gitea
2. Configure Gitea OAuth2 authentication
3. Point to Authentik as OAuth2 provider
4. Users authenticate via SQRL → Authentik → Gitea

### Success Criteria
- [ ] Gitea deployed and configured
- [ ] OAuth2 authentication working
- [ ] Users can access Git repos after SQRL authentication
- [ ] Documentation complete

### Risk Assessment
- **Risk Level:** 🟢 **LOW** (standard OAuth2 integration)
- **Complexity:** Medium
- **Recommendation:** Standard OAuth2 flow, well-documented

---

## Dependency Version Matrix

### Current State (After All Stages)

| Dependency | Version | Purpose | Stage |
|-----------|---------|---------|-------|
| **Go** | 1.25.4 | Runtime | Stage 0 ✅ |
| **yeqown/go-qrcode** | v2.3.1 | QR generation | Stage 1 |
| **etcd/client/v3** | v3.6.6 | Distributed storage | Stage 3 |
| **golang.org/x/oauth2** | v0.15.0 | OAuth2 client | Stage 4 |
| **Authentik** | 2025.8+ | Identity provider | Stage 4 |
| **Gitea** | Latest | Git hosting | Stage 5 |

### Compatibility Matrix

| Component | Go 1.25.4 | etcd v3.6 | Authentik 2025.8 | Gitea |
|-----------|-----------|-----------|------------------|-------|
| SQRL SSP | ✅ Native | ✅ Client | ✅ API | ✅ OAuth2 |
| yeqown/go-qrcode | ✅ Compatible | N/A | N/A | N/A |
| etcd client/v3 | ✅ Compatible | ✅ Native | N/A | N/A |
| Authentik | N/A (API) | ✅ Can use | ✅ Native | ✅ OAuth2 |
| Gitea | N/A (separate) | N/A | ✅ OAuth2 | ✅ Native |

---

## Testing Strategy by Stage

### Stage 1: QR Code Replacement
- **Unit Tests:** QR generation with various inputs
- **Integration Tests:** Full `/png.sqrl` endpoint
- **Visual Tests:** Manual QR code scanning
- **Performance Tests:** QR generation throughput

### Stage 2: Test Coverage
- **Unit Tests:** Comprehensive coverage of all components
- **Integration Tests:** Full authentication flows
- **Security Tests:** Input validation, injection attempts
- **Benchmark Tests:** Performance baselines

### Stage 3: etcd Integration
- **Unit Tests:** Hoard and AuthStore implementations
- **Integration Tests:** Multi-server shared state
- **Failure Tests:** etcd node failures, network partitions
- **Performance Tests:** etcd operation latency

### Stage 4: Authentik Integration
- **Unit Tests:** API client functionality
- **Integration Tests:** Full SQRL → Authentik → OAuth2 flow
- **Security Tests:** Token validation, user creation
- **Load Tests:** Concurrent authentication requests

### Stage 5: Gitea Integration
- **Integration Tests:** OAuth2 authentication flow
- **End-to-End Tests:** Git operations with SQRL authentication
- **User Tests:** Manual testing of Git workflows

---

## Rollback Procedures

### General Rollback Process
```bash
# 1. Identify problematic stage
git log --oneline --graph

# 2. Revert PR merge
git revert -m 1 <merge-commit-hash>

# 3. Or delete feature branch (if not merged)
git branch -D claude/dependency-upgrade-stageN-xxx

# 4. Or checkout previous stable state
git checkout <previous-stable-tag>

# 5. Redeploy
./deploy.sh
```

### Stage-Specific Rollbacks

**Stage 1 (QR Code):**
- Impact: Low, isolated to PNG generation
- Rollback: Revert go.mod and handers.go

**Stage 2 (Tests):**
- Impact: None (only adds tests)
- Rollback: Not needed (tests can remain)

**Stage 3 (etcd):**
- Impact: High if deployed with etcd
- Rollback: Remove -etcd flag, use MapHoard

**Stage 4 (Authentik):**
- Impact: Medium, affects OAuth2 integration
- Rollback: Remove Authentik components, direct SQRL flow

**Stage 5 (Gitea):**
- Impact: Low, separate service
- Rollback: Disable Gitea OAuth2, use alternative auth

---

## Timeline & Resource Planning

### Minimum Viable Upgrade (Stages 0-2)
**Duration:** 4-5 weeks
**Resources:** 1 developer
**Deliverables:**
- Modern QR library
- 80% test coverage
- Production-ready for single-server deployment

### Full Production Deployment (Stages 0-3)
**Duration:** 8-10 weeks
**Resources:** 1-2 developers
**Deliverables:**
- Modern QR library
- 80% test coverage
- Distributed storage with etcd
- Horizontal scaling capability

### Complete Integration (All Stages)
**Duration:** 16-20 weeks
**Resources:** 2-3 developers
**Deliverables:**
- All of the above
- OAuth2/OIDC support via Authentik
- Git hosting integration
- Full enterprise feature set

### Gantt Chart

```
Stage 0: Baseline              [====] ✅ Complete
Stage 1: QR Code               [----] Week 1
Stage 2: Test Coverage         [--------] Weeks 2-4
Stage 3: etcd                  [------] Weeks 5-7
Stage 4: Authentik (Optional)  [--------] Weeks 8-11
Stage 5: Gitea (Optional)      [------] Weeks 12-14
```

---

## Decision Matrix for Optional Stages

| Criterion | Stage 3 (etcd) | Stage 4 (Authentik) | Stage 5 (Gitea) |
|-----------|----------------|---------------------|-----------------|
| **Multi-server needed?** | Required | Optional | No |
| **OAuth2/OIDC needed?** | No | Required | No |
| **Git hosting needed?** | No | No | Required |
| **Complexity** | High | Very High | Medium |
| **Maintenance burden** | Medium | High | Medium |
| **Infrastructure cost** | Medium | High | Medium |

### Recommendation Flow

```
Do you need horizontal scaling?
├─ NO  → Skip Stage 3 (use MapHoard)
└─ YES → Implement Stage 3 (etcd)
           │
           ▼
        Do you need OAuth2/OIDC?
        ├─ NO  → Skip Stage 4
        └─ YES → Implement Stage 4 (Authentik)
                   │
                   ▼
                Do you need Git hosting?
                ├─ NO  → Done
                └─ YES → Implement Stage 5 (Gitea)
```

---

## Success Metrics

### Stage 1 Success Metrics
- [x] Zero downtime migration
- [x] QR codes scannable by all SQRL clients
- [x] No performance regression
- [x] All tests passing

### Stage 2 Success Metrics
- [x] Coverage ≥80% (enforced by CI)
- [x] All critical paths tested
- [x] Security tests in place
- [x] Test execution time <60s

### Stage 3 Success Metrics
- [x] Multi-server shared state working
- [x] etcd cluster highly available
- [x] Performance: ≥400 QPS with etcd
- [x] Automatic failover working

### Stage 4 Success Metrics
- [x] OAuth2 tokens issued correctly
- [x] SQRL → Authentik user sync working
- [x] Third-party apps can authenticate
- [x] Token refresh working

### Stage 5 Success Metrics
- [x] Git operations work with SQRL authentication
- [x] OAuth2 flow to Gitea working
- [x] Users can push/pull repos
- [x] Access control enforced

---

## Monitoring & Observability

### Metrics to Track (All Stages)

**Performance:**
- Request latency (p50, p95, p99)
- Throughput (requests/second)
- Error rate
- Test coverage percentage

**Security:**
- Failed authentication attempts
- Vulnerability scan results
- Dependency update lag time

**Reliability:**
- Uptime percentage
- etcd cluster health (Stage 3+)
- Authentik availability (Stage 4+)

### Alerting Thresholds

```yaml
# Prometheus-style alerts
groups:
  - name: sqrl_ssp
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.05

      - alert: TestCoverageBelowThreshold
        expr: test_coverage_percent < 80

      - alert: EtcdDown
        expr: up{job="etcd"} == 0
```

---

## Communication Plan

### Stakeholder Updates

**Weekly Status (During Active Development):**
- Stage progress
- Blockers identified
- Next week plan

**Stage Completion (After Each Stage):**
- Success criteria met
- Deployment notes
- Known issues

**Go/No-Go Decision Points:**
- Before Stage 3: Evaluate need for distributed storage
- Before Stage 4: Evaluate need for OAuth2/OIDC
- Before Stage 5: Evaluate need for Git hosting

---

## Appendix: Compatibility Testing Matrix

### Test Combinations

| Stage | Go 1.25.4 | Tests Pass | Coverage ≥80% | etcd v3.6 | Authentik | Gitea |
|-------|-----------|------------|---------------|-----------|-----------|-------|
| 0 | ✅ | ✅ | ❌ (29.4%) | N/A | N/A | N/A |
| 1 | ✅ | ✅ | ❌ (29.4%) | N/A | N/A | N/A |
| 2 | ✅ | ✅ | ✅ (80%+) | N/A | N/A | N/A |
| 3 | ✅ | ✅ | ✅ | ✅ | N/A | N/A |
| 4 | ✅ | ✅ | ✅ | ✅ | ✅ | N/A |
| 5 | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |

---

## Conclusion

This staged dependency upgrade plan provides a safe, incremental path to modernizing the SQRL SSP codebase. Each stage is independently testable, deployable, and reversible, ensuring system stability throughout the upgrade process.

### Recommended Path for Most Users
1. **Stage 1:** Replace QR library (security & maintenance)
2. **Stage 2:** Increase test coverage (production readiness)
3. **Stage 3:** Add etcd ONLY if horizontal scaling needed
4. **Stages 4-5:** Evaluate based on specific requirements

### Next Steps
1. Review this plan with stakeholders
2. Decide which stages are required for your use case
3. Schedule Stage 1 implementation
4. Execute staged rollout with testing at each phase

---

**Document Version:** 1.0
**Last Updated:** November 18, 2025
**Status:** Ready for Review
**Owner:** Development Team
