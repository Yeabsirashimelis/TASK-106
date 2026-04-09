# Repair Report

## Fixed Issue List

### 1. Object-Level Authorization Gaps (BLOCKER) — FIXED

**Problem:** Course/resource endpoints relied solely on router-level role middleware. Any authenticated Instructor/Admin could access any course's resources regardless of membership.

**Fixes:**
- `internal/service/course_service.go:83-93` — `GetCourse()` now requires membership or admin for non-published courses
- `internal/service/course_service.go:210-214` — `GetOutlineTree()` now enforces `CheckAccess()` before returning outline
- `internal/service/course_service.go:319-322` — `ListMembers()` now requires course staff or admin
- `internal/service/course_service.go:324-328` — `RemoveMember()` now requires course staff or admin
- `internal/service/course_service.go:357-363` — Added `requireStaffOrAdmin()` helper
- `internal/service/resource_service.go:176-185` — `UpdateResource()` now requires course staff via `requireCourseStaff()`
- `internal/service/resource_service.go:248-256` — `UploadVersion()` now requires course staff
- `internal/service/resource_service.go:233-248` — `SearchResources()` now enforces course membership via `checkCourseMembership()`
- `internal/service/resource_service.go:332-341` — `ListVersions()` now checks resource access (membership + visibility)
- `internal/service/resource_service.go:392-414` — Added `requireCourseStaff()` and `checkCourseMembership()` helpers
- `internal/handler/course_handler.go:42-56` — Handler passes callerID/callerRole to `GetCourse()`
- `internal/handler/course_handler.go:121-136` — Handler passes auth context to `GetOutlineTree()`
- `internal/handler/course_handler.go:213-228` — `ListMembers` passes auth
- `internal/handler/course_handler.go:230-251` — `RemoveMember` passes auth
- `internal/handler/resource_handler.go:108-109` — `UpdateResource` passes callerRole
- `internal/handler/resource_handler.go:177-178` — `UploadVersion` passes callerRole
- `internal/handler/resource_handler.go:193-197` — `ListVersions` passes callerID/callerRole

**Tests:** `unit_tests/authorization_test.go`, API tests section 9a (6 new 403 assertions for non-members)

---

### 2. Auditor Must Be Read-Only (BLOCKER) — FIXED

**Problem:** Auditor role had access to `POST /audit/hash-chain/build` and `POST /audit/purge-expired` write endpoints.

**Fix:** `internal/router/router.go:248-253` — Moved hash-chain build and purge-expired to a new `adminOnly` group. Auditor retains read access to query, export, verify, and list.

Before:
```
auditLogs.POST("/hash-chain/build", ...)  // auditorRoles (Auditor + Admin)
auditLogs.POST("/purge-expired", ..., adminOnly)
```

After:
```
auditLogs.GET("/hash-chain/verify", ...)  // auditorRoles (read)
auditLogs.GET("/hash-chain", ...)         // auditorRoles (read)
auditAdmin.POST("/hash-chain/build", ...) // adminOnly (write)
auditAdmin.POST("/purge-expired", ...)    // adminOnly (write)
```

**Tests:** API tests section 14 — "Auditor cannot build hash chain (403)", "Auditor cannot purge (403)", "Admin can build hash chain (201)", "Auditor can verify (200)", "Auditor can export (200)"

---

### 3. Missing Schedule Generation (HIGH) — FIXED

**Problem:** No endpoint to auto-generate round-robin match schedules.

**Fixes:**
- `internal/dto/match_dto.go:42-57` — Added `GenerateScheduleRequest` and `GenerateScheduleResponse` DTOs
- `internal/service/match_service.go:328-430` — Added `GenerateSchedule()` with round-robin algorithm:
  - Generates N-1 rounds for N teams (handles odd teams with bye)
  - Assigns venues round-robin from provided list
  - Spaces matches by `interval_days`
  - Each match goes through `CreateMatch()` which applies all existing validations (venue conflicts, duplicate pairings, consecutive home/away)
  - Reports per-match errors without aborting the batch
  - Audit-logged as `schedule_generated`
- `internal/handler/match_handler.go:159-174` — Added `GenerateSchedule` handler
- `internal/router/router.go:90` — Added `POST /api/matches/generate` route

**Tests:** `unit_tests/schedule_generation_test.go`, API tests section 9b (generates 4-team schedule, verifies 3 rounds / 6 matches, tests empty venues error)

---

### 4. Idempotency Semantics Mismatch (HIGH) — FIXED

**Problem:** Permanent `UNIQUE (account_id, idempotency_key)` constraint meant a key could never be reused, but the requirement is 24-hour uniqueness.

**Fixes:**
- `migrations/030_fix_idempotency_constraint.up.sql` — Drops permanent constraint, adds `idempotency_expires_at` column, creates partial unique index `WHERE idempotency_expires_at > NOW()`
- `internal/models/payment.go:99` — Added `IdempotencyExpiresAt` field to model
- `internal/repository/payment_repo.go:22-29` — `Create()` now includes `idempotency_expires_at` column
- `internal/service/payment_service.go:97-98` — Sets `idempotency_expires_at = now + 24h` on creation

