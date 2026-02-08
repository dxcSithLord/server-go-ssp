# Go Version Alignment Plan

**Date:** February 7, 2026
**Scope:** server-go-ssp + server-go-ssp-gormauthstore
**Session Reference:** https://claude.ai/code/session_01BbgaGXFqQXGpvrzSxpNN69 (gormauthstore interface work)

---

## 1. Current State Analysis

### Go Versions in Use

| Component | go.mod `go` directive | go.mod `toolchain` | Local Go binary | Latest available |
|-----------|----------------------|--------------------|-----------------|------------------|
| **server-go-ssp** | `go 1.24` | _(none)_ | go1.24.7 | go1.24.13 / go1.25.7 |
| **server-go-ssp-gormauthstore** | `go 1.24.0` | `go1.24.7` | go1.24.7 | go1.24.13 / go1.25.7 |

### Latest Stable Releases (as of 2026-02-07)

- **Go 1.25.7** (released 2026-02-04) -- current supported branch
- **Go 1.24.13** (released 2026-02-04) -- current supported branch
- **Go 1.26** -- imminent (February 2026), which will end support for Go 1.24.x

### Documentation vs Reality Discrepancy

server-go-ssp documentation (PROJECT_ROADMAP.md, SECURITY_REVIEW.md, CURRENT_STATE_SUMMARY.md) claims Phase 1 completed a Go 1.25.0/1.25.4 upgrade. However:
- `go.mod` currently says `go 1.24` with **no toolchain directive**
- MVP_WEEK1_COMPLETION.md (Dec 17, 2025) explains: "Temporary go.mod adjustment: `go 1.24` (due to network blocker for 1.25.4 toolchain)"
- The Go 1.25 upgrade was effectively **reverted** and never re-applied

### Dependency Relationship

```
server-go-ssp-gormauthstore
  └── depends on: github.com/dxcSithLord/server-go-ssp v0.0.0-20260202110616-66529f78b7f1
                   (pinned to master commit 66529f7)
```

gormauthstore **imports** the `ssp.AuthStore` interface and `ssp.SqrlIdentity` struct from server-go-ssp. The go.mod `go` directive in server-go-ssp sets the **minimum Go version** that consumers must use.

---

## 2. Key Decisions Required

### DECISION-GO-001: Target Go Version

**Options:**

| Option | Version | Pros | Cons |
|--------|---------|------|------|
| A | Go 1.24.13 | Minimal change, both repos already on 1.24.x; patch-level security fixes | Will go out of support when Go 1.26 releases (imminent, Feb 2026) |
| B | Go 1.25.7 | Current latest stable; supported for ~6 more months; aligns with original Phase 1 plan | Requires testing both repos; may surface breaking changes |
| C | Wait for Go 1.26 | Latest features (Green Tea GC default); longest support window | Not yet released; delays alignment |

**Recommendation: Option B -- Go 1.25.7**

Rationale:
1. Go 1.24.x will lose support when Go 1.26 ships (imminent), making Option A a dead end
2. The original server-go-ssp plan already targeted Go 1.25; completing that work is overdue
3. Go 1.25.7 has the latest security patches (2026-02-04)
4. Both repos must align to the same minimum version since gormauthstore depends on server-go-ssp
5. Waiting for 1.26 adds unnecessary delay; upgrading 1.25 -> 1.26 later is straightforward

### DECISION-GO-002: Upgrade Sequence

The upgrade **must** follow this order due to the dependency chain:

```
Step 1: Upgrade server-go-ssp to Go 1.25.7
        (this is the upstream dependency)
             │
             ▼
Step 2: Tag/publish server-go-ssp at the new Go version
        (gormauthstore needs a resolved module version)
             │
             ▼
Step 3: Upgrade server-go-ssp-gormauthstore to Go 1.25.7
        Update its server-go-ssp dependency to the new tag/commit
             │
             ▼
Step 4: Verify integration tests pass across both repos
```

