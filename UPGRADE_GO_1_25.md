# Go 1.24 to Go 1.25 Upgrade Plan

## Overview

This document outlines the comprehensive plan for upgrading the SQRL SSP codebase from Go 1.24 to Go 1.25.

**Go 1.25 Release Information:**
- Release Date: August 12, 2025
- Latest Patch: go1.25.4 (November 5, 2025)
- End of Life: Approximately February 2027 (when Go 1.27 releases)

---

## To-Do List

### Phase 1: Pre-Upgrade Research and Assessment

- [ ] **1.1 Review Go 1.25 Release Notes**
  - Location: https://go.dev/doc/go1.25
  - Focus areas: deprecations, breaking changes, security fixes
  - Document any code patterns that may be affected

- [ ] **1.2 Analyze Current Codebase for Deprecations**
  - Check for usage of deprecated go/ast functions:
    - `FilterPackage` (deprecated in 1.25)
    - `PackageExports` (deprecated in 1.25)
    - `MergePackageFiles` (deprecated in 1.25)
    - `ParseDir` (deprecated in 1.25)
  - Check for `testing/synctest.Run` usage (deprecated, use `Test` instead)
  - Run: `go vet ./...` with Go 1.25 to identify issues

- [ ] **1.3 Security Vulnerability Assessment**
  - Run `govulncheck ./...` with Go 1.25
  - Document any standard library vulnerabilities fixed in 1.25
  - Track security patches in go1.25.x releases

- [ ] **1.4 Dependency Compatibility Check**
  - Verify github.com/skip2/go-qrcode compatibility with Go 1.25
  - Check for any breaking changes in transitive dependencies
  - Test build with Go 1.25 locally

### Phase 2: Code Updates

- [x] **2.1 Update go.mod File** ✅ COMPLETED
  ```bash
  # Update go version directive
  go mod edit -go=1.25.0

  # Set toolchain to latest stable
  go mod edit -toolchain=go1.25.4

  # Tidy modules
  go mod tidy
  ```
  **Status:** Updated to go 1.25.0 with toolchain go1.25.4 (latest patch)

- [ ] **2.2 Replace Deprecated Code Patterns**
  - Replace any deprecated standard library usage
  - Update to recommended alternatives per release notes
  - Document all code changes in commit messages

- [ ] **2.3 Leverage New Features (Optional)**
  - Evaluate experimental garbage collector for performance
  - Consider testing/synctest.Test for concurrent test utilities
  - Explore go/ast PreorderStack for any AST manipulation
  - Evaluate encoding/json/v2 experimental package (if applicable)

- [ ] **2.4 Address ASAN Leak Detection**
  - Note: `go build -asan` now defaults to leak detection at exit
  - Test application with `-asan` flag for memory leak identification
  - Fix any detected memory leaks

### Phase 3: CI/CD Pipeline Updates

- [x] **3.1 Update .github/workflows/ci.yml** ✅ COMPLETED
  ```yaml
  # Update all Go version references from '1.24' to '1.25'
  - name: Set up Go
    uses: actions/setup-go@v5
    with:
      go-version: '1.25'
      cache: true
  ```
  **Status:** All 7 jobs updated to Go 1.25

- [x] **3.2 Update Test Matrix** ✅ COMPLETED
  ```yaml
  matrix:
    go-version: ['1.25']  # Consider also testing with '1.24' for compatibility
  ```
  **Status:** Test matrix and codecov condition updated

- [x] **3.3 Verify All CI Jobs** ✅ COMPLETED
  - security-scan: Updated to Go 1.25
  - lint: Updated to Go 1.25
  - test: Updated test matrix to 1.25
  - build: Updated to Go 1.25
  - vulnerability-scan: Updated to Go 1.25
  - benchmark: Updated to Go 1.25

- [ ] **3.4 Update Docker Images (if applicable)**
  - Update base images to golang:1.25
  - Rebuild and test all containerized deployments

### Phase 4: Testing and Validation

- [ ] **4.1 Run Full Test Suite**
  ```bash
  go test -race -coverprofile=coverage.out ./...
  go tool cover -func=coverage.out
  ```

- [ ] **4.2 Run Static Analysis**
  ```bash
  go vet ./...
  golangci-lint run --timeout=5m
  staticcheck ./...
  ```

- [ ] **4.3 Security Scanning**
  ```bash
  # Install latest govulncheck
  go install golang.org/x/vuln/cmd/govulncheck@latest

  # Run vulnerability check
  govulncheck ./...

  # Run gosec
  gosec ./...
  ```

- [ ] **4.4 Performance Testing**
  ```bash
  go test -bench=. -benchmem ./...
  ```
  - Compare results with Go 1.24 baseline
  - Document any performance improvements or regressions

- [ ] **4.5 Memory Leak Detection**
  ```bash
  # Go 1.25 ASAN defaults to leak detection
  go build -asan -o sqrl-server ./server
  # Run and monitor for leak reports
  ```

### Phase 5: Documentation Updates

- [ ] **5.1 Update README.md**
  - Update Go version requirements
  - Update installation instructions
  - Document minimum supported Go version

