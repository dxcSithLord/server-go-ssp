# Consolidated TODO List
**Project:** SQRL SSP (Server-Side Protocol)
**Repository:** github.com/dxcSithLord/server-go-ssp
**Date:** November 18, 2025
**Status:** Comprehensive Planning Document

---

## Overview

This document consolidates ALL outstanding TODO items from code comments, documentation files, and planning documents into a single prioritized action list. Items are categorized by priority, area, and dependencies.

---

## Priority Levels

- 🔴 **CRITICAL:** Security vulnerabilities, blocking issues, production readiness
- 🟠 **HIGH:** Important features, significant technical debt
- 🟡 **MEDIUM:** Nice-to-have features, optimizations
- 🟢 **LOW:** Future enhancements, optional improvements

---

## TODO Items by Source

### From Code Comments (7 items)

#### 1. handers.go:20 - Additional SQRL Parameters
**Priority:** 🟡 MEDIUM
**Category:** Feature Enhancement
**Status:** Not Started

```go
// TODO sin, ask and 1-9 params
```

**Description:** Implement additional SQRL protocol parameters:
- `sin` - Server information
- `ask` - User prompt messages
- Buttons 1-9 - User response options

**Implementation Plan:**
1. Add `sin` field to `ServerResponse` struct
2. Implement `Ask` mechanism for user prompts
3. Add button response handling in `CliRequest`
4. Update tests for new parameters

**Effort:** Medium (2-3 days)
**Dependencies:** None
**References:** GRC SQRL Specification Section 6.3

---

#### 2. cli_handler.go:231 - Clear PreviousIDMatch Flag
**Priority:** 🟡 MEDIUM
**Category:** Security/Correctness
**Status:** Not Started

```go
// TODO should we clear the PreviousIDMatch here?
```

**Context:** During identity key rotation, should PreviousIDMatch TIF flag be cleared after successful rotation?

**Decision Required:**
- **Option A:** Clear flag immediately after successful rotation
  - Pro: Cleaner state, prevents confusion
  - Con: May break multi-step key rotation flows

- **Option B:** Keep flag until session expires
  - Pro: Allows verification of rotation in same session
  - Con: May leak information about previous identity

**Recommendation:** Clear the flag after successful rotation (Option A) for better security hygiene.

**Effort:** Low (1 hour)
**Dependencies:** Identity rotation tests
**Action:** Add to Stage 2 (Test Coverage) - decision + implementation + tests

---

#### 3. cli_handler.go:317 - Remove After Signature Failure
**Priority:** 🟡 MEDIUM
**Category:** Security/Error Handling
**Status:** Not Started

```go
// TODO: remove since sig check failed here?
```

**Context:** When signature verification fails, should the identity be automatically removed?

**Decision Required:**
- **Option A:** Remove identity on signature failure
  - Pro: Prevents attacks using stolen identity data
  - Con: May lock out legitimate users with clock skew or corrupted data

- **Option B:** Disable identity on signature failure
  - Pro: Allows recovery via unlock process
  - Con: Requires additional unlock step

- **Option C:** Increment failure counter, remove after N failures
  - Pro: Balances security and usability
  - Con: More complex implementation

**Recommendation:** Option C (failure counter) with configurable threshold (default: 3-5 failures)

**Effort:** Medium (1 day)
**Dependencies:** AuthStore needs failure tracking
**Action:** Defer to post-Stage 3 (requires distributed counter)

---

#### 4. cli_response.go:238 - Support Version Ranges
**Priority:** 🟢 LOW
**Category:** Protocol Compliance
**Status:** Not Started

```go
// TODO be less lazy and support ranges
```

**Context:** SQRL protocol allows version ranges (e.g., "1-2,4"), currently only supports single version "1"

**Implementation:**
1. Parse version ranges in `NewFromRequest`
2. Select highest supported version from range
3. Return selected version in response

**Effort:** Low (2-3 hours)
**Dependencies:** None
**References:** SQRL Spec Section 4.4

---

#### 5. api.go:65 - Track Previous Identity Key
**Priority:** 🟡 MEDIUM
**Category:** Data Model
**Status:** Not Started

