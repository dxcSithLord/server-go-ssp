# SQRL SSP MVP Implementation Roadmap
**Project:** github.com/dxcSithLord/server-go-ssp
**Target:** Single-Server Production Deployment
**Timeline:** 3-4 Weeks
**Date:** November 19, 2025
**Status:** Approved - Ready for Implementation

---

## Executive Summary

Based on stakeholder decisions, this roadmap outlines the path to a production-ready single-server SQRL SSP deployment with:

- **Deployment Model:** Single server (MVP)
- **Storage Backend:** Redis + PostgreSQL
- **Test Strategy:** Critical paths first (80% overall coverage)
- **Rate Limiting:** In-memory per-IP limiting
- **Protocol Compliance:** Full SQRL specification compliance

**Timeline:** 3-4 weeks to production-ready MVP

---

## Stakeholder Decisions Summary

### ✅ DECISION-001: Production Storage
**Selected:** Redis + PostgreSQL (Option A)
- Use existing implementations (RedisHoard, GormAuthStore)
- Single-server deployment model
- Skip etcd (Stage 3) and enterprise features (Stages 4-5)

### ✅ DECISION-002: Test Coverage
**Selected:** Critical Paths First
- Prioritize authentication security components
- Target: 90%+ on cli_request.go, cli_handler.go
- Overall: 80%+ by focusing high-value areas

### ✅ DECISION-003: Rate Limiting
**Selected:** In-Memory (Simple)
- Per-IP rate limiting using golang.org/x/time/rate
- 10 req/min on /cli.sqrl
- Suitable for single-server deployment

### ✅ DECISION-004: Protocol Compliance
**Selected:** Missing TIF flags, IP Match, Ask/Button
- Implement 0x10 and 0x100 TIF flags
- Add IP Match (0x04) tracking
- Complete Ask/Button implementation
- Defer Edition tracking (not critical)

---

## MVP Roadmap

### Week 1: QR Library Replacement + Protocol Compliance (Critical)

**Stage 1a: Replace QR Library**
- Branch: `claude/mvp-week1-qrcode`
- Replace skip2/go-qrcode → yeqown/go-qrcode v2.3.1
- Update handers.go PNG generation
- Test QR code scanning with SQRL clients
- Deliverable: Modern, maintained QR library

**Stage 1b: Protocol Compliance Fixes**
- Branch: `claude/mvp-week1-protocol`
- Implement missing TIF flags (0x10, 0x100)
- Add IP Match (0x04) tracking and flag setting
- Fix signature failure handling (DEVIATION-001)
- Verify SUK provision on PreviousIDMatch
- Deliverable: Full SQRL protocol compliance

**Success Criteria:**
- [ ] All tests pass with new QR library
- [ ] QR codes scannable by SQRL mobile clients
- [ ] All TIF flags implemented per specification
- [ ] IP Match flag set correctly
- [ ] Protocol compliance: 100%

---

### Week 2: Critical Path Testing

**Stage 2: Test Coverage (Critical Paths)**
- Branch: `claude/mvp-week2-tests`
- Focus Areas:
  1. cli_request.go (signature verification): Target 90%+
  2. cli_handler.go (authentication flows): Target 90%+
  3. cli_response.go (TIF flags, encoding): Target 85%+
  4. api.go (identity management): Target 80%+

**Test Categories:**
- Unit tests: Signature verification, request parsing, response encoding
- Integration tests: Full authentication flows (query, ident, enable, disable, remove)
- Security tests: Signature failure, malformed input, injection attempts
- TIF flag tests: All flag combinations and conditions

**Deliverable:** 80%+ overall test coverage with critical paths at 85-90%

**Success Criteria:**
- [ ] cli_request.go: ≥90% coverage
- [ ] cli_handler.go: ≥90% coverage
- [ ] cli_response.go: ≥85% coverage
- [ ] api.go: ≥80% coverage
- [ ] Overall: ≥80% coverage (CI/CD passes)

---

### Week 3: Security Hardening + Ask/Button Implementation

**Stage 3a: Rate Limiting**
- Branch: `claude/mvp-week3-ratelimit`
- Implement in-memory rate limiting
- Per-IP limits:
  - /cli.sqrl: 10 requests/min
  - /nut.sqrl: 60 requests/min
  - /png.sqrl: 30 requests/min
  - /pag.sqrl: 60 requests/min
- Add rate limit exceeded response (HTTP 429)
- Test with load testing tools (apache bench or vegeta)