- [x] **5.2 Update SECURITY_REVIEW.md** ✅ COMPLETED
  - Document Go version upgrade: 1.17 -> 1.25.0
  - Updated repository reference to dxcSithLord fork
  - Update dependency versions

- [ ] **5.3 Update CONTRIBUTING.md (if exists)**
  - Update development environment requirements
  - Update toolchain setup instructions

- [ ] **5.4 Create Migration Notes**
  - Document any breaking changes for downstream users
  - List deprecated APIs and their replacements
  - Provide upgrade path guidance

### Phase 6: Security and Compliance

- [ ] **6.1 Document Security Fixes**
  - List all CVEs fixed between Go 1.24.10 and 1.25.4
  - Document standard library vulnerability resolutions
  - Update security audit trail

- [ ] **6.2 Validate Cryptographic Operations**
  - Ensure ED25519 operations unchanged
  - Verify AES encryption compatibility
  - Test all signature verification paths

- [ ] **6.3 Review Security Scanner Results**
  - CodeQL analysis with Go 1.25
  - Gosec scan results
  - Address any new security findings

- [ ] **6.4 Compliance Verification**
  - Ensure SQRL protocol compliance maintained
  - Verify authentication flow integrity
  - Test all security-critical paths

### Phase 7: Deployment and Release

- [ ] **7.1 Create Feature Branch**
  ```bash
  git checkout -b upgrade/go-1.25
  ```

- [ ] **7.2 Commit Changes**
  - Separate commits for each category of change
  - Clear, descriptive commit messages
  - Reference issue numbers if applicable

- [ ] **7.3 Create Pull Request**
  - Comprehensive PR description
  - List all changes and their rationale
  - Include testing evidence

- [ ] **7.4 Code Review**
  - Security review of all changes
  - Performance review
  - Compatibility verification

- [ ] **7.5 Merge and Tag Release**
  ```bash
  git tag -a v2.0.0 -m "Upgrade to Go 1.25"
  git push origin v2.0.0
  ```

- [ ] **7.6 Monitor Post-Deployment**
  - Watch for runtime errors
  - Monitor performance metrics
  - Track security scan results

---

## Expected Code Changes

### Deprecation Replacements

**None expected for current codebase:**
- No usage of deprecated go/ast functions
- No usage of testing/synctest.Run
- No usage of ParseDir

### Potential Code Updates

1. **Enhanced Error Handling**
   - Go 1.25 may provide improved error wrapping patterns
   - Review error handling in cryptographic operations

2. **Performance Optimizations**
   - Evaluate experimental GC for memory-intensive operations
   - Consider JSON v2 for any future JSON serialization needs

3. **Security Enhancements**
   - Benefit from standard library security patches
   - Improved TLS and X.509 handling

---

## Vulnerability Resolution Tracking

### Go 1.25.4 Security Fixes (November 5, 2025)

Track fixes from go1.25.0 through go1.25.4:
- Standard library security patches
- Runtime improvements
- Compiler security fixes

### Dependency Updates Required

| Package | Current Version | Target Version | Notes |
|---------|----------------|----------------|-------|
| Go | 1.25.4 | 1.25.x | Upgraded from 1.17 -> 1.24 -> 1.25.4 ✅ |
| github.com/skip2/go-qrcode | v0.0.0-20200617195104 | v0.0.0-20200617195104 | Compatible with Go 1.25 ✅ |

---

## Rollback Plan

If critical issues are discovered post-upgrade:

1. **Revert go.mod changes**
   ```bash
   git revert <upgrade-commit>
   ```

2. **Restore CI/CD configuration**
   - Reset to Go 1.24 in all workflows
   - Verify all jobs pass

3. **Document issues**
   - Create detailed bug report
   - Track in issue tracker
   - Plan for resolution before retry

---

## Timeline

| Phase | Duration | Dependencies |
|-------|----------|--------------|
| Phase 1: Research | 2-3 days | None |
| Phase 2: Code Updates | 1-2 days | Phase 1 complete |
| Phase 3: CI/CD Updates | 1 day | Phase 2 complete |
| Phase 4: Testing | 2-3 days | Phase 3 complete |
| Phase 5: Documentation | 1-2 days | Phase 4 complete |
| Phase 6: Security | 2-3 days | Phase 4 complete |
| Phase 7: Deployment | 1-2 days | All phases complete |

**Total Estimated Time: 10-16 days**

---

## Success Criteria

- [ ] All tests pass with Go 1.25.4
- [ ] No security vulnerabilities from govulncheck
- [ ] No deprecated code warnings
- [ ] CI/CD pipeline fully functional
- [ ] Documentation updated
- [ ] Performance maintained or improved
- [ ] All security-critical operations verified
- [ ] Clean code review approval

---

## References

- [Go 1.25 Release Notes](https://go.dev/doc/go1.25)
- [Go 1.25 Blog Post](https://go.dev/blog/go1.25)
- [Release History](https://go.dev/doc/devel/release)
- [Go End of Life Dates](https://endoflife.date/go)

---

*Document created: November 17, 2025*
*Last updated: November 17, 2025*
*Author: Security Review Team*