```go
Pidk string `json:"pidk"` // TODO do we need to keep track of Pidk?
```

**Decision Required:**
Should `Pidk` be stored persistently, or only during rotation?

**Analysis:**
- **Current:** Pidk stored in `SqrlIdentity` struct
- **Use Case:** Track identity key rotation history
- **Privacy:** May expose identity rotation timeline
- **Security:** Useful for detecting suspicious rotation patterns

**Options:**
- **Option A:** Keep Pidk permanently
  - Pro: Full audit trail of identity changes
  - Con: Larger storage, potential privacy concern

- **Option B:** Clear Pidk after successful rotation
  - Pro: Minimal storage, better privacy
  - Con: No historical audit trail

- **Option C:** Store rotation history separately (audit log)
  - Pro: Audit trail without bloating identity record
  - Con: Additional storage implementation

**Recommendation:** Option C - move to separate audit log table (future feature)

**Effort:** Low (remove field) OR Medium (implement audit log)
**Dependencies:** Audit logging system (not yet implemented)
**Action:** Defer until audit logging is designed (post-MVP)

---

#### 6. cli_request.go:78 - Support Version Ranges (Duplicate)
**Priority:** 🟢 LOW
**Category:** Protocol Compliance
**Status:** Not Started

```go
// TODO be less lazy and support ranges
```

**Same as item #4** - consolidate implementation in both files.

---

#### 7. cli_request.go:145 - Handle Multiple Versions and Ranges
**Priority:** 🟢 LOW
**Category:** Protocol Compliance
**Status:** Not Started

```go
// TODO handle multiple versions and ranges
```

**Same as items #4 and #6** - part of comprehensive version negotiation implementation.

---

### From SECURITY_REVIEW.md (Multiple items)

#### 8. Increase Test Coverage: 29.4% → 80%
**Priority:** 🔴 CRITICAL
**Category:** Testing/Quality
**Status:** ⚠️ IN PROGRESS (29.4% achieved, target 80%)

**Components Requiring Coverage:**

| Component | Current | Target | Priority | Effort |
|-----------|---------|--------|----------|--------|
| cli_request.go | ~10% | 90% | 🔴 CRITICAL | 3-4 days |
| cli_handler.go | 0% | 85% | 🔴 CRITICAL | 4-5 days |
| cli_response.go | ~15% | 90% | 🟠 HIGH | 2-3 days |
| api.go | 0% | 80% | 🟠 HIGH | 2-3 days |
| handers.go | 0% | 85% | 🟠 HIGH | 2-3 days |
| map_auth_store.go | 0% | 90% | 🟠 HIGH | 1-2 days |
| grc_tree.go | ~40% | 95% | 🟡 MEDIUM | 1-2 days |
| random_tree.go | ~40% | 95% | 🟡 MEDIUM | 1-2 days |
| map_hoard.go | ~50% | 95% | 🟡 MEDIUM | 1-2 days |
| secure_clear.go | NEW | 100% | 🔴 CRITICAL | 1 day |
| secure_log.go | Partial | 95% | 🟠 HIGH | 1 day |

**Total Effort:** 2-3 weeks
**Status:** Mapped to Stage 2 in DEPENDENCY_UPGRADE_PLAN.md
**CI Enforcement:** Already configured (pipeline will fail if <80%)

---

#### 9. Implement Rate Limiting
**Priority:** 🟠 HIGH
**Category:** Security
**Status:** Not Started

**Vulnerability:** No protection against brute force attacks on `/cli.sqrl`

**Implementation Options:**
- **Option A:** In-memory rate limiter (golang.org/x/time/rate)
  - Pro: Simple, no dependencies
  - Con: Not shared across servers

- **Option B:** Redis-based rate limiter
  - Pro: Shared across servers
  - Con: Requires Redis dependency

- **Option C:** etcd-based rate limiter
  - Pro: Uses existing etcd infrastructure (Stage 3+)
  - Con: Higher latency than Redis

**Recommendation:**
- Stage 1-2: Option A (in-memory)
- Stage 3+: Option C (etcd-based)