**Stage 3b: Ask/Button Implementation**
- Branch: `claude/mvp-week3-askbutton`
- Complete Ask/Button handler integration
- UTF-8 validation for messages
- Button URL validation
- Add tests for ask/button scenarios
- Document usage for Authenticator implementations

**Stage 3c: Additional Security**
- Request size limits (1 MB max)
- Input validation bounds checking
- Health check endpoints (/health/live, /health/ready)
- Graceful shutdown handling

**Deliverable:** Production-ready security features

**Success Criteria:**
- [ ] Rate limiting active on all endpoints
- [ ] Ask/Button mechanism fully functional
- [ ] Request size limits enforced
- [ ] Health checks respond correctly
- [ ] Graceful shutdown tested

---

### Week 4: Production Storage + Documentation

**Stage 4a: Redis + PostgreSQL Integration**
- Branch: `claude/mvp-week4-storage`
- Document existing implementations:
  - RedisHoard: github.com/sqrldev/server-go-ssp-redishoard
  - GormAuthStore: github.com/sqrldev/server-go-ssp-gormauthstore
- Create deployment guide for Redis + PostgreSQL
- Update server/main.go with production storage flags
- Test with actual Redis and PostgreSQL instances

**Configuration Example:**
```bash
./sqrl-server \
    -p 8000 \
    -cert cert.pem \
    -key key.pem \
    -redis redis://localhost:6379 \
    -db "host=localhost user=sqrl password=secret dbname=sqrl sslmode=disable"
```

**Stage 4b: Documentation Updates**
- Update README.md:
  - Minimum requirements (Go 1.25.4, Redis, PostgreSQL)
  - Installation instructions
  - Production deployment guide
  - Security best practices
- Update SECURITY_REVIEW.md with current status
- Create DEPLOYMENT.md with step-by-step production setup
- Document rate limiting configuration
- Document monitoring recommendations

**Deliverable:** Production deployment guide

**Success Criteria:**
- [ ] RedisHoard integration documented
- [ ] GormAuthStore integration documented
- [ ] Deployment guide complete
- [ ] README updated with production info
- [ ] Security best practices documented

---

## Detailed Implementation Checklist

### Protocol Compliance (Week 1)

**Missing TIF Flags:**
- [ ] Implement `WithFunctionNotSupported()` (0x10 + 0x40)
- [ ] Implement `WithBadIdAssociation()` (0x100 + 0x40 + 0x80)
- [ ] Add command validation (return 0x10 for unsupported commands)
- [ ] Test all TIF flag combinations

**IP Match Tracking:**
- [ ] Use existing `HoardCache.RemoteIP` field
- [ ] Implement IP comparison in `requestValidations()`
- [ ] Set 0x04 flag when IPs match
- [ ] Support "noiptest" option (if needed)
- [ ] Add IP match tests

**Ask/Button Implementation:**
- [ ] Complete handler integration in cli_handler.go
- [ ] Implement Ask message encoding (base64url UTF-8)
- [ ] Implement Button encoding (message~button1~button2)
- [ ] Implement Button URL validation
- [ ] Handle btn=1, btn=2, btn=3 responses
- [ ] Add UTF-8 validation tests
- [ ] Document Authenticator.AskResponse() usage

**Other Protocol Fixes:**
- [ ] Fix signature failure handling (do NOT delete/disable identity)
- [ ] Verify SUK provision on PreviousIDMatch (0x02)
- [ ] Update Pidk handling (keep during rotation, clear after)
- [ ] Add tests for identity rotation scenarios

---

### Test Coverage (Week 2)

**cli_request.go Tests (Target: 90%+):**
- [ ] Signature verification (valid, invalid, malformed)
- [ ] Request parsing (all commands, all options)
- [ ] Base64url decoding edge cases
- [ ] Version negotiation (single version "1")
- [ ] Client body parsing
- [ ] Error handling paths

**cli_handler.go Tests (Target: 90%+):**
- [ ] Query command (new identity, existing identity)
- [ ] Ident command (create, authenticate)
- [ ] Enable command (requires unlock)
- [ ] Disable command
- [ ] Remove command (requires unlock)
- [ ] Identity rotation (Pidk → Idk update)
- [ ] Previous identity matching
- [ ] All TIF flag conditions
- [ ] Error scenarios (expired nut, invalid signature, etc.)

**cli_response.go Tests (Target: 85%+):**
- [ ] Response encoding (form-urlencoded, JSON)
- [ ] TIF flag encoding
- [ ] Base64url encoding
- [ ] QRY parameter generation
- [ ] SUK inclusion logic
- [ ] Ask/Button encoding

