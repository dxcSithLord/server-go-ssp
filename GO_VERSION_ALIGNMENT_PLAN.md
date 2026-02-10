# Go Version Alignment Plan

**Date:** February 8, 2026
**Scope:** server-go-ssp + server-go-ssp-gormauthstore
**Session Reference:** https://claude.ai/code/session_01BbgaGXFqQXGpvrzSxpNN69 (gormauthstore interface work)
**Status:** Phase A COMPLETE

---

## 1. Current State Analysis

### Go Versions in Use

| Component | go.mod `go` directive | go.mod `toolchain` | Local Go binary | CI version |
|-----------|----------------------|--------------------|-----------------|------------|
| **server-go-ssp** | `go 1.25.0` | `go1.25.7` | go1.25.1 (also go1.24.7) | `1.25` (setup-go resolves latest) |
| **server-go-ssp-gormauthstore** | `go 1.24.0` | `go1.24.7` | go1.24.7 | `1.24.x` |

### Latest Stable Releases (as of 2026-02-08)

- **Go 1.25.7** (released 2026-02-04) -- current supported branch
- **Go 1.24.13** (released 2026-02-04) -- current supported branch (EOL when 1.26 ships)
- **Go 1.26** -- imminent (February 2026)

### Network Constraint

This environment cannot download Go toolchains or modules from the internet. Go 1.25.1 is the highest available locally. The `toolchain go1.25.7` directive allows CI/CD (which has internet) to use the latest patch release.

### Dependency Relationship

```
server-go-ssp-gormauthstore
  └── depends on: github.com/dxcSithLord/server-go-ssp v0.0.0-20260202110616-66529f78b7f1
                   (pinned to master commit 66529f7)
```

gormauthstore imports the `ssp.AuthStore` interface and `ssp.SqrlIdentity` struct from server-go-ssp. The `go` directive in server-go-ssp sets the **minimum Go version** that consumers must use.

---

## 2. Decision Alignment Between Repos

### Cross-Reference: server-go-ssp vs gormauthstore Decisions

| server-go-ssp Decision | gormauthstore Decision | Alignment Status |
|------------------------|----------------------|------------------|
| DECISION-001: Redis + PostgreSQL storage | DP-002: Interface coordination | ALIGNED -- gormauthstore provides GormAuthStore for PostgreSQL |
| DECISION-002: Critical path testing | (internal) 98.8% coverage achieved | gormauthstore AHEAD -- server-go-ssp at 29.4% |
| DECISION-003: In-memory rate limiting | N/A | server-go-ssp only |
| DECISION-004: Signature failure handling | N/A | server-go-ssp protocol compliance |
| N/A | DP-001: context.Context support | IMPLEMENTED in gormauthstore; server-go-ssp AuthStore interface unchanged |
| N/A | DP-003: Field-level encryption | DEFERRED to v1.1.0 (not context.Context!) |
| N/A | DP-004: Goauthentik integration | RESOLVED -- no action needed |

### CORRECTION: DP-003 Is NOT context.Context

Previous session notes incorrectly identified gormauthstore DP-003 as "context.Context API design". In fact:
- **DP-001** = context.Context support -- **IMPLEMENTED** (WithContext variants added, backward compatible)
- **DP-003** = field-level encryption -- **DEFERRED** to v1.1.0

### gormauthstore Progress (Corrected)

Previous estimate: 73% (32/44 tasks). **Actual: 91% (48/53 tasks)**

| Phase | Status | Tasks |
|-------|--------|-------|
| Phase 1: GORM v2 Migration | COMPLETE | 20/20 (1 deferred) |
| Phase 2: Security & Testing | COMPLETE | 14/14 |
| Phase 3.1: Production Hardening | COMPLETE | 4/4 |
| Phase 3.2: Release v1.0.0 | IN PROGRESS | 2/6 done |
| Documentation | COMPLETE | 9/9 |

**Remaining tasks (4):**
- TASK-041: Tag v1.0.0
- TASK-042: GitHub Release page
- TASK-043: Revert module path to sqrldev
- TASK-044: Submit to pkg.go.dev

---

## 3. Go Version Upgrade: Phase A Complete

### What Was Done

1. Updated `go.mod`:
   ```
   go 1.25.0
   toolchain go1.25.7
   ```

2. Ran `go mod tidy` -- dependencies unchanged
3. Ran `go vet ./...` -- clean
4. Ran `go build ./...` -- clean
5. Ran `go test -race ./...` -- **all 89 tests pass**
6. Pre-existing test failure `TestPngEndpoint_InvalidNut` now **passes** on Go 1.25.1

### CI Compatibility

The CI workflow (`ci.yml`) already specifies `go-version: '1.25'` throughout. The `setup-go` action resolves this to the latest Go 1.25.x available, which will be 1.25.7.

---

## 4. Remaining Phases

### Phase B: Publish server-go-ssp

1. Merge this branch's changes to master
2. Note the new commit hash for gormauthstore to reference

### Phase C: Upgrade gormauthstore -- IN PROGRESS

_(Being done in the gormauthstore repo, separate session)_

1. Update go.mod: `go 1.25.0` / `toolchain go1.25.7`
2. Update server-go-ssp dependency to new commit
3. Run `go mod tidy`
4. Run full test suite (100 tests, 10 benchmarks)
5. Proceed with TASK-041 (tag v1.0.0)

**Status (2026-02-08):** gormauthstore is being upgraded to Go 1.25.

### Phase D: Verify Integration

1. Use gormauthstore's commented-out replace directive to test against local server-go-ssp
2. Verify AuthStore interface compatibility
3. Verify SqrlIdentity struct compatibility

---

## 5. Risk Assessment (Updated)

| Risk | Impact | Likelihood | Status |
|------|--------|------------|--------|
| Go 1.25 introduces breaking changes | HIGH | LOW | **MITIGATED** -- all 89 tests pass |
| Interface mismatch after upgrade | HIGH | LOW | **MITIGATED** -- gormauthstore DP-001 maintains backward compat |
| golang.org/x/image incompatibility | MEDIUM | LOW | Unchanged at v0.18.0 -- works |
| CI pipeline fails on 1.25 | MEDIUM | LOW | CI already targets 1.25 |
| Go 1.26 releases during this work | LOW | HIGH | Not a problem -- 1.25 remains supported |
| Network blocks toolchain download | MEDIUM | CONFIRMED | **MITIGATED** -- go1.25.1 available locally |

---

## Appendix: Go Toolchain Behavior

With go.mod containing:
```
go 1.25.0
toolchain go1.25.7
```

- `go 1.25.0`: minimum language version; consumers need Go >= 1.25.0
- `toolchain go1.25.7`: preferred toolchain; Go will auto-download if needed
- `GOTOOLCHAIN=go1.25.1`: forces use of local 1.25.1 (for offline environments)
- CI with `setup-go: '1.25'` resolves to latest 1.25.x (currently 1.25.7)
- gormauthstore will need to upgrade from `go 1.24.0` to at least `go 1.25.0` to consume this module

---

**Document Owner:** Project Maintainer
**Phase A:** COMPLETE (2026-02-08)
**Next:** Phase B (merge to master), Phase C (upgrade gormauthstore)