**Suggested Limits:**
- `/cli.sqrl`: 10 requests per minute per IP
- `/nut.sqrl`: 60 requests per minute per IP
- `/png.sqrl`: 30 requests per minute per IP

**Effort:** Medium (2-3 days)
**Dependencies:** Stage 3 for distributed rate limiting
**Action:** Add to Stage 2 for single-server, enhance in Stage 3

---

#### 10. Implement Request Size Limits
**Priority:** 🔴 CRITICAL
**Category:** Security (DoS Prevention)
**Status:** Not Started

**Vulnerability:** No protection against large request body DoS attacks

**Implementation:**
```go
// In server/main.go or middleware
server := &http.Server{
    MaxHeaderBytes: 1 << 20, // 1 MB
    ReadTimeout:    15 * time.Second,
    WriteTimeout:   15 * time.Second,
    IdleTimeout:    60 * time.Second,
}

// Add max body size middleware
func maxBodySize(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB max
        next.ServeHTTP(w, r)
    })
}
```

**Suggested Limits:**
- Request Headers: 1 MB
- Request Body: 1 MB (SQRL requests are typically <10 KB)
- ReadTimeout: 15 seconds

**Effort:** Low (1-2 hours)
**Dependencies:** None
**Action:** Add to Stage 1 (immediate security fix)

---

#### 11. Add Input Validation Bounds Checking
**Priority:** 🟠 HIGH
**Category:** Security
**Status:** Not Started

**Areas Requiring Validation:**
1. Nut length validation (should be exactly 22 characters)
2. Signature length validation (ED25519 signatures are 64 bytes)
3. Public key length validation (ED25519 keys are 32 bytes)
4. Command validation (only allow: query, ident, enable, disable, remove)
5. Base64url decoding error handling

**Implementation:**
```go
// In cli_request.go
func (cr *CliRequest) Validate() error {
    if len(cr.Client.Idk) != 44 { // Base64-encoded 32 bytes
        return fmt.Errorf("invalid idk length: %d", len(cr.Client.Idk))
    }

    validCommands := map[string]bool{
        "query": true, "ident": true, "enable": true,
        "disable": true, "remove": true,
    }

    if !validCommands[cr.Client.Cmd] {
        return fmt.Errorf("invalid command: %s", cr.Client.Cmd)
    }

    // ... more validation
}
```

**Effort:** Medium (2-3 days including tests)
**Dependencies:** Stage 2 (comprehensive testing)
**Action:** Add to Stage 2

---

### From UPGRADE_GO_1_25.md (Multiple items)

#### 12. Run govulncheck and Address Vulnerabilities
**Priority:** 🟠 HIGH
**Category:** Security
**Status:** Partially Complete (Go upgraded, govulncheck needed)

**Remaining Actions:**
```bash
# Install latest govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest

# Run vulnerability check
govulncheck ./...

# Address any findings
```

**Effort:** Low (1-2 hours if no vulnerabilities found)
**Schedule:** Weekly automated scans (via CI/CD)
**Action:** Add to Stage 1 CI/CD pipeline

---

#### 13. Performance Benchmarking (Go 1.25.4)
**Priority:** 🟡 MEDIUM
**Category:** Performance
**Status:** Not Started

**Objective:** Establish performance baselines with Go 1.25.4

**Benchmarks to Create:**
```bash
# Cryptographic operations
BenchmarkED25519Verify
BenchmarkAESEncrypt
BenchmarkAESDecrypt

# Nut generation
BenchmarkRandomTreeGenerate
BenchmarkGrcTreeGenerate

# Request processing
BenchmarkCliRequestParse
BenchmarkCliResponseEncode

# Full flow
BenchmarkEndToEndAuthentication
```

**Effort:** Medium (2-3 days)
**Dependencies:** Stage 2 (test coverage)
**Action:** Add to Stage 2

---

#### 14. Memory Leak Detection with ASAN
**Priority:** 🟡 MEDIUM
**Category:** Quality/Security
**Status:** Not Started

**Implementation:**
```bash
# Build with Address Sanitizer (Go 1.25 feature)
go build -asan -o sqrl-server-asan ./server

# Run under load and monitor for leaks
./sqrl-server-asan &
# Run load tests
# Check for leak reports on exit
```