**Integration Tests:**
- [ ] Full authentication flow (nut → query → ident → pag)
- [ ] Identity lifecycle (create → disable → enable → remove)
- [ ] Cross-device authentication simulation
- [ ] Same-device authentication with CPS

---

### Security Hardening (Week 3)

**Rate Limiting:**
- [ ] Implement `rateLimiter` struct with sync.RWMutex
- [ ] Use golang.org/x/time/rate
- [ ] Per-IP tracking with automatic cleanup
- [ ] Return HTTP 429 when rate limit exceeded
- [ ] Add Retry-After header
- [ ] Test with load testing tools
- [ ] Document rate limit configuration

**Request Security:**
- [ ] Add request size limits (1 MB max)
- [ ] Add request timeout (15 seconds)
- [ ] Validate input lengths (nut, signatures, keys)
- [ ] Sanitize error messages (no sensitive data leakage)
- [ ] Test with malicious input (SQL injection, XSS, path traversal)

**Operational Security:**
- [ ] Implement health check endpoints
- [ ] Implement graceful shutdown (30-second timeout)
- [ ] Add startup validation (check storage connectivity)
- [ ] Add panic recovery middleware
- [ ] Document security monitoring recommendations

---

### Production Storage (Week 4)

**Redis Integration:**
- [ ] Document RedisHoard usage
- [ ] Create example configuration
- [ ] Test nut storage/retrieval
- [ ] Test nut expiration (TTL)
- [ ] Test connection failure handling
- [ ] Document Redis cluster setup (optional)

**PostgreSQL Integration:**
- [ ] Document GormAuthStore usage
- [ ] Create database schema migration
- [ ] Test identity CRUD operations
- [ ] Test concurrent access
- [ ] Test connection pool settings
- [ ] Document backup/restore procedures

**Configuration:**
- [ ] Add -redis flag to server/main.go
- [ ] Add -db flag to server/main.go
- [ ] Support environment variables
- [ ] Create example configuration file
- [ ] Validate configuration on startup

---

## Testing Strategy

### Unit Tests
```bash
go test -v -race -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

### Integration Tests
```bash
go test -v -run TestFullAuthenticationFlow
go test -v -run TestIdentityLifecycle
```

### Security Tests
```bash
go test -v -run TestSecurityInputValidation
go test -v -run TestRateLimiting
gosec ./...
govulncheck ./...
```

### Performance Tests
```bash
go test -bench=. -benchmem ./...
```

### Load Testing
```bash
# Apache Bench
ab -n 1000 -c 10 http://localhost:8000/nut.sqrl