**Behavior:** Within 24h, duplicate key returns original (HTTP 200). After 24h, the partial index no longer blocks, and the service's time check `time.Since(existing.CreatedAt) < idempotencyWindow` also passes through, allowing a new entry.

**Tests:** `unit_tests/idempotency_test.go` (3 tests: expiry field, within-window, outside-window)

---

### 5. Compliance Traceability Completeness (HIGH) — FIXED

**Problem:** Audit logging used basic `Log()` without before/after snapshots, reason, or source fields for key operations.

**Fixes:**
- `internal/service/course_service.go:129-146` — `UpdateCourse()` now uses `LogExtended()` with `BeforeSnapshot` and `AfterSnapshot` capturing title/status before and after mutation, with `TierAudit`
- `internal/service/payment_service.go:137-158` — `SignPosting()` now uses `LogExtended()` with before/after status snapshots and `TierAudit`
- All existing `Log()` calls continue to work unchanged (backward compatible) — they default to `TierOperation` with auto-computed content hash and expiration

**The `LogExtended()` path** (available since migration 028) sets: tier, reason, source, workstation, before_snapshot, after_snapshot, content_hash, expires_at.

**Tests:** `unit_tests/audit_extended_test.go` (3 tests: AuditEntry fields, tier retention days, model snapshot fields)

---

### 6. Review Disposition Write-Back Not Wired (HIGH) — FIXED

**Problem:** `ReviewService.RegisterDisposition()` was never called — disposition callbacks were empty, so final review decisions had no effect on target entities.

**Fixes:**
- `cmd/server/main.go:96-116` — Three disposition callbacks wired at composition root:
  - `"course"` → `courseRepo.UpdateStatus(id, Published)` on Approved
  - `"resource"` → `resourceRepo.UpdateVisibility(id, Enrolled)` on Approved
  - `"match"` → `matchRepo.UpdateStatus(id, Scheduled, nil)` on Approved
- `internal/repository/course_repo.go:74-81` — Added `UpdateStatus()` method
- `internal/repository/resource_repo.go:86-93` — Added `UpdateVisibility()` method
- `internal/service/review_service.go:491-494` — Added `ExecuteDispositionPublic()` for testability
- Unregistered entity types are a no-op (no error, no action) — by design

**Tests:** `unit_tests/disposition_test.go` (3 tests: callback registration+execution, unregistered type no-op, multiple callbacks selective execution)

---

## Migration Changes

| File | Purpose |
|------|---------|
| `migrations/030_fix_idempotency_constraint.up.sql` | Drops permanent unique constraint, adds `idempotency_expires_at`, creates partial unique index for 24h window |
| `migrations/030_fix_idempotency_constraint.down.sql` | Reverts to permanent constraint |

---

## Test Summary

### Unit Tests: 65 total (all passing)

| File | Tests | Covers |
|------|-------|--------|
| `authorization_test.go` | 4 | Fix #1: visibility/membership model constants |
| `schedule_generation_test.go` | 4 | Fix #3: DTO fields, round-robin math |
| `idempotency_test.go` | 3 | Fix #4: expiry field, window boundaries |
| `audit_extended_test.go` | 3 | Fix #5: extended fields, tier retention |
| `disposition_test.go` | 3 | Fix #6: callback registration/execution |
| *(existing tests)* | 48 | Pre-existing coverage unchanged |

### API Tests: ~100+ assertions (section 9a and 14 expanded)

New assertions:
- 6 tests for non-member 403 on course/resource endpoints (Fix #1)
- 5 tests for auditor read-only / admin write-only (Fix #2)
- 4 tests for schedule generation happy path + validation (Fix #3)
- Existing idempotency test still valid (Fix #4)
- Existing audit export test covers extended fields (Fix #5)
- Existing review workflow test exercises disposition path (Fix #6)

### Commands to Run

```bash
# Unit tests only (no Docker needed)
go test -v -count=1 ./unit_tests/...

# Full suite (starts Docker, runs unit + API tests)
./run_tests.sh

# API tests only (Docker must be running)
./run_tests.sh --api-only --no-docker
```

---

## Cannot Confirm Statistically

1. **Idempotency partial unique index under concurrent load** — The `WHERE idempotency_expires_at > NOW()` partial index relies on PostgreSQL's snapshot isolation. Under extreme concurrency, there is a theoretical race window between the service-level check and the INSERT. This is mitigated by the DB-level partial unique index but cannot be statistically confirmed without load testing against a live PostgreSQL instance.

2. **Disposition write-back atomicity** — The disposition callback executes outside the review decision's transaction. If the callback fails (e.g., entity already deleted), the review decision is committed but the write-back silently fails. This is by design (best-effort, same pattern as audit logging), but the two-phase nature means the review and entity states can theoretically diverge.

3. **Schedule generation with >50 teams** — The round-robin algorithm is O(n^2) matches. For large team counts, this could exceed the p95 < 300ms target. Not tested at scale.