**Why this order matters:** If gormauthstore is upgraded first but server-go-ssp still declares `go 1.24`, the Go toolchain will use the maximum of all `go` directives in the dependency graph. But if server-go-ssp later upgrades to 1.25 and introduces 1.25-only features, gormauthstore would break unless it has already upgraded. Upgrading upstream first prevents this.

---

## 3. Outstanding Tasks and Blockers

### server-go-ssp (this repo)

| # | Task | Status | Notes |
|---|------|--------|-------|
| 1 | go.mod says `go 1.24` but docs claim 1.25 | Discrepancy | Must resolve -- either upgrade or correct docs |
| 2 | No toolchain directive in go.mod | Missing | Should add `toolchain go1.25.7` |
| 3 | CI workflow (ci.yml) references Go version matrix | Needs update | Verify CI uses correct Go version |
| 4 | Test coverage at 29.4% (target 80%) | In progress | Phase 2 work, independent of version upgrade |
| 5 | Pre-existing test failure: TestPngEndpoint_InvalidNut | Open | Should be fixed before or during upgrade |
| 6 | golang.org/x/image v0.18.0 -- check for updates | Check | May have newer version for Go 1.25 |
| 7 | MVP Week 2-4 work not started | Planned | Testing, security hardening, production storage |

### server-go-ssp-gormauthstore

| # | Task | Status | Notes |
|---|------|--------|-------|
| 1 | TASK-034: Merge and tag v0.3.0-rc1 | Pending | Blocks Phase 3 |
| 2 | Phase 3 not started (context support, deployment docs) | Planned | 6-8 hours remaining |
| 3 | Decision DP-003: context.Context API design | Open | Affects AuthStore interface |
| 4 | Module path reversion (dxcSithLord -> sqrldev) | Phase 3.2 | Before upstream submission |
| 5 | Integration test: confirm no GORM v1/v2 module conflict | Needed | gormauthstore uses GORM v2; server-go-ssp has no GORM dep |
| 6 | Dependency on server-go-ssp pinned to commit 66529f7 | Will need update | After server-go-ssp is upgraded and tagged |

### Interface Between Projects (from session_01BbgaGXFqQXGpvrzSxpNN69)

The gormauthstore session is working on the interface between the two projects. Key interface points:
- `ssp.AuthStore` interface: `FindIdentity`, `SaveIdentity`, `DeleteIdentity`
- `ssp.SqrlIdentity` struct shared between both repos
- gormauthstore wraps `SqrlIdentity` in an internal `identityRecord` for GORM v2 tags
- Decision DP-003 (context.Context support) could change the interface signatures

---

## 4. Alignment Plan

### Phase A: Prepare server-go-ssp (this repo)

**Branch:** `claude/review-tasks-docs-Wcsi2`

1. **Reconcile documentation with reality**
   - Update docs to acknowledge go.mod is at 1.24, not 1.25
   - Or proceed directly with the upgrade (preferred)

2. **Upgrade go.mod**
   ```
   go 1.25.7
   toolchain go1.25.7
   ```

3. **Run `go mod tidy`** to resolve any dependency changes

4. **Update golang.org/x/image** to latest compatible version

5. **Verify all tests pass**
   ```
   go test -race ./...
   go vet ./...
   ```

6. **Update CI workflow** (.github/workflows/ci.yml)
   - Set Go version matrix to include 1.25.x
   - Remove any 1.24-specific entries if present

7. **Fix TestPngEndpoint_InvalidNut** if feasible (pre-existing failure)

8. **Commit and push** to branch, create PR

### Phase B: Publish server-go-ssp

1. **Merge PR** from Phase A
2. **Note the commit hash** on master after merge -- gormauthstore will reference this

### Phase C: Upgrade server-go-ssp-gormauthstore

_(To be done in the gormauthstore repo, coordinated with session_01BbgaGXFqQXGpvrzSxpNN69)_

1. **Update go.mod**
   ```
   go 1.25.7
   toolchain go1.25.7
   ```

2. **Update server-go-ssp dependency** to the new commit from Phase B
   ```
   go get github.com/dxcSithLord/server-go-ssp@<new-commit-hash>
   ```

3. **Run `go mod tidy`**

