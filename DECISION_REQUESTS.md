# Decision Requests: Conflicts, Gaps, and Alternatives
**Project:** SQRL SSP (Server-Side Protocol)
**Repository:** github.com/dxcSithLord/server-go-ssp
**Date:** November 18, 2025
**Purpose:** Stakeholder Decision Points

---

## Overview

This document identifies areas where conflicts exist, requirements have gaps, or multiple alternatives are available. Each section presents the issue, available options, and requests stakeholder guidance to proceed.

**Format:**
- **Issue ID:** Unique identifier for tracking
- **Priority:** Impact level (Critical/High/Medium/Low)
- **Category:** Area affected
- **Decision Required:** What needs to be decided
- **Options:** Available alternatives
- **Recommendation:** Technical team suggestion (if any)
- **Impact:** Consequences of decision
- **Deadline:** When decision is needed

---

## TABLE OF CONTENTS

1. [CRITICAL DECISIONS](#critical-decisions)
2. [HIGH PRIORITY DECISIONS](#high-priority-decisions)
3. [MEDIUM PRIORITY DECISIONS](#medium-priority-decisions)
4. [LOW PRIORITY DECISIONS](#low-priority-decisions)
5. [ARCHITECTURAL DECISIONS](#architectural-decisions)
6. [INTEGRATION DECISIONS](#integration-decisions)
7. [SECURITY DECISIONS](#security-decisions)

---

## CRITICAL DECISIONS

### DECISION-001: Production Storage Backend Choice

**Priority:** 🔴 CRITICAL
**Category:** Architecture/Infrastructure
**Deadline:** Before Stage 3 planning (Week 6)

#### Issue
Current `MapHoard` and `MapAuthStore` are in-memory only and NOT suitable for production. Must choose production storage backend.

#### Options

**Option A: Redis + PostgreSQL (Existing Implementations)**
- **Description:** Use existing community implementations
  - RedisHoard: `github.com/sqrldev/server-go-ssp-redishoard`
  - PostgresAuthStore: `github.com/sqrldev/server-go-ssp-gormauthstore`
- **Pros:**
  - ✅ Implementations already exist (1-2 day integration)
  - ✅ Well-understood technologies
  - ✅ Separate concerns (ephemeral vs persistent)
  - ✅ PostgreSQL provides rich querying
- **Cons:**
  - ❌ Two separate systems to deploy and maintain
  - ❌ No distributed coordination (sticky sessions required)
  - ❌ Horizontal scaling requires external session management

**Option B: etcd (All-in-One Distributed)**
- **Description:** Use etcd v3.6.6 for both Hoard and AuthStore (Stage 3 plan)
- **Pros:**
  - ✅ Single distributed system
  - ✅ Native horizontal scaling (no sticky sessions)
  - ✅ Strong consistency guarantees
  - ✅ Built-in TTL for nut expiration
  - ✅ Multi-datacenter support
- **Cons:**
  - ❌ Higher implementation effort (2-3 weeks)
  - ❌ etcd cluster complexity (3-5 nodes)
  - ❌ Higher operational complexity
  - ❌ Not ideal for large-scale identity storage

**Option C: Hybrid (Redis + PostgreSQL + etcd)**
- **Description:** Redis for Hoard, PostgreSQL for AuthStore, etcd for coordination
- **Pros:**
  - ✅ Best tool for each job
  - ✅ Optimal performance
  - ✅ Distributed coordination via etcd
- **Cons:**
  - ❌ Three systems to maintain
  - ❌ Highest operational complexity
  - ❌ Most expensive infrastructure

**Option D: Defer (Continue with MapHoard)**
- **Description:** Postpone decision, continue with in-memory storage
- **Pros:**
  - ✅ No immediate effort required
  - ✅ Simplest deployment
- **Cons:**
  - ❌ NOT production-ready
  - ❌ No horizontal scaling
  - ❌ Data loss on restart
  - ❌ **UNACCEPTABLE FOR PRODUCTION**

#### Recommendation
**Short-term (MVP, single-server):** Option A (Redis + PostgreSQL)
- Fast time to market
- Proven implementations
- Good enough for single-server production

**Long-term (multi-server, enterprise):** Option B (etcd)
- True horizontal scaling
- Better architecture for distributed deployments
- Worth the investment for multi-datacenter

**Hybrid approach:** Start with Option A, migrate to Option B in Stage 3 when scaling is needed.

#### Impact
- **Performance:** Option B (etcd) may have higher latency than Option A (Redis)
- **Cost:** Option C most expensive, Option A moderate, Option B moderate (3-node cluster)
- **Complexity:** Option A lowest, Option B medium, Option C highest
- **Scalability:** Option B best, Option C good, Option A limited

#### Request for Advice
**Questions for stakeholders:**
1. What is the expected deployment model? (single-server vs multi-server vs multi-datacenter)
2. What is the expected scale? (users, requests/second, geographic distribution)
3. What is the operational team's comfort level with etcd vs Redis/PostgreSQL?
4. Is there existing infrastructure (Redis, PostgreSQL) that can be reused?
5. What is the budget for infrastructure?

**Please provide guidance on:**
- Preferred approach (A, B, or C)
- Timeline constraints
- Infrastructure preferences

---

### DECISION-002: Test Coverage Priority

**Priority:** 🔴 CRITICAL
**Category:** Quality/Security
**Deadline:** Before Stage 2 begins (Week 2)

#### Issue
Current test coverage is 29.4%, target is 80%. Need to prioritize which components to test first within Stage 2's 2-3 week timeline.

#### Options

**Option A: Critical Path First**
- **Priority Order:**
  1. cli_handler.go (authentication logic) - 4-5 days
  2. cli_request.go (signature verification) - 3-4 days
  3. secure_clear.go (memory safety) - 1 day
  4. Continue with others until 80% reached
- **Pros:**
  - ✅ Highest risk areas covered first
  - ✅ Can deploy with partial coverage if time-constrained
- **Cons:**
  - ❌ May not reach 80% overall

**Option B: Breadth First**
- **Priority Order:**
  1. Add 10-15% coverage to all files
  2. Iterate until 80% overall reached
- **Pros:**
  - ✅ Even coverage across codebase
  - ✅ Reaches 80% threshold faster
- **Cons:**
  - ❌ Critical paths may remain under-tested
  - ❌ Lower quality tests (just hitting percentage)

**Option C: Blended Approach**
- **Priority Order:**
  1. Week 1: Critical paths (cli_handler, cli_request, secure_clear) to 70%+
  2. Week 2: API surface (api.go, handers.go) to 60%+
  3. Week 3: Storage and utilities to 85%+, raise others to 90%+
- **Pros:**
  - ✅ Balanced approach
  - ✅ Critical paths well-covered
  - ✅ Achieves 80% overall
- **Cons:**
  - ❌ Requires full 3 weeks

#### Recommendation
**Option C (Blended Approach)** - Best balance of risk mitigation and comprehensive coverage.

#### Impact
- **Security:** Option A provides better security coverage early
- **CI/CD:** CI will fail if <80%, so Option B/C more likely to pass
- **Timeline:** Option C requires full 3 weeks
- **Quality:** Option C produces highest quality tests

#### Request for Advice
**Questions for stakeholders:**
1. Is 3-week timeline for Stage 2 acceptable?
2. Can we deploy with <80% if critical paths are well-tested?
3. Should we prioritize CI passing (80%) or critical path coverage?

**Please provide guidance on:**
- Preferred option (A, B, or C)
- Timeline flexibility
- Minimum acceptable coverage for critical files

---

## HIGH PRIORITY DECISIONS

### DECISION-003: Rate Limiting Strategy

**Priority:** 🟠 HIGH
**Category:** Security
**Deadline:** Before Stage 2 completion (Week 5)

#### Issue
No rate limiting currently implemented. Need to prevent brute force attacks.

#### Options

**Option A: In-Memory Rate Limiting**
- **Implementation:** `golang.org/x/time/rate` package
- **Pros:**
  - ✅ Simple implementation (1-2 days)
  - ✅ No external dependencies
  - ✅ Very fast (in-process)
- **Cons:**
  - ❌ Per-server only (not distributed)
  - ❌ Attacker can bypass by hitting different servers
  - ❌ Requires sticky sessions or external coordination

**Option B: Distributed Rate Limiting (etcd)**
- **Implementation:** etcd v3.6.6 distributed counters
- **Pros:**
  - ✅ Works across all servers
  - ✅ Effective against distributed attacks
  - ✅ Uses existing etcd infrastructure (Stage 3)
- **Cons:**
  - ❌ Requires etcd deployment
  - ❌ Higher latency per request (~5-10ms)
  - ❌ Only available in Stage 3+

**Option C: Redis Rate Limiting**
- **Implementation:** Redis with sliding window algorithm
- **Pros:**
  - ✅ Distributed across servers
  - ✅ Very fast (<1ms latency)
  - ✅ Battle-tested pattern
- **Cons:**
  - ❌ Requires Redis deployment
  - ❌ Additional dependency (if not already using Redis)

**Option D: Progressive Approach**
- **Implementation:**
  - Stage 1-2: Option A (in-memory) for MVP
  - Stage 3: Migrate to Option B (etcd) or Option C (Redis)
- **Pros:**
  - ✅ Fast initial implementation
  - ✅ Upgradable path
  - ✅ Provides basic protection immediately
- **Cons:**
  - ❌ Need to implement twice
  - ❌ Early deployments not fully protected

#### Recommendation
**Option D (Progressive Approach)**
- Start with in-memory for single-server deployments
- Upgrade to etcd in Stage 3 for multi-server

**Suggested Rate Limits:**
```
/cli.sqrl:  10 requests/minute per IP
/nut.sqrl:  60 requests/minute per IP
/png.sqrl:  30 requests/minute per IP
/pag.sqrl: 120 requests/minute per IP (polling endpoint)
```

#### Impact
- **Security:** All options provide basic protection; distributed options (B/C) more robust
- **Performance:** A fastest, B slowest, C fast
- **Complexity:** A simplest, B/C moderate
- **Cost:** A free, B included in etcd, C requires Redis

#### Request for Advice
**Questions for stakeholders:**
1. Is single-server deployment acceptable initially?
2. Is Redis already part of infrastructure?
3. What is acceptable latency overhead for rate limiting?

**Please provide guidance on:**
- Preferred option
- Acceptable rate limits (requests/minute)
- IP-based vs identity-based rate limiting

---

### DECISION-004: Signature Verification Failure Handling

**Priority:** 🟠 HIGH
**Category:** Security/UX
**Deadline:** Before Stage 2 testing (Week 4)
**Related TODO:** CONSOLIDATED_TODO.md Item #3

#### Issue
When ED25519 signature verification fails, what should happen to the identity?

#### Context
- Signature failure could indicate:
  - Attack with stolen identity data
  - Clock skew between client and server
  - Corrupted SQRL client data
  - Legitimate client with software bug

#### Options

**Option A: Immediate Identity Removal**
```go
if !verifySignature() {
    authStore.DeleteIdentity(identity.Idk)
    return errorResponse("Signature verification failed")
}
```
- **Pros:**
  - ✅ Maximum security (prevents attack continuation)
  - ✅ Forces attacker to re-compromise identity
- **Cons:**
  - ❌ Locks out legitimate users with transient issues
  - ❌ No recovery path without re-enrollment
  - ❌ Poor user experience

**Option B: Disable Identity on Failure**
```go
if !verifySignature() {
    identity.Disabled = true
    authStore.SaveIdentity(identity)
    return errorResponse("Account disabled due to security concern")
}
```
- **Pros:**
  - ✅ Prevents further attacks
  - ✅ User can recover via unlock process (URS signature)
  - ✅ Better UX than removal
- **Cons:**
  - ❌ Attacker still has disabled identity in database
  - ❌ Requires unlock flow implementation

**Option C: Failure Counter with Threshold**
```go
identity.FailureCount++
if identity.FailureCount >= 3 {
    identity.Disabled = true
}
if identity.FailureCount >= 10 {
    authStore.DeleteIdentity(identity.Idk)
}
authStore.SaveIdentity(identity)
```
- **Pros:**
  - ✅ Tolerates transient failures (clock skew, network)
  - ✅ Progressive response (warn → disable → remove)
  - ✅ Balances security and usability
- **Cons:**
  - ❌ Gives attacker 3 attempts
  - ❌ Requires failure counter in data model
  - ❌ Need to reset counter on success

**Option D: Log and Continue (No Action)**
```go
if !verifySignature() {
    logSecurityEvent("Signature verification failed", identity.Idk)
    return errorResponse("Invalid signature")
}
```
- **Pros:**
  - ✅ No impact on legitimate users
  - ✅ Simplest implementation
- **Cons:**
  - ❌ No protection against attacks
  - ❌ **UNACCEPTABLE FROM SECURITY PERSPECTIVE**

#### Recommendation
**Option C (Failure Counter)** with thresholds:
- 1-2 failures: Log + error response
- 3-5 failures: Disable account
- 10+ failures: Remove identity

**Implementation:**
```go
type SqrlIdentity struct {
    // ... existing fields ...
    FailureCount    int       `json:"failure_count"`
    LastFailureTime time.Time `json:"last_failure_time"`
}

func handleSignatureFailure(identity *SqrlIdentity) {
    identity.FailureCount++
    identity.LastFailureTime = time.Now()

    switch {
    case identity.FailureCount >= 10:
        authStore.DeleteIdentity(identity.Idk)
        logSecurity("Identity removed after 10 failures", identity.Idk)

    case identity.FailureCount >= 3:
        identity.Disabled = true
        authStore.SaveIdentity(identity)
        logSecurity("Identity disabled after 3 failures", identity.Idk)

    default:
        authStore.SaveIdentity(identity)
        logSecurity("Signature failure", identity.Idk)
    }
}

func handleSignatureSuccess(identity *SqrlIdentity) {
    identity.FailureCount = 0
    identity.LastFailureTime = time.Time{}
    authStore.SaveIdentity(identity)
}
```

#### Impact
- **Security:** Option C provides good protection with tolerance for transient issues
- **UX:** Option C best user experience
- **Complexity:** Option C requires data model change + logic
- **Operations:** Need monitoring for failure rate spikes

#### Request for Advice
**Questions for stakeholders:**
1. What is acceptable false-positive rate (legitimate users locked out)?
2. What is acceptable false-negative rate (attacks succeeding)?
3. Should there be a time-based reset (e.g., reset counter after 24 hours)?
4. Should there be notification to user when account is disabled?

**Please provide guidance on:**
- Preferred option (A, B, or C)
- Failure thresholds (if Option C)
- Recovery process for disabled accounts

---

## MEDIUM PRIORITY DECISIONS

### DECISION-005: Previous Identity Key (Pidk) Storage

**Priority:** 🟡 MEDIUM
**Category:** Data Model/Privacy
**Deadline:** Before Stage 3 (Week 8)
**Related TODO:** CONSOLIDATED_TODO.md Item #5

#### Issue
Should `Pidk` (Previous Identity Key) be stored permanently, temporarily, or not at all?

#### Context
- SQRL supports identity key rotation
- During rotation, client sends both new Idk and previous Pidk
- Question: How long should Pidk be retained?

#### Options

**Option A: Permanent Storage**
```go
type SqrlIdentity struct {
    Idk  string
    Pidk string // Keep forever
}
```
- **Pros:**
  - ✅ Full audit trail of identity changes
  - ✅ Can detect suspicious rotation patterns
  - ✅ Useful for security forensics
- **Cons:**
  - ❌ Larger storage footprint
  - ❌ Potential privacy concern (rotation history linkable)
  - ❌ GDPR "right to erasure" complications

**Option B: Temporary Storage (Clear After Successful Rotation)**
```go
// During rotation
identity.Pidk = request.Pidk

// After rotation completes successfully
identity.Pidk = ""
authStore.SaveIdentity(identity)
```
- **Pros:**
  - ✅ Minimal storage
  - ✅ Better privacy (no historical linkage)
  - ✅ Easier GDPR compliance
- **Cons:**
  - ❌ No audit trail
  - ❌ Cannot detect repeated rotations

**Option C: Separate Audit Log**
```go
type SqrlIdentity struct {
    Idk  string
    Pidk string // Empty after rotation
}

type IdentityRotationEvent struct {
    Timestamp    time.Time
    OldIdk       string // Hashed or truncated
    NewIdk       string // Hashed or truncated
    RemoteIP     string // Masked
    Success      bool
}
```
- **Pros:**
  - ✅ Audit trail preserved
  - ✅ Identity record stays clean
  - ✅ Can apply different retention policies
  - ✅ Better privacy controls
- **Cons:**
  - ❌ Requires separate audit table
  - ❌ More implementation effort

**Option D: No Storage (Pidk Only Used During Request)**
```go
// Pidk never persisted, only used for signature verification during rotation
```
- **Pros:**
  - ✅ Simplest implementation
  - ✅ Best privacy
- **Cons:**
  - ❌ No rotation tracking at all
  - ❌ Cannot detect attacks

#### Recommendation
**Option C (Separate Audit Log)** - Best balance of security, privacy, and compliance.

**Rationale:**
- Security: Maintains audit trail for forensics
- Privacy: Identity record doesn't expose rotation history
- Compliance: Can apply retention policies to audit log (e.g., 90 days)
- Flexible: Can query audit log when needed, doesn't bloat main identity record

#### Impact
- **Storage:** Option A highest, Option D lowest, Option C moderate
- **Privacy:** Option D best, Option A worst, Option C good
- **Security:** Option A best audit trail, Option D no audit trail
- **Compliance:** Option C easiest for GDPR

#### Request for Advice
**Questions for stakeholders:**
1. Is identity rotation audit trail required for compliance?
2. What is the retention period for audit logs?
3. Are there privacy regulations (GDPR, CCPA) that apply?
4. Do we need to detect suspicious rotation patterns?

**Please provide guidance on:**
- Preferred option (A, B, C, or D)
- If Option C, what retention period for audit logs?

---

### DECISION-006: PreviousIDMatch TIF Flag Clearing

**Priority:** 🟡 MEDIUM
**Category:** Protocol Behavior
**Deadline:** Before Stage 2 testing (Week 4)
**Related TODO:** CONSOLIDATED_TODO.md Item #2

#### Issue
During identity key rotation, should the `PreviousIDMatch` TIF flag be cleared immediately after successful rotation?

#### Context
- When client rotates identity (Idk → NewIdk), server verifies both IDS (new key) and PIDS (previous key) signatures
- Server sets `PreviousIDMatch` TIF flag (0x02) to indicate previous identity verified
- Question: When should this flag be cleared?

#### Options

**Option A: Clear Immediately After Successful Rotation**
```go
if rotationSuccess {
    identity.Pidk = newIdentity.Idk
    identity.Idk = newIdentity.NewIdk
    // Don't set PreviousIDMatch in response
}
```
- **Pros:**
  - ✅ Clean state after rotation
  - ✅ Prevents confusion in subsequent requests
  - ✅ More secure (less information leakage)
- **Cons:**
  - ❌ Client cannot verify rotation completed in same session

**Option B: Keep Flag Until Session Expires**
```go
if rotationSuccess {
    identity.Pidk = newIdentity.Idk
    identity.Idk = newIdentity.NewIdk
    // Keep setting PreviousIDMatch in responses until nut expires
}
```
- **Pros:**
  - ✅ Client can verify rotation succeeded
  - ✅ Useful for multi-step rotation flows
- **Cons:**
  - ❌ Exposes rotation information longer
  - ❌ More complex state management

**Option C: Clear on Next Request**
```go
// First request after rotation: still set PreviousIDMatch
// Second request after rotation: clear flag
```
- **Pros:**
  - ✅ Client gets one confirmation
  - ✅ Then state is cleared
- **Cons:**
  - ❌ Adds complexity
  - ❌ Relies on client making second request

#### Recommendation
**Option A (Clear Immediately)** - Simplest and most secure.

**Rationale:**
- SQRL client can verify rotation by checking that server accepted the ident command
- No need to expose PreviousIDMatch in subsequent responses
- Aligns with principle of least privilege

#### Impact
- **Security:** Option A most secure, Option B least secure
- **UX:** Option B slightly better for client verification
- **Complexity:** Option A simplest, Option C most complex

#### Request for Advice
**Questions for stakeholders:**
1. Do SQRL clients rely on PreviousIDMatch flag for rotation confirmation?
2. Is multi-step key rotation required?
3. Should rotation be verifiable in same session?

**Please provide guidance on:**
- Preferred option (A, B, or C)

---

## LOW PRIORITY DECISIONS

### DECISION-007: Version Range Support Implementation

**Priority:** 🟢 LOW
**Category:** Protocol Compliance
**Deadline:** Future (post-MVP)
**Related TODO:** CONSOLIDATED_TODO.md Items #4, #6, #7

#### Issue
SQRL protocol supports version ranges (e.g., "1-2,4"), currently only version "1" is implemented.

#### Options

**Option A: Implement Full Version Range Support**
- Parse ranges like "1-2,4"
- Select highest supported version
- Return selected version in response
- **Effort:** Low (2-3 hours)

**Option B: Defer Indefinitely**
- Current implementation sufficient
- SQRL protocol version unlikely to change soon
- **Effort:** None

#### Recommendation
**Option B (Defer)** - Not critical, SQRL protocol stable at version 1.

**Revisit if:** SQRL protocol version 2 is released.

---

### DECISION-008: Additional SQRL Parameters (sin, ask, buttons)

**Priority:** 🟢 LOW
**Category:** Feature Enhancement
**Deadline:** Future (post-MVP)
**Related TODO:** CONSOLIDATED_TODO.md Item #1

#### Issue
SQRL protocol supports additional parameters for enhanced UX:
- `sin`: Server information
- `ask`: User prompt messages
- Buttons 1-9: User response options

These are optional protocol features, currently not implemented.

#### Options

**Option A: Implement Ask/Button Mechanism**
- Allows server to prompt user for decisions
- Example: "Create new account?" [Yes] [No]
- **Effort:** Medium (2-3 days)

**Option B: Defer Indefinitely**
- Current implementation sufficient for basic authentication
- **Effort:** None

#### Recommendation
**Option B (Defer)** - Not required for core authentication flow.

**Revisit if:** User feedback indicates need for enhanced prompts.

---

## ARCHITECTURAL DECISIONS

### DECISION-009: Monitoring and Metrics Framework

**Priority:** 🟠 HIGH
**Category:** Operations
**Deadline:** Before Stage 3 (Week 8)
**Related TODO:** CONSOLIDATED_TODO.md Item #22

#### Issue
No monitoring/metrics currently implemented. Need to choose observability framework.

#### Options

**Option A: Prometheus + Grafana**
- Industry standard
- Rich ecosystem
- Pull-based metrics
- **Libraries:** `github.com/prometheus/client_golang`
- **Effort:** Medium (3-4 days)

**Option B: OpenTelemetry**
- Modern, vendor-neutral
- Unified metrics, traces, logs
- More future-proof
- **Libraries:** `go.opentelemetry.io/otel`
- **Effort:** Medium-High (4-5 days)

**Option C: Cloud-Native (CloudWatch, Stackdriver, etc.)**
- If deploying to AWS/GCP/Azure
- Integrated with cloud platform
- **Effort:** Low-Medium (varies by platform)

**Option D: Defer**
- Implement later
- **Risk:** Cannot troubleshoot production issues

#### Recommendation
**Option A (Prometheus)** for self-hosted deployments, **Option C** for cloud deployments.

**Rationale:**
- Prometheus is battle-tested and well-documented
- Low vendor lock-in
- Easy to add Grafana dashboards

#### Impact
- **Operations:** Enables troubleshooting, capacity planning
- **Cost:** Open-source (free), requires hosting
- **Complexity:** Low-Medium

#### Request for Advice
**Questions for stakeholders:**
1. What is the deployment environment? (self-hosted, AWS, GCP, Azure, other)
2. Is there existing monitoring infrastructure?
3. Are there compliance requirements for metrics retention?

**Please provide guidance on:**
- Preferred option
- Deployment environment

---

## INTEGRATION DECISIONS

### DECISION-010: Gogs vs Gitea for Git Hosting

**Priority:** 🟢 LOW
**Category:** Integration
**Deadline:** Before Stage 5 (if pursued)
**Related:** DEPENDENCY_UPGRADE_PLAN.md Stage 5

#### Issue
If Git hosting integration is needed, which platform should be used?

#### Background
- Gogs: Lightweight, minimal features, **NO OAuth2/OIDC support**
- Gitea: Gogs fork, more features, **HAS OAuth2/OIDC support**

#### Options

**Option A: Use Gogs (as originally mentioned)**
- Requires custom authentication plugin OR reverse proxy auth
- Higher integration complexity
- **Effort:** High (custom auth integration)

**Option B: Use Gitea (Recommended Alternative)**
- Native OAuth2/OIDC support
- Can integrate via Authentik (Stage 4)
- Standard OAuth2 flow
- **Effort:** Medium (standard OAuth2 integration)

**Option C: Neither (Skip Git Hosting)**
- Use existing Git hosting (GitHub, GitLab, Bitbucket)
- No self-hosted Git required
- **Effort:** None

#### Recommendation
**Option B (Gitea)** if Git hosting is needed, otherwise **Option C (skip)**.

**Rationale:**
- Gitea has native OAuth2 support (Gogs does not)
- Easier integration with SQRL via Authentik OAuth2 bridge
- More actively maintained than Gogs

#### Impact
- **Complexity:** Gitea easier to integrate
- **Features:** Gitea has more features
- **Maintenance:** Gitea more actively maintained

#### Request for Advice
**Questions for stakeholders:**
1. Is self-hosted Git hosting actually required?
2. Can existing Git hosting (GitHub, GitLab) be used instead?
3. If self-hosted is needed, is OAuth2/OIDC integration acceptable?

**Please provide guidance on:**
- Whether Git hosting integration is needed at all
- If needed, preference between Gogs and Gitea

---

## SECURITY DECISIONS

### DECISION-011: Secure Memory Clearing Aggressiveness

**Priority:** 🟡 MEDIUM
**Category:** Security
**Deadline:** Before Stage 1 completion (Week 1)

#### Issue
Current `secure_clear.go` implements basic memory clearing. Should we implement more aggressive clearing?

#### Options

**Option A: Current Implementation (Single-Pass Zero)**
```go
func ClearBytes(b []byte) {
    for i := range b {
        b[i] = 0
    }
    runtime.KeepAlive(b)
}
```
- **Pros:** Fast, prevents compiler optimization
- **Cons:** May not prevent all memory recovery techniques

**Option B: Multi-Pass Clearing (DoD 5220.22-M Standard)**
```go
func ClearBytesSecure(b []byte) {
    for pass := 0; pass < 3; pass++ {
        for i := range b {
            b[i] = byte(rand.Intn(256))
        }
    }
    for i := range b {
        b[i] = 0
    }
    runtime.KeepAlive(b)
}
```
- **Pros:** More thorough, meets security standards
- **Cons:** 3-4x slower, may impact performance

**Option C: Platform-Specific Secure Zero**
```go
// Unix: mlock + explicit_bzero / memset_s
// Windows: SecureZeroMemory
```
- **Pros:** OS-level guarantees
- **Cons:** Platform-dependent, complex implementation

#### Recommendation
**Option A (Current Implementation)** for now, with Option B available for sensitive operations.

**Rationale:**
- Option A is sufficient for most attacks (prevents simple memory dumps)
- Option B can be used for most sensitive data (keys)
- Option C is overkill for Go's managed memory

#### Impact
- **Security:** Option C best, Option A good enough
- **Performance:** Option A fastest, Option B slower, Option C platform-dependent
- **Complexity:** Option A simplest, Option C most complex

#### Request for Advice
**Questions for stakeholders:**
1. What is the threat model? (physical access, memory dumps, advanced attacks)
2. Are there compliance requirements for memory clearing (e.g., FIPS 140-2)?
3. Is performance impact of multi-pass clearing acceptable?

**Please provide guidance on:**
- Preferred option for default clearing
- Whether to provide both fast and secure clearing options

---

## SUMMARY TABLE

| ID | Priority | Category | Decision | Deadline | Impact |
|----|----------|----------|----------|----------|--------|
| **DECISION-001** | 🔴 CRITICAL | Architecture | Production Storage Backend | Week 6 | High |
| **DECISION-002** | 🔴 CRITICAL | Quality | Test Coverage Priority | Week 2 | High |
| **DECISION-003** | 🟠 HIGH | Security | Rate Limiting Strategy | Week 5 | Medium |
| **DECISION-004** | 🟠 HIGH | Security | Signature Failure Handling | Week 4 | Medium |
| **DECISION-005** | 🟡 MEDIUM | Data Model | Pidk Storage | Week 8 | Low |
| **DECISION-006** | 🟡 MEDIUM | Protocol | PreviousIDMatch Flag | Week 4 | Low |
| **DECISION-007** | 🟢 LOW | Protocol | Version Range Support | Future | Very Low |
| **DECISION-008** | 🟢 LOW | Features | Ask/Button Parameters | Future | Very Low |
| **DECISION-009** | 🟠 HIGH | Operations | Monitoring Framework | Week 8 | High |
| **DECISION-010** | 🟢 LOW | Integration | Gogs vs Gitea | Stage 5 | Low |
| **DECISION-011** | 🟡 MEDIUM | Security | Memory Clearing | Week 1 | Medium |

---

## RECOMMENDED DECISION TIMELINE

### Immediate (Week 1):
- **DECISION-011:** Secure Memory Clearing

### Before Stage 2 (Week 2):
- **DECISION-002:** Test Coverage Priority

### During Stage 2 (Weeks 3-5):
- **DECISION-004:** Signature Failure Handling
- **DECISION-006:** PreviousIDMatch Flag
- **DECISION-003:** Rate Limiting Strategy

### Before Stage 3 (Weeks 6-8):
- **DECISION-001:** Production Storage Backend ← **MOST CRITICAL**
- **DECISION-005:** Pidk Storage
- **DECISION-009:** Monitoring Framework

### Future:
- **DECISION-007, 008, 010:** Low priority items

---

## HOW TO RESPOND

Please provide decisions via:

1. **GitHub Issue:** Create issue for each DECISION-XXX
2. **Pull Request:** Update this document with decisions
3. **Email/Chat:** Informal feedback on recommendations
4. **Meeting:** Discuss complex decisions (DECISION-001, 002, 009)

**For each decision, please specify:**
- **Choice:** Which option (A, B, C, D, etc.)
- **Rationale:** Why this option was chosen
- **Timeline:** Any modifications to suggested deadline
- **Constraints:** Budget, timeline, technical constraints
- **Follow-up:** Any additional information needed

---

## CONFLICTS AND GAPS IDENTIFIED

### Conflicts

**CONFLICT-1: Storage Backend vs Timeline**
- **Issue:** Redis+PostgreSQL (fast) vs etcd (better architecture) vs timeline pressure
- **Impact:** May need to deploy with MapHoard initially, refactor later
- **Risk:** Technical debt

**CONFLICT-2: Test Coverage vs Feature Development**
- **Issue:** 3 weeks for testing vs pressure to add features
- **Impact:** May need to prioritize either coverage or features
- **Risk:** Quality vs velocity tradeoff

**CONFLICT-3: Security vs Usability**
- **Issue:** Aggressive signature failure handling vs user experience
- **Impact:** Balance security (remove identity) vs UX (allow retries)
- **Risk:** Either security vulnerabilities or locked-out users

### Gaps

**GAP-1: Audit Logging**
- **Issue:** No audit logging implementation
- **Impact:** Cannot track security events
- **Recommendation:** Add to Stage 3 or 4

**GAP-2: User Notifications**
- **Issue:** No mechanism to notify users of security events
- **Impact:** Users unaware of disabled accounts, security issues
- **Recommendation:** Design notification system (future work)

**GAP-3: Compliance Documentation**
- **Issue:** No GDPR/CCPA/compliance documentation
- **Impact:** May not meet regulatory requirements
- **Recommendation:** Add compliance review post-MVP

**GAP-4: Disaster Recovery**
- **Issue:** No backup/restore procedures documented
- **Impact:** Risk of data loss
- **Recommendation:** Document in Stage 3 (with etcd or database)

**GAP-5: Multi-Tenancy**
- **Issue:** No support for multiple domains/tenants
- **Impact:** Each deployment serves one domain only
- **Recommendation:** Future enhancement if needed

---

**Document Version:** 1.0
**Last Updated:** November 18, 2025
**Status:** Awaiting Stakeholder Input
**Next Review:** After decisions provided