**Effort:** Low (1 day)
**Dependencies:** Go 1.25.4 (✅ complete)
**Action:** Add to Stage 2 (quality assurance)

---

#### 15. Update README.md for Go 1.25 Requirements
**Priority:** 🟡 MEDIUM
**Category:** Documentation
**Status:** Not Started

**Updates Needed:**
- Minimum Go version: 1.25.0
- Recommended Go version: 1.25.4
- Installation instructions
- Build instructions

**Effort:** Low (1 hour)
**Dependencies:** None
**Action:** Add to Stage 1

---

### From DEPENDENCY_UPGRADE_PLAN.md (Multiple items)

#### 16. Stage 1: Replace QR Code Library
**Priority:** 🟠 HIGH
**Category:** Security/Maintenance
**Status:** ✅ PLANNED (detailed in upgrade plan)

**Summary:**
- Remove: `github.com/skip2/go-qrcode` (unmaintained since 2020)
- Add: `github.com/yeqown/go-qrcode` v2.3.1
- Effort: 1 week
- Status: Fully documented in DEPENDENCY_UPGRADE_PLAN.md Stage 1

---

#### 17. Stage 3: Implement etcd Distributed Storage
**Priority:** 🟡 MEDIUM (Optional, only if multi-server needed)
**Category:** Scalability
**Status:** ✅ PLANNED (detailed in upgrade plan)

**Summary:**
- Add: `go.etcd.io/etcd/client/v3` v3.6.6
- Implement: `EtcdHoard`, `EtcdAuthStore`
- Effort: 2-3 weeks
- Status: Fully documented in DEPENDENCY_UPGRADE_PLAN.md Stage 3

---

#### 18. Stage 4: Integrate Authentik for OAuth2/OIDC
**Priority:** 🟢 LOW (Optional, only if OAuth2 needed)
**Category:** Integration
**Status:** ✅ PLANNED (detailed in upgrade plan)

**Summary:**
- Integrate: Authentik 2025.8+
- Implement: SQRL → Authentik → OAuth2 bridge
- Effort: 3-4 weeks
- Status: Fully documented in DEPENDENCY_UPGRADE_PLAN.md Stage 4

---

#### 19. Stage 5: Integrate Git Hosting (Gitea)
**Priority:** 🟢 LOW (Optional, only if Git hosting needed)
**Category:** Integration
**Status:** ✅ PLANNED (detailed in upgrade plan)

**Summary:**
- Deploy: Gitea (latest)
- Configure: OAuth2 authentication via Authentik
- Effort: 2-3 weeks
- Status: Fully documented in DEPENDENCY_UPGRADE_PLAN.md Stage 5

---

### Missing/Inferred TODOs (Not explicitly documented)

#### 20. Implement Production-Ready Storage Backends
**Priority:** 🟠 HIGH
**Category:** Production Readiness
**Status:** Partially Complete (interfaces exist, implementations needed)

**Current State:**
- ✅ `MapHoard` - in-memory (NOT for production)
- ✅ `MapAuthStore` - in-memory (NOT for production)
- ❌ Production hoard implementations
- ❌ Production auth store implementations

**Recommended Implementations:**

**Option 1: Redis + PostgreSQL**
```go
// RedisHoard - already exists in separate repo
import "github.com/sqrldev/server-go-ssp-redishoard"

// PostgresAuthStore via GORM
import "github.com/sqrldev/server-go-ssp-gormauthstore"
```

**Option 2: etcd (all-in-one)**
```go
// Stage 3: EtcdHoard + EtcdAuthStore
// Single distributed storage backend
```

**Effort:**
- Option 1: Low (1-2 days) - existing implementations
- Option 2: Medium (2-3 weeks) - Stage 3 plan

**Dependencies:**
- Option 1: Redis + PostgreSQL/MySQL deployment
- Option 2: etcd cluster deployment

**Action:**
- Short-term: Document existing Redis/GORM implementations
- Long-term: Implement etcd backends (Stage 3)

---

#### 21. Implement Audit Logging
**Priority:** 🟡 MEDIUM
**Category:** Compliance/Security
**Status:** Not Started