4. **Run full test suite** (77 tests, 10 benchmarks)

5. **Verify no GORM v1/v2 module conflicts**

6. **Proceed with TASK-034** (tag v0.3.0-rc1)

### Phase D: Verify Integration

1. **Use gormauthstore's replace directive** (already commented out in go.mod) to test against local server-go-ssp
2. **Run integration tests** confirming the AuthStore interface works
3. **Verify SqrlIdentity struct compatibility** across both repos

---

## 5. Risk Assessment

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Go 1.25 introduces breaking changes | HIGH | LOW | Go has strong backward compatibility; 1.25 release notes document all changes |
| Interface mismatch after upgrade | HIGH | LOW | Both repos already compile on 1.24; upgrading go directive doesn't change interfaces |
| golang.org/x/image incompatibility | MEDIUM | LOW | Pin to known working version |
| CI pipeline fails on 1.25 | MEDIUM | MEDIUM | Test locally first; update CI matrix |
| Go 1.26 releases during this work | LOW | HIGH | Not a problem -- 1.25 will remain supported; upgrade to 1.26 can follow later |
| context.Context interface change (DP-003) | HIGH | MEDIUM | Decouple from Go version upgrade; treat as separate work stream |

---

## 6. Go 1.24 End-of-Support Timeline

```
Go 1.24.13 (current) ──── Go 1.26 releases (Feb 2026) ──── Go 1.24 EOL
     │                         │                                 │
  We are here             ~Days away                     No more patches
```

**Urgency: MODERATE-HIGH.** Go 1.24 will lose security patch support when Go 1.26 ships. Both repos should be on Go 1.25 before that happens to remain on a supported release.

---

## 7. Relationship to Existing Plans

### server-go-ssp Plans Updated

| Existing Plan | Impact |
|---------------|--------|
| PROJECT_ROADMAP.md Phase 1 | Mark Go 1.25 upgrade as **incomplete** (was incorrectly marked done) |
| PROJECT_ROADMAP.md Phase 2 | No change -- security/testing work is independent |
| MVP_ROADMAP.md Week 1 | QR library done; Go version is a separate work item |
| UPGRADE_GO_1_25.md | Remains the reference guide for the upgrade process |
| CONSOLIDATED_TODO.md Item 15 | Update README for Go 1.25 -- still needed |
| Notice_of_Decisions.md | No protocol impact from Go version change |

### gormauthstore Plans Updated

| Existing Plan | Impact |
|---------------|--------|
| Phase 2: TASK-034 (tag v0.3.0-rc1) | Should include Go 1.25 upgrade |
| Phase 3: Context support (DP-003) | Independent of Go version, but benefits from 1.25 stdlib improvements |
| Phase 3.2: Module path reversion | Schedule after Go version alignment |

---

## 8. Immediate Next Steps

1. **In this repo (server-go-ssp):** Attempt Go 1.25.7 upgrade
   - If local toolchain is only 1.24.7, the `toolchain` directive will instruct `go` to download 1.25.7 automatically (Go 1.21+ toolchain management)
   - If network is unavailable, document the blocker and set `go 1.24.13` as interim target

2. **Communicate to gormauthstore session:** Go version alignment is planned; hold off on Phase 3 module path reversion until both repos are on 1.25

3. **Update MEMORY.md** with Go version alignment decisions for future sessions

---

## Appendix: Go Toolchain Behavior

When `go.mod` contains:
```
go 1.25.7
toolchain go1.25.7
```

- The `go` directive sets the **minimum language version** required
- The `toolchain` directive specifies the **preferred toolchain** for building
- If a developer has Go 1.24.x installed, `go` will automatically download go1.25.7 (since Go 1.21+)
- Consumers (like gormauthstore) must have `go >= 1.25.7` or the toolchain auto-download kicks in

This means upgrading server-go-ssp's `go` directive effectively forces all consumers to use Go 1.25+.

---

**Document Owner:** Project Maintainer
**Review Required:** Before executing Phase A
**Coordination:** With gormauthstore session (session_01BbgaGXFqQXGpvrzSxpNN69)
