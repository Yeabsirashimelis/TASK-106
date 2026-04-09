# Audit Report #1 Fix Check (Rewritten)

## Verdict
- Conclusion: **Pass** (for the exact issue set from Self-Test Report #1)
- Scope of this check: only the previously reported 9 issues were re-verified statically.

## Fix Status Summary
- Fixed: **9 / 9**
- Partially fixed: **0 / 9**
- Not fixed: **0 / 9**

## Issue-by-Issue Verification

1) **Blocker - Object-level authorization gaps in course/resource modules**
- Status: **Fixed**
- Evidence:
  - Course member list/remove now enforce staff/admin at service layer: `repo/internal/service/course_service.go:349`, `repo/internal/service/course_service.go:356`.
  - Resource update/upload/list versions enforce course staff/access checks: `repo/internal/service/resource_service.go:190`, `repo/internal/service/resource_service.go:283`, `repo/internal/service/resource_service.go:375`.
  - Resource search now has visibility-aware branch for enrolled users: `repo/internal/service/resource_service.go:270`, `repo/internal/repository/resource_repo.go:123`.

2) **Blocker - Auditor role not read-only**
- Status: **Fixed**
- Evidence:
  - Auditor group now has read/verify/list only: `repo/internal/router/router.go:238`, `repo/internal/router/router.go:247`.
  - Audit write endpoints are Administrator-only: `repo/internal/router/router.go:251`, `repo/internal/router/router.go:252`, `repo/internal/router/router.go:253`.

3) **High - Schedule generation capability missing**
- Status: **Fixed**
- Evidence:
  - Route exists: `POST /api/matches/generate` at `repo/internal/router/router.go:89`.
  - Handler exists: `repo/internal/handler/match_handler.go:160`.
  - Service generation flow exists: `repo/internal/service/match_service.go:332`.

4) **High - Payment idempotency semantics conflict with schema**
- Status: **Fixed**
- Evidence:
  - Active-key lookup is window-scoped: `repo/internal/repository/payment_repo.go:158`, `repo/internal/repository/payment_repo.go:161`.
  - Service now uses active-key table and returns original payment if key is still active: `repo/internal/service/payment_service.go:75`, `repo/internal/service/payment_service.go:80`.
  - Dedicated idempotency table migration added: `repo/migrations/031_create_idempotency_keys.up.sql:6`.

5) **High - Compliance traceability fields underutilized (source/reason/before/after)**
- Status: **Fixed**
- Evidence:
  - Extended audit structure is implemented and persisted: `repo/internal/service/audit_service.go:46`, `repo/internal/repository/audit_log_repo.go:22`.
  - Extended audit usage now appears in key flows (auth/course/payment): `repo/internal/service/auth_service.go:87`, `repo/internal/service/course_service.go:150`, `repo/internal/service/payment_service.go:167`.

6) **High - Review disposition write-back not wired in composition root**
- Status: **Fixed**
- Evidence:
  - Disposition callbacks are registered in main composition root: `repo/cmd/server/main.go:99`, `repo/cmd/server/main.go:105`, `repo/cmd/server/main.go:111`.
  - Final approval/rejection still triggers disposition execution: `repo/internal/service/review_service.go:370`, `repo/internal/service/review_service.go:387`.

7) **High - Resource full-text extraction scope (PDF/DOCX-only) not enforced**
- Status: **Fixed**
- Evidence:
  - Enforced in upload path: rejects extracted text for non-PDF/DOCX MIME: `repo/internal/service/resource_service.go:293`, `repo/internal/service/resource_service.go:294`.
  - Constraint source allowlist: `repo/internal/models/resource.go:118`.

8) **Medium - Account creation maps validation/conflict to 500**
- Status: **Fixed**
- Evidence:
  - Invalid role now maps to 400: `repo/internal/handler/account_handler.go:37`.
  - Duplicate username now maps to 409: `repo/internal/handler/account_handler.go:40`.
  - Password policy maps to 400: `repo/internal/handler/account_handler.go:34`.

9) **Medium - Access-tier logging not implemented in auth/request paths**
- Status: **Fixed**
- Evidence:
  - Login failure/success now logged with `TierAccess`: `repo/internal/service/auth_service.go:91`, `repo/internal/service/auth_service.go:146`.
  - Refresh/logout now logged with `TierAccess`: `repo/internal/service/auth_service.go:225`, `repo/internal/service/auth_service.go:252`.

## Additional Static Evidence (Tests Updated)
- API tests now include object-level access checks and auditor read-only checks: `repo/API_tests/run_api_tests.sh:441`, `repo/API_tests/run_api_tests.sh:698`.
- API tests now include schedule generation endpoint checks: `repo/API_tests/run_api_tests.sh:489`.

## Boundary Note
- This report is static-only and does not claim runtime performance or live migration execution outcomes.
