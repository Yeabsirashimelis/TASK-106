# Audit Fix Repair Report

All 6 issues from `audit_report-1-fix_check.md` are now **Fixed**.

---

## 1. Search visibility split — Fixed

**Problem:** `SearchResources` checked membership but returned all visibility levels regardless of caller's course role.

**Fix:**
- `internal/service/resource_service.go:245-272` — `SearchResources` determines caller role in course: Admin → unfiltered `Search()`; Staff → unfiltered `Search()`; Enrolled → `SearchWithVisibility(ctx, courseID, query, VisibilityEnrolled, ...)`.
- `internal/repository/resource_repo.go:123-140` — `SearchWithVisibility()` method filters by `AND r.visibility = $5` in the FTS search query.
- `internal/handler/resource_handler.go:142-145` — Maps `ErrNotCourseMember` to HTTP 403 in search handler.

**Tests:** `fix2_resource_authz_test.go` and `fix3_all_audit_fixes_test.go` — visibility constants, membership role checks, admin bypass, non-member denial.

---

## 2. Resource creation requires course staff — Fixed

**Problem:** `CreateResource` accepted any authenticated caller without checking course staff membership.

**Fix:**
- `internal/service/resource_service.go:75-77` — Calls `requireCourseStaff(ctx, courseID, actorID, callerRole)` before any creation logic.
- `internal/service/resource_service.go:437-449` — `requireCourseStaff()` checks admin bypass, then verifies membership role = Staff. Returns `ErrResourceAccessDenied` if not staff.
- `internal/handler/resource_handler.go:38-40` — Maps `ErrResourceAccessDenied` to HTTP 403.

**Tests:** `fix2_resource_authz_test.go` and `fix3_all_audit_fixes_test.go` — error sentinel, membership role distinction, admin global access.

---

## 3. Idempotency 24h semantics robust in DB + service — Fixed

**Problem:** Migration 030 used `WHERE idempotency_expires_at > NOW()` (volatile) in a partial unique index. Migration 031 created a dedicated table but with a permanent unique on `(account_id, idempotency_key)` that blocked key reuse after 24h.

**Fix (Option A — dedicated table with deterministic index):**
- `migrations/031_create_idempotency_keys.up.sql` — Dedicated `idempotency_keys` table with `window_start`, `window_end` columns. Drops volatile partial index.
- `migrations/032_fix_idempotency_unique_index.up.sql:6-8` — Unique index changed to `(account_id, idempotency_key, window_start)` — deterministic, no volatile `NOW()` in predicate. Different `window_start` values allow reuse after 24h.
- `internal/repository/payment_repo.go:158-165` — `FindActiveIdempotencyKey()` queries `WHERE window_end > $3` (time-scoped).
- `internal/repository/payment_repo.go:186-194` — `DeleteExpiredIdempotencyKeyForAccount()` — targeted cleanup for specific (account_id, key) before new insert.
- `internal/service/payment_service.go:77-84` — Checks `FindActiveIdempotencyKey` first; if found returns original payment.
- `internal/service/payment_service.go:87-89` — Calls targeted `DeleteExpiredIdempotencyKeyForAccount()` then global `DeleteExpiredIdempotencyKeys()` before creating new key.

**Behavior:**
- Within 24h: `FindActiveIdempotencyKey` finds active key (`window_end > now`) → returns original payment (HTTP 200).
- After 24h: `FindActiveIdempotencyKey` returns nothing → expired key deleted → new payment + new idempotency key row (different `window_start`) succeeds (HTTP 201).
- DB uniqueness is deterministic (no volatile predicate).

**Tests:** `fix2_idempotency_table_test.go` and `fix3_all_audit_fixes_test.go` — duplicate within 24h, new create after 24h, window precision, boundary behavior.

---

## 4. Enforce extracted_text only for PDF/DOCX — Fixed

**Problem:** `UploadVersion` accepted `extracted_text` for any MIME type.

**Fix:**
- `internal/service/resource_service.go:32` — `ErrExtractedTextNotAllowed` sentinel defined.
- `internal/service/resource_service.go:293-295` — Rejects with `ErrExtractedTextNotAllowed` if `extractedText` is non-empty and MIME is not in `TextExtractableMimeTypes` (only `application/pdf` and DOCX).
- `internal/handler/resource_handler.go:194-195` — Maps `ErrExtractedTextNotAllowed` to HTTP 400.
- `internal/models/resource.go:119-122` — `TextExtractableMimeTypes` allowlist contains exactly 2 entries.