**Requirements:**
- Log all authentication attempts (success/failure)
- Log identity lifecycle events (create, disable, enable, remove)
- Log key rotation events
- Store logs in append-only storage
- Support log retention policies

**Implementation:**
```go
// audit_log.go
package ssp

type AuditEvent struct {
    Timestamp  time.Time
    EventType  string // "auth", "create", "disable", etc.
    Idk        string // Identity key (truncated)
    RemoteIP   string // Masked IP
    Command    string
    Success    bool
    TIF        int
    Error      string
}

type AuditLogger interface {
    Log(event *AuditEvent) error
    Query(filter AuditFilter) ([]AuditEvent, error)
}
```

**Storage Options:**
- File-based (JSON lines)
- Database (PostgreSQL with partitioning)
- Log aggregation (CloudWatch, Elasticsearch)

**Effort:** Medium (1 week)
**Dependencies:** None
**Action:** Defer to post-MVP (not critical for initial deployment)

---

#### 22. Implement Monitoring and Metrics
**Priority:** 🟠 HIGH
**Category:** Operations
**Status:** Not Started

**Metrics to Expose:**
- Request counters (by endpoint, status code)
- Request duration histograms (p50, p95, p99)
- Nut generation rate
- Active nut count
- Identity count
- Authentication success/failure rate
- etcd operation latency (Stage 3+)

**Implementation:**
```go
// metrics.go
package ssp

import "github.com/prometheus/client_golang/prometheus"

var (
    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "sqrl_http_request_duration_seconds",
            Help: "HTTP request duration in seconds",
        },
        []string{"endpoint", "status"},
    )

    authAttempts = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "sqrl_auth_attempts_total",
            Help: "Total authentication attempts",
        },
        []string{"command", "success"},
    )
)
```

**Effort:** Medium (3-4 days)
**Dependencies:** None (stdlib compatible)
**Action:** Add to Stage 2 or Stage 3

---

#### 23. Implement Health Check Endpoints
**Priority:** 🟠 HIGH
**Category:** Operations
**Status:** Not Started

**Endpoints to Add:**
```
GET /health/live   - Liveness probe (server is running)
GET /health/ready  - Readiness probe (dependencies available)
GET /health/startup - Startup probe (initialization complete)
```

**Implementation:**
```go
// health.go
package ssp

func (api *SqrlSspAPI) Healthz(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}

func (api *SqrlSspAPI) Readiness(w http.ResponseWriter, r *http.Request) {
    // Check etcd connection
    // Check database connection
    // Check tree is generating nuts

    if allHealthy {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("READY"))
    } else {
        w.WriteHeader(http.StatusServiceUnavailable)
        w.Write([]byte("NOT READY"))
    }
}
```

**Effort:** Low (1 day)
**Dependencies:** None
**Action:** Add to Stage 2

---

#### 24. Implement Graceful Shutdown
**Priority:** 🟡 MEDIUM
**Category:** Reliability
**Status:** Not Started

**Implementation:**
```go
// server/main.go
func main() {
    server := &http.Server{...}

    // Start server in goroutine
    go func() {
        if err := server.ListenAndServe(); err != http.ErrServerClosed {
            log.Fatalf("Server failed: %v", err)
        }
    }()

    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Shutting down server...")

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        log.Fatalf("Server forced to shutdown: %v", err)
    }

    log.Println("Server exited gracefully")
}
```

**Effort:** Low (2-3 hours)
**Dependencies:** None
**Action:** Add to Stage 2

---

#### 25. Configuration Management
**Priority:** 🟡 MEDIUM
**Category:** Operations
**Status:** Partial (command-line flags exist)

**Current:** Command-line flags only
**Needed:**
- Environment variables
- Configuration file (YAML/TOML)
- etcd-based config (Stage 3+)

