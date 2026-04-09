# Collegiate Athletics & Learning Ops API Specification

**Version:** 1.0.0
**Stack:** Go (Echo) + sqlx + PostgreSQL
**Deployment:** Single-node Docker, fully offline
**Base URL:** Public routes at `/`, protected routes at `/api`

---

## Table of Contents

1. [General Conventions](#1-general-conventions)
2. [Authentication](#2-authentication)
3. [Accounts](#3-accounts)
4. [Seasons, Teams & Venues](#4-seasons-teams--venues)
5. [Matches & Scheduling](#5-matches--scheduling)
6. [Match Assignments](#6-match-assignments)
7. [Courses](#7-courses)
8. [Course Outline](#8-course-outline)
9. [Course Memberships](#9-course-memberships)
10. [Resources](#10-resources)
11. [Resource Versions](#11-resource-versions)
12. [Moderation](#12-moderation)
13. [Reports](#13-reports)
14. [Review Workflow](#14-review-workflow)
15. [Payments](#15-payments)
16. [Reconciliation](#16-reconciliation)
17. [Audit Logs](#17-audit-logs)
18. [Health & Metrics](#18-health--metrics)

---

## 1. General Conventions

### 1.1 Authentication

Every request to `/api/*` must include a bearer token:

```
Authorization: Bearer <access_token>
```

Public endpoints: `POST /auth/login`, `POST /auth/refresh`, `GET /health`, `GET /health/detailed`, `GET /metrics`.

### 1.2 Error Response

```json
{ "message": "human-readable error description" }
```

Standard HTTP status codes: 400 (validation), 401 (unauthorized), 403 (forbidden), 404 (not found), 409 (conflict/invalid transition), 423 (locked), 429 (rate limited), 500 (internal error).

### 1.3 Pagination

List endpoints accept `?offset=0&limit=20`. Default limit: 20. Maximum: 100.

### 1.4 Rate Limiting

- General: configurable RPS per account (default 10/s, burst 20).
- Write operations: 60/minute per account.
- Reports: 10/day per reporter.

---

## 2. Authentication

| Method | Path | Auth | Roles | Description |
|---|---|---|---|---|
| POST | `/auth/login` | Public | — | Login with username/password. Returns access + refresh tokens. |
| POST | `/auth/refresh` | Public | — | Exchange refresh token for new token pair. |
| POST | `/api/auth/logout` | JWT | Any | Revoke refresh token. |

**POST /auth/login**
```json
// Request
{ "username": "admin", "password": "SecurePass123!", "device_fingerprint": { "user_agent": "...", "attributes": {"screen": "1920x1080"} } }
// Response 200
{ "access_token": "eyJ...", "refresh_token": "a1b2c3...", "expires_in": 1800, "account": { "id": "uuid", "username": "admin", "role": "Administrator", "status": "Active" } }
```

Login lockout: 423 after 5 failed attempts within 15 minutes.

---

## 3. Accounts

| Method | Path | Auth | Roles | Description |
|---|---|---|---|---|
| POST | `/api/accounts` | JWT | Administrator | Create account. |
| GET | `/api/accounts` | JWT | Administrator | List accounts. |
| GET | `/api/accounts/:id` | JWT | Administrator or self | Get account by ID. |
| PUT | `/api/accounts/:id/status` | JWT | Administrator | Update status (Active/Frozen/Deactivated). |
| PUT | `/api/accounts/:id/password` | JWT | Self only | Change password. |

**Error mapping:** Invalid role → 400. Duplicate username → 409. Password policy violation → 400.

---

## 4. Seasons, Teams & Venues

| Method | Path | Auth | Roles | Description |
|---|---|---|---|---|
| POST | `/api/seasons` | JWT | Scheduler, Admin | Create season. |
| GET | `/api/seasons` | JWT | Any | List seasons. |
| GET | `/api/seasons/:id` | JWT | Any | Get season. |
| POST | `/api/teams` | JWT | Scheduler, Admin | Create team in season. |
| GET | `/api/teams/season/:season_id` | JWT | Any | List teams by season. |
| POST | `/api/venues` | JWT | Scheduler, Admin | Create venue. |
| GET | `/api/venues` | JWT | Any | List venues. |

---

## 5. Matches & Scheduling

| Method | Path | Auth | Roles | Description |
|---|---|---|---|---|
| POST | `/api/matches` | JWT | Scheduler, Admin | Create match (with validation). |
| POST | `/api/matches/import` | JWT | Scheduler, Admin | Bulk import matches. |
| POST | `/api/matches/generate` | JWT | Scheduler, Admin | Generate round-robin schedule. |
| GET | `/api/matches?season_id=&round=` | JWT | Any | List matches. |
| GET | `/api/matches/:id` | JWT | Any | Get match. |
| PUT | `/api/matches/:id` | JWT | Scheduler, Admin | Update match (Draft only). |
| PUT | `/api/matches/:id/status` | JWT | Scheduler, Admin | Transition status. |

**POST /api/matches/generate**
```json
// Request
{ "season_id": "uuid", "venue_ids": ["uuid1","uuid2"], "start_date": "2025-09-01", "interval_days": 7, "start_time": "14:00" }
// Response 201
{ "created": 6, "rounds": 3, "errors": [], "matches": [...] }
```

Scheduling violations return 409 with details. Override requires `override_reason` field.

---

## 6. Match Assignments

| Method | Path | Auth | Roles | Description |
|---|---|---|---|---|
| POST | `/api/assignments` | JWT | Scheduler, Admin | Assign referee/staff. |
| GET | `/api/assignments/match/:match_id` | JWT | Any | List assignments. |
| PUT | `/api/assignments/:id/reassign` | JWT | Scheduler, Admin | Reassign (reason required). |
| DELETE | `/api/assignments/:id` | JWT | Scheduler, Admin | Remove assignment. |

Locked once match is In-Progress or beyond (409).

---

## 7. Courses

| Method | Path | Auth | Roles | Description |
|---|---|---|---|---|
| POST | `/api/courses` | JWT | Instructor, Admin | Create course (creator auto-added as staff). |
| GET | `/api/courses` | JWT | Any | List (admin/instructor see all; others see Published). |
| GET | `/api/courses/:id` | JWT | Any* | Get course (*non-published requires membership). |
| PUT | `/api/courses/:id` | JWT | Instructor, Admin | Update (course staff only). |

---

## 8. Course Outline

| Method | Path | Auth | Roles | Description |
|---|---|---|---|---|
| POST | `/api/outline-nodes` | JWT | Instructor, Admin | Create chapter/unit (course staff). |
| GET | `/api/outline-nodes/course/:course_id` | JWT | Any* | Get tree (*requires course membership). |
| PUT | `/api/outline-nodes/:id` | JWT | Instructor, Admin | Update node (course staff). |
| DELETE | `/api/outline-nodes/:id` | JWT | Instructor, Admin | Delete node (course staff). |

---

## 9. Course Memberships

| Method | Path | Auth | Roles | Description |
|---|---|---|---|---|
| POST | `/api/courses/:course_id/members` | JWT | Instructor, Admin | Add member (course staff). |
| GET | `/api/courses/:course_id/members` | JWT | Any* | List members (*requires course staff/admin). |
| DELETE | `/api/courses/:course_id/members/:id` | JWT | Instructor, Admin | Remove member (course staff/admin). |

---

## 10. Resources

| Method | Path | Auth | Roles | Description |
|---|---|---|---|---|
| POST | `/api/resources` | JWT | Instructor, Admin | Create resource (course staff required). |
| GET | `/api/resources?course_id=` | JWT | Any* | List (*membership required; enrolled see only Enrolled-visibility). |
| GET | `/api/resources/search?course_id=&q=` | JWT | Any* | Full-text search (*visibility-aware). |
| GET | `/api/resources/:id` | JWT | Any* | Get resource (*visibility-enforced). |
| PUT | `/api/resources/:id` | JWT | Instructor, Admin | Update (course staff required). |

Tags: max 20 per resource, max 32 characters each. Managed via `tags` array in create/update.

---

## 11. Resource Versions

| Method | Path | Auth | Roles | Description |
|---|---|---|---|---|
| POST | `/api/resources/:id/versions` | JWT | Instructor, Admin | Upload version (multipart, course staff required). |
| GET | `/api/resources/:id/versions` | JWT | Any* | List versions (*visibility-enforced). |
| GET | `/api/resources/versions/:version_id/download` | JWT | Any* | Download with SHA-256 integrity check. |
| GET | `/api/resources/versions/:version_id/preview` | JWT | Any* | Inline preview. |

`extracted_text` form field accepted only for PDF and DOCX MIME types (400 otherwise). MIME validated against allowlist.

---

## 12. Moderation

### Sensitive Word Dictionaries (Administrator only)

| Method | Path | Description |
|---|---|---|
| POST | `/api/moderation/dictionaries` | Create dictionary. |
| GET | `/api/moderation/dictionaries` | List dictionaries. |
| GET | `/api/moderation/dictionaries/:id` | Get dictionary. |
| PUT | `/api/moderation/dictionaries/:id` | Update dictionary. |
| DELETE | `/api/moderation/dictionaries/:id` | Delete dictionary + words. |
| POST | `/api/moderation/dictionaries/:dict_id/words` | Add word. |
| POST | `/api/moderation/dictionaries/:dict_id/words/bulk` | Bulk add words. |
| GET | `/api/moderation/dictionaries/:dict_id/words` | List words. |
| DELETE | `/api/moderation/words/:id` | Delete word. |

### Content Check (any authenticated)

| Method | Path | Description |
|---|---|---|
| POST | `/api/moderation/check` | Check text against active dictionaries. Returns `{ clean, matches[] }`. |

### Moderation Reviews (Reviewer, Admin)

| Method | Path | Description |
|---|---|---|
| POST | `/api/moderation/reviews` | Create review (Pending). |
| GET | `/api/moderation/reviews?status=` | List reviews. |
| GET | `/api/moderation/reviews/:id` | Get review. |
| PUT | `/api/moderation/reviews/:id/decide` | Approve/Reject (reason required). |

---

## 13. Reports

| Method | Path | Auth | Roles | Description |
|---|---|---|---|---|
| POST | `/api/reports` | JWT | Any | Create report (10/day limit). |
| GET | `/api/reports?status=` | JWT | Reviewer, Admin | List reports. |
| GET | `/api/reports/:id` | JWT | Reviewer, Admin | Get report. |
| PUT | `/api/reports/:id/status` | JWT | Reviewer, Admin | Update status. |
| PUT | `/api/reports/:id/assign` | JWT | Reviewer, Admin | Assign to reviewer. |
| POST | `/api/reports/:id/evidence` | JWT | Reviewer, Admin | Upload evidence file. |
| GET | `/api/reports/:id/evidence` | JWT | Reviewer, Admin | List evidence. |
| GET | `/api/reports/evidence/:evidence_id/download` | JWT | Reviewer, Admin | Download evidence. |
| POST | `/api/reports/:id/notes` | JWT | Reviewer, Admin | Add progress note. |
| GET | `/api/reports/:id/notes` | JWT | Reviewer, Admin | List notes. |

---

## 14. Review Workflow

### Review Configs (Administrator only)

| Method | Path | Description |
|---|---|---|
| POST | `/api/reviews/configs` | Create config (1-3 levels). |
| GET | `/api/reviews/configs` | List configs. |
| GET | `/api/reviews/configs/:id` | Get config. |
| PUT | `/api/reviews/configs/:id` | Update config. |
| DELETE | `/api/reviews/configs/:id` | Delete config. |

### Review Requests

| Method | Path | Auth | Roles |
|---|---|---|---|
| POST | `/api/reviews/requests` | JWT | Any |
| GET | `/api/reviews/requests?status=` | JWT | Reviewer, Admin |
| GET | `/api/reviews/requests/by-entity?entity_type=&entity_id=` | JWT | Reviewer, Admin |
| GET | `/api/reviews/requests/:id` | JWT | Reviewer, Admin |
| POST | `/api/reviews/requests/:id/resubmit` | JWT | Reviewer, Admin |
| GET | `/api/reviews/requests/:id/follow-up-requests` | JWT | Reviewer, Admin |
| POST | `/api/reviews/requests/:id/follow-ups` | JWT | Reviewer, Admin |
| GET | `/api/reviews/requests/:id/follow-ups` | JWT | Reviewer, Admin |
| GET | `/api/reviews/my-assignments` | JWT | Any |

### Review Levels (Reviewer, Admin)

| Method | Path | Description |
|---|---|---|
| GET | `/api/reviews/levels/request/:request_id` | List levels for request. |
| PUT | `/api/reviews/levels/:id/assign` | Assign reviewer to level. |
| PUT | `/api/reviews/levels/:id/decide` | Decide: Approved/Rejected/Returned with annotation. |

---

## 15. Payments

| Method | Path | Auth | Roles | Description |
|---|---|---|---|---|
| POST | `/api/payments` | JWT | Finance, Admin | Create obligation (idempotent: 24h window). |
| GET | `/api/payments?status=` | JWT | Finance, Admin | List payments. |
| GET | `/api/payments/:id` | JWT | Finance, Admin | Get payment. |
| GET | `/api/payments/account/:account_id` | JWT | Finance, Admin | List by account. |
| GET | `/api/payments/failed-retriable` | JWT | Finance, Admin | List retriable failures. |
| PUT | `/api/payments/:id/sign` | JWT | Finance, Admin | Sign posting → Settled. |
| PUT | `/api/payments/:id/fail` | JWT | Finance, Admin | Mark failed. |
| PUT | `/api/payments/:id/retry` | JWT | Finance, Admin | Retry (max 3) → back to Obligation. |

**Idempotency:** Duplicate key within 24h returns HTTP 200 with original. After 24h, same key creates new entry (HTTP 201).

---

## 16. Reconciliation

| Method | Path | Auth | Roles | Description |
|---|---|---|---|---|
| GET | `/api/reconciliation/summary?date=` | JWT | Finance, Admin | Daily summary. |
| GET | `/api/reconciliation/summary/range?start_date=&end_date=` | JWT | Finance, Admin | Date range summary. |
| POST | `/api/reconciliation/reports` | JWT | Finance, Admin | Generate report + CSV. |
| GET | `/api/reconciliation/reports` | JWT | Finance, Admin | List reports. |
| GET | `/api/reconciliation/reports/:id` | JWT | Finance, Admin | Get report. |
| GET | `/api/reconciliation/reports/:id/csv` | JWT | Finance, Admin | Download CSV. |

---

## 17. Audit Logs

### Read Operations (Auditor, Administrator)

| Method | Path | Description |
|---|---|---|
| GET | `/api/audit/logs?actor_id=&entity_type=&action=&tier=&start_time=&end_time=` | Query with filters. |
| GET | `/api/audit/logs/export` | CSV export (same filters). |
| GET | `/api/audit/logs/by-entity?entity_type=&entity_id=` | By entity. |
| GET | `/api/audit/logs/by-actor/:actor_id` | By actor. |
| GET | `/api/audit/logs/:id` | Get single entry. |
| GET | `/api/audit/logs/tier-counts` | Count by tier. |
| GET | `/api/audit/hash-chain/verify?date=` | Verify daily hash chain. |
| GET | `/api/audit/hash-chain` | List hash chain entries. |

### Write Operations (Administrator only)

| Method | Path | Description |
|---|---|---|
| POST | `/api/audit/hash-chain/build` | Build daily hash chain. |
| POST | `/api/audit/purge-expired` | Purge expired entries. |

---

## 18. Health & Metrics

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/health` | Public | Simple OK check. |
| GET | `/health/detailed` | Public | DB, disk, pool stats, goroutines, uptime. |
| GET | `/metrics` | Public | Request count, errors, p50/p95/p99 latency, sessions, memory. |