**Tests:** `fix2_resource_authz_test.go` and `fix3_all_audit_fixes_test.go` — PDF allowed, DOCX allowed, video rejected, image rejected, plain text rejected, legacy .doc rejected, error sentinel, subset-of-allowed check.

---

## 5. Account create error mapping — Fixed

**Problem:** Handler mapped many domain errors to HTTP 500.

**Fix:**
- `internal/service/account_service.go:19-20` — `ErrInvalidRole` and `ErrDuplicateUsername` sentinels.
- `internal/service/account_service.go:48` — Invalid role returns `fmt.Errorf("%w: %s", ErrInvalidRole, ...)`.
- `internal/service/account_service.go:67-69` — Duplicate username detected via error message, wrapped with `ErrDuplicateUsername`.
- `internal/handler/account_handler.go:34-43` — Error mapping: `ErrPasswordPolicy` → 400, `ErrInvalidRole` → 400, `ErrDuplicateUsername` → 409, only unknown/unexpected → 500.

**Tests:** `fix2_account_errors_test.go` and `fix3_all_audit_fixes_test.go` — conflict mapping (409), invalid role mapping (400), password policy mapping (400), all sentinels distinct, password validation rejects weak, accepts strong.

---

## 6. Access-tier audit logs for auth paths — Fixed

**Problem:** Auth login/refresh/logout not explicitly logged at tier=access.

**Fix:**
- `internal/service/auth_service.go:87-97` — Login failure (invalid password): `LogExtended` with `TierAccess`, source=`auth/login`, reason=`invalid password`.
- `internal/service/auth_service.go:141-153` — Login success: `LogExtended` with `TierAccess`, source=`auth/login`.
- `internal/service/auth_service.go:185-195` — Refresh token reuse: `LogExtended` with `TierAccess`, source=`auth/refresh`, reason=`token reuse detected`.
- `internal/service/auth_service.go:200-210` — Refresh token expired: `LogExtended` with `TierAccess`, source=`auth/refresh`, reason=`refresh token expired`.
- `internal/service/auth_service.go:254-262` — Refresh success: `LogExtended` with `TierAccess`, source=`auth/refresh`.
- `internal/service/auth_service.go:281-289` — Logout: `LogExtended` with `TierAccess`, source=`auth/logout`.

**No sensitive secrets in logs:** Password values and raw tokens are never logged. Only username, IP address, and account ID appear in details.

**Tests:** `fix2_auth_audit_test.go` and `fix3_all_audit_fixes_test.go` — login success/failure/locked entries, refresh success/reuse/expired entries, logout entry, access tier retention (30 days), no secrets in details, all auth audit actions distinct.

---

## Migration Changes

| File | Purpose |
|------|---------|
| `migrations/031_create_idempotency_keys.up.sql` | Dedicated idempotency table; drops volatile partial index |
| `migrations/031_create_idempotency_keys.down.sql` | Reverts to partial index |
| `migrations/032_fix_idempotency_unique_index.up.sql` | Changes unique index to `(account_id, idempotency_key, window_start)` — deterministic reuse |
| `migrations/032_fix_idempotency_unique_index.down.sql` | Reverts to permanent unique `(account_id, idempotency_key)` |

---

## Test Summary

**All tests passing.** (`go test ./unit_tests/... -count=1` → PASS)

| Test file | Tests | Covers |
|-----------|-------|--------|
| `fix2_resource_authz_test.go` | 12 | Fix 1 (visibility constants, membership roles), Fix 2 (staff requirement), Fix 4 (extracted_text) |
| `fix2_account_errors_test.go` | 10 | Fix 5 (error sentinels, wrapping, password validation, role validation) |
| `fix2_idempotency_table_test.go` | 9 | Fix 3 (model fields, window precision, active/expired, reuse, duplicate) |
| `fix2_auth_audit_test.go` | 10 | Fix 6 (access tier, all auth actions, entry fields, wiring, retention) |
| `fix3_all_audit_fixes_test.go` | 35 | All 6 fixes: comprehensive coverage across all issues |

---

## Cannot Confirm Statistically

None. All fixes are deterministic and testable:
- Idempotency index uses `(account_id, idempotency_key, window_start)` — no volatile predicate.
- Visibility filtering uses `WHERE visibility = $5` — deterministic SQL.
- Error mapping uses Go sentinel errors with `errors.Is()` — deterministic.
- Audit logging uses explicit `LogExtended()` calls with `TierAccess` — deterministic code paths.