**Implementation:**
```go
// config.go
package ssp

type Config struct {
    Server struct {
        Port         int
        TLSCertFile  string
        TLSKeyFile   string
        ReadTimeout  time.Duration
        WriteTimeout time.Duration
    }

    SQRL struct {
        NutExpiration time.Duration
        PagExpiration time.Duration
        Domain        string
        RootPath      string
    }

    Storage struct {
        Type          string // "memory", "etcd", "redis"
        EtcdEndpoints []string
        RedisAddr     string
    }
}

func LoadConfig(path string) (*Config, error) {
    // Load from YAML/TOML file
    // Override with environment variables
    // Override with command-line flags
}
```

**Effort:** Medium (2-3 days)
**Dependencies:** None
**Action:** Defer to post-MVP (current flags sufficient)

---

## Consolidated Priority Roadmap

### Immediate (Stage 1) - 1-2 Weeks

1. ✅ **Replace QR Code Library** (Item #16) - PLANNED
2. 🔴 **Add Request Size Limits** (Item #10) - NEW
3. 🟠 **Run govulncheck** (Item #12) - NEW
4. 🟡 **Update README for Go 1.25** (Item #15) - NEW

**Deliverables:** Secure, maintained dependencies

---

### Foundation (Stage 2) - 2-3 Weeks

1. 🔴 **Increase Test Coverage to 80%** (Item #8) - PLANNED
2. 🟠 **Implement Rate Limiting** (Item #9) - NEW
3. 🟠 **Add Input Validation** (Item #11) - NEW
4. 🟠 **Health Check Endpoints** (Item #23) - NEW
5. 🟡 **Graceful Shutdown** (Item #24) - NEW
6. 🟡 **Performance Benchmarks** (Item #13) - NEW
7. 🟡 **Memory Leak Detection** (Item #14) - NEW

**Deliverables:** Production-ready quality

---

### Scaling (Stage 3) - 2-3 Weeks (OPTIONAL)

1. ✅ **Implement etcd Storage** (Item #17) - PLANNED
2. 🟠 **Monitoring and Metrics** (Item #22) - NEW
3. 🟡 **etcd-based Rate Limiting** (enhance Item #9) - NEW

**Deliverables:** Horizontal scaling capability

---

### Enterprise (Stage 4+) - 3-4 Weeks (OPTIONAL)

1. ✅ **Integrate Authentik** (Item #18) - PLANNED
2. ✅ **Integrate Gitea** (Item #19) - PLANNED
3. 🟡 **Audit Logging** (Item #21) - NEW
4. 🟡 **Configuration Management** (Item #25) - NEW

**Deliverables:** Enterprise integration

---

### Future / Low Priority

1. 🟡 **Additional SQRL Parameters** (Item #1)
2. 🟡 **Version Range Support** (Items #4, #6, #7)
3. 🟡 **Clear PreviousIDMatch** (Item #2) - needs decision
4. 🟡 **Remove After Sig Failure** (Item #3) - needs decision
5. 🟡 **Pidk Storage Decision** (Item #5) - needs decision

**Deliverables:** Protocol completeness

---

## Summary Statistics

**Total TODO Items:** 25

**By Priority:**
- 🔴 CRITICAL: 3 items (12%)
- 🟠 HIGH: 9 items (36%)
- 🟡 MEDIUM: 12 items (48%)
- 🟢 LOW: 1 item (4%)

**By Status:**
- ✅ Planned: 4 items (already in DEPENDENCY_UPGRADE_PLAN.md)
- ⚠️ In Progress: 1 item (test coverage at 29.4%)
- ❌ Not Started: 20 items

**By Category:**
- Security: 6 items
- Testing/Quality: 5 items
- Features: 4 items
- Operations: 4 items
- Integration: 3 items
- Documentation: 2 items
- Performance: 1 item

**Estimated Total Effort:**
- Stage 1: 1-2 weeks
- Stage 2: 2-3 weeks
- Stage 3: 2-3 weeks (optional)
- Stage 4+: 3-4 weeks (optional)
- Future: TBD

**Minimum Viable Product (MVP):** Stages 1-2 (3-5 weeks)
**Production Multi-Server:** Stages 1-3 (7-10 weeks)
**Enterprise Platform:** All Stages (14-18 weeks)

---

**Document Version:** 1.0
**Last Updated:** November 18, 2025
**Status:** Comprehensive Planning Document
**Next Review:** After Stage 1 completion