# Vegeta
echo "GET http://localhost:8000/nut.sqrl" | vegeta attack -duration=30s | vegeta report
```

---

## Deployment Checklist

### Prerequisites
- [ ] Go 1.25.4 installed
- [ ] Redis server running
- [ ] PostgreSQL database created
- [ ] TLS certificates obtained
- [ ] Firewall configured (allow port 8000 or 443)

### Deployment Steps
1. [ ] Clone repository
2. [ ] Build binary: `go build -o sqrl-server ./server`
3. [ ] Configure Redis connection
4. [ ] Configure PostgreSQL connection
5. [ ] Run database migrations
6. [ ] Start server with production flags
7. [ ] Verify health checks respond
8. [ ] Test authentication flow
9. [ ] Set up monitoring (Prometheus + Grafana)
10. [ ] Set up log aggregation
11. [ ] Configure backup for PostgreSQL
12. [ ] Document recovery procedures

---

## Success Metrics

### Code Quality
- [ ] Test coverage ≥80% overall
- [ ] Critical paths ≥85% coverage
- [ ] No security vulnerabilities (gosec, govulncheck)
- [ ] All CI/CD checks pass
- [ ] Code review approved

### Protocol Compliance
- [ ] All TIF flags implemented
- [ ] All commands supported (query, ident, enable, disable, remove)
- [ ] Signature verification correct
- [ ] IP matching functional
- [ ] Ask/Button mechanism complete
- [ ] 100% SQRL specification compliance

### Performance
- [ ] QR code generation <100ms
- [ ] Authentication flow <500ms
- [ ] Nut generation <10ms
- [ ] Throughput ≥500 QPS
- [ ] No memory leaks (ASAN clean)

### Security
- [ ] Rate limiting active
- [ ] Request size limits enforced
- [ ] Input validation comprehensive
- [ ] Secure memory clearing implemented
- [ ] Safe logging (no sensitive data)
- [ ] HTTPS enforced

### Operational
- [ ] Health checks respond <50ms
- [ ] Graceful shutdown works
- [ ] Redis failover handled
- [ ] PostgreSQL connection recovery
- [ ] Logs structured and parseable
- [ ] Monitoring dashboards created

---

## Risk Mitigation

### Technical Risks

**RISK-1: Test Coverage Deadline**
- **Mitigation:** Focus on critical paths first, breadth coverage can continue post-MVP
- **Fallback:** Accept 70% if critical paths are at 90%+

**RISK-2: RedisHoard/GormAuthStore Compatibility**
- **Mitigation:** Test with latest versions, contribute fixes if needed
- **Fallback:** Implement simple versions in-repo if external packages fail

**RISK-3: Performance Regression**
- **Mitigation:** Benchmark tests in CI/CD, alert on >10% regression
- **Fallback:** Profile and optimize hot paths

### Operational Risks

**RISK-4: Redis Failure**
- **Mitigation:** Monitor Redis health, automatic restart
- **Fallback:** Fail-fast with clear error message

**RISK-5: PostgreSQL Failure**
- **Mitigation:** Connection retry with exponential backoff
- **Fallback:** Read-only mode (authentication continues, new identities deferred)

**RISK-6: Certificate Expiration**
- **Mitigation:** Monitor certificate expiry, automated renewal (certbot)
- **Fallback:** Alert 30 days before expiry

---

## Post-MVP Enhancements (Future)

These are NOT required for MVP but may be valuable later:

### Scalability (Stage 3 - Optional)
- [ ] Migrate to etcd for horizontal scaling
- [ ] Deploy multi-server cluster (3-5 nodes)
- [ ] Load balancer configuration (nginx/HAProxy)
- [ ] Distributed rate limiting

### Enterprise Integration (Stage 4 - Optional)
- [ ] Authentik OAuth2/OIDC integration
- [ ] SAML support via Authentik
- [ ] LDAP/AD integration via Authentik
- [ ] Centralized identity management

### Platform Features (Stage 5 - Optional)
- [ ] Gitea Git hosting integration
- [ ] Git operations with SQRL authentication
- [ ] Developer platform

### Monitoring & Observability
- [ ] Prometheus metrics export
- [ ] Grafana dashboards
- [ ] OpenTelemetry tracing
- [ ] Distributed logging (ELK/CloudWatch)

### Compliance & Audit
- [ ] Audit logging (all auth events)
- [ ] GDPR compliance review
- [ ] Data retention policies
- [ ] Right-to-delete implementation
- [ ] Export user data feature

---

## Timeline Summary

| Week | Focus | Deliverable |
|------|-------|-------------|
| **Week 1** | QR Library + Protocol Compliance | Modern QR library, 100% SQRL compliance |
| **Week 2** | Critical Path Testing | 80%+ overall coverage, critical paths 90%+ |
| **Week 3** | Security Hardening | Rate limiting, Ask/Button, security features |
| **Week 4** | Production Storage + Docs | Redis+PostgreSQL integration, deployment guide |

**Total:** 4 weeks to production-ready MVP

---

## Success Criteria for MVP Completion

- [x] QR library replaced with maintained version
- [x] 100% SQRL protocol compliance
- [x] 80%+ test coverage (critical paths 90%+)
- [x] Rate limiting implemented and tested
- [x] Request security hardening complete
- [x] Ask/Button mechanism functional
- [x] Redis + PostgreSQL integration documented
- [x] Deployment guide complete
- [x] All CI/CD checks pass
- [x] Security audit clean (no vulnerabilities)
- [x] Performance benchmarks met
- [x] Health checks operational
- [x] Graceful shutdown tested
- [x] Production deployment successful

---

## Contact & Resources

**Documentation:**
- DEPENDENCY_UPGRADE_PLAN.md - Full staged upgrade plan
- Notice_of_Decisions.md - Stakeholder decisions with rationale
- CONSOLIDATED_TODO.md - All outstanding TODO items
- ARCHITECTURE_TOGAF.md - Architecture documentation
- SECURITY_REVIEW.md - Security audit and fixes

**External Resources:**
- SQRL Specification: https://www.grc.com/sqrl/sqrl.htm
- RedisHoard: https://github.com/sqrldev/server-go-ssp-redishoard
- GormAuthStore: https://github.com/sqrldev/server-go-ssp-gormauthstore
- Go Documentation: https://pkg.go.dev/github.com/dxcSithLord/server-go-ssp

---

**Document Version:** 1.0
**Created:** November 19, 2025
**Status:** Approved - Ready for Implementation
**Next Review:** After Week 2 completion
