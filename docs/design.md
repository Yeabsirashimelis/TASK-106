# Collegiate Athletics & Learning Ops API Platform — Design Document

Version: 1.0
Date: 2026-04-09
Status: Final

---

## Table of Contents

1. [Introduction & Scope](#1-introduction--scope)
2. [System Architecture](#2-system-architecture)
3. [Technology Stack](#3-technology-stack)
4. [Database Design](#4-database-design)
5. [Authentication & Security](#5-authentication--security)
6. [Authorization & RBAC](#6-authorization--rbac)
7. [Scheduling & Match Lifecycle](#7-scheduling--match-lifecycle)
8. [Courses & Teaching Resources](#8-courses--teaching-resources)
9. [Content Moderation & Risk Control](#9-content-moderation--risk-control)
10. [Review Workflow](#10-review-workflow)
11. [Payments & Reconciliation](#11-payments--reconciliation)
12. [Audit Logs & Compliance](#12-audit-logs--compliance)
13. [Observability](#13-observability)
14. [Deployment & Performance](#14-deployment--performance)

---

## 1. Introduction & Scope

### 1.1 Purpose

This platform is a backend-only API that unifies league scheduling, instructional content governance, and compliance traceability for a collegiate athletics department running seasons and training programs on a single offline server. It handles match scheduling, course management, content moderation, multi-level review workflows, offline payment ledger, and tamper-evident audit logging.

### 1.2 Boundaries

- **Backend API only.** No frontend, no server-rendered HTML. All interaction occurs through JSON HTTP endpoints.
- **Fully offline.** Single Docker node on a local network. No outbound internet, no SaaS integrations, no external authentication providers.
- **Single-node deployment.** PostgreSQL and the Go application server run in Docker containers on the same host.

### 1.3 Target Users

| Role | Description |
|---|---|
| Administrator | Full system configuration, user provisioning, security policy, audit review, review config |
| Scheduler | Season/match scheduling, referee/staff assignments, venue management |
| Instructor | Course creation/management, teaching resource uploads, course membership |
| Reviewer | Multi-level approval workflows, moderation decisions, report handling |
| Finance Clerk | Payment ledger management, posting verification, reconciliation |
| Auditor | Read-only audit log access, hash chain verification, compliance export |

---

## 2. System Architecture

### 2.1 Component Overview

```
+-------------------------------------------------------------+
|  Docker Compose (single node)                               |
|                                                             |
|  +------------------+      +-----------------------------+  |
|  | Go (Echo)        |      | PostgreSQL 16               |  |
|  | HTTP Server      |<---->| (sqlx, raw SQL)             |  |
|  | (port 8080)      |      |                             |  |
|  +------------------+      +-----------------------------+  |
|         |                                                   |
|         v                                                   |
|  +------------------+                                       |
|  | Local File       |                                       |
|  | System           |                                       |
|  | /app/storage/    |                                       |
|  +------------------+                                       |
+-------------------------------------------------------------+
```

### 2.2 Request Lifecycle

1. Echo accepts the TCP connection and parses the HTTP request.
2. Metrics middleware records latency and active session count.
3. Rate-limiting middleware checks per-account/IP counters.
4. Write-limiting middleware enforces 60 writes/min per account (anti-spam).
5. JWT authentication middleware extracts and validates the bearer token.
6. Role-guard middleware checks the caller's role against the endpoint's allowed roles.
7. Handler deserializes the request, delegates to the service layer.
8. Service layer enforces object-level authorization (course membership, visibility).
9. Repository layer executes parameterized SQL via sqlx.
10. Audit service logs the action (best-effort, non-blocking).
11. Response is serialized to JSON.

### 2.3 Layered Architecture

```
Handler  →  Service  →  Repository  →  PostgreSQL
   ↓            ↓
  DTO        Models
```

- **Handler:** HTTP request/response mapping, input validation, error-to-status-code translation.
- **Service:** Business logic, authorization enforcement, audit logging, state machine transitions.
- **Repository:** Raw SQL queries via sqlx, no ORM. Each entity has its own repository.
- **Models:** Database-mapped structs with `db:` tags for sqlx and `json:` tags for serialization.
- **DTOs:** Request/response structs. Response DTOs mask sensitive fields (password hashes never serialized).

---

## 3. Technology Stack

| Component | Technology | Rationale |
|---|---|---|
| Language | Go 1.22 | Performance, strong concurrency, single binary deployment |
| HTTP Framework | Echo v4 | Lightweight, middleware-friendly, good context propagation |
| Database Access | sqlx | Type-safe raw SQL, no ORM magic, auditable queries |
| Database | PostgreSQL 16 | ACID, JSONB for audit details, GIN for full-text search, ENUM types |
| Authentication | golang-jwt/jwt/v5 (HS256) | Offline-compatible, no external auth provider needed |
| Password Hashing | bcrypt (cost 12) | Industry standard, timing-safe |
| File Hashing | SHA-256 | Deduplication and integrity verification |
| Rate Limiting | golang.org/x/time/rate | In-memory token bucket, no external dependency |
| Migrations | golang-migrate/v4 | Auto-run on startup, file-based, up/down support |
| Decimal Math | shopspring/decimal | Precise financial arithmetic (NUMERIC 12,2) |
| Deployment | Docker + docker-compose | Single `docker compose up`, no external dependencies |

---

## 4. Database Design

### 4.1 Schema Overview

31 migrations producing the following table groups:

**Authentication & Accounts**
- `accounts` — UUID PK, unique username, bcrypt password_hash, role ENUM, status ENUM
- `refresh_tokens` — SHA-256 hashed tokens, revocation flag, device linkage
- `devices` — salted SHA-256 fingerprint hashes, unique per (account, fingerprint)
- `login_attempts` — success/fail tracking with IP, indexed for lockout queries

**Scheduling**
- `seasons` — name, date range, status (Planning/Active/Completed/Archived)
- `teams` — name unique per season
- `venues` — name, location, capacity
- `matches` — round, home/away teams, venue, scheduled_at, status (Draft→Scheduled→In-Progress→Final→Canceled), override_reason
- `match_assignments` — referee/staff linked to matches with role, reassignment tracking

**Courses & Resources**
- `courses` — title, description, status (Draft/Published/Archived)
- `course_outline_nodes` — tree via parent_id + order_index, types: Chapter/Unit
- `course_memberships` — Staff/Enrolled role per account per course
- `resources` — Document/Video/Link, visibility (Staff/Enrolled), tsvector for FTS
- `resource_versions` — immutable versions with SHA-256, size, MIME, storage_path
- `resource_tags` — max 20 per resource, 32 chars each

**Moderation**
- `sensitive_word_dictionaries` — named dictionaries with active toggle
- `sensitive_words` — words with severity, case-insensitive index
- `moderation_reviews` — Pending/Approved/Rejected workflow
- `reports` — category ENUM, status lifecycle, evidence attachments, progress notes

**Review Workflow**
- `review_configs` — 1-3 approval levels per review type
- `review_requests` — multi-level review linked to entity, parent_id for follow-ups
- `review_levels` — per-level decision tracking with assignee
- `review_follow_ups` — supplementary material per level

**Payments**
- `payments_ledger` — amount_usd (NUMERIC 12,2), channel ENUM, status (Obligation/Settled/Failed), retry_count
- `idempotency_keys` — dedicated table for 24h uniqueness with deterministic window
- `reconciliation_reports` — daily snapshots with CSV export path

**Audit & Compliance**
- `audit_logs` — extended with tier (access/operation/audit), before/after snapshots, content_hash, expires_at
- `audit_hash_chain` — daily tamper-evident chain linking batch hashes

### 4.2 Key Constraints

- All PKs are UUID (gen_random_uuid)
- Foreign keys with appropriate CASCADE/RESTRICT/SET NULL
- CHECK constraints: season dates, team != opponent, approval levels 1-3
- Unique constraints: username, (account, fingerprint), (team, season), (account, idempotency_key)
- GIN indexes for full-text search on resources and resource versions

---

## 5. Authentication & Security

### 5.1 JWT Tokens

- **Access tokens:** 30-minute HS256 JWT with claims: sub (account_id), username, role, exp, iat.
- **Refresh tokens:** Cryptographically random 32-byte value. Only SHA-256 hash stored in DB. Rotating: each use revokes current and issues new pair.
- **Reuse detection:** If a revoked refresh token is presented, all tokens for that account are revoked (potential theft indicator).

### 5.2 Password Policy

- Minimum 12 characters, at least 1 uppercase, 1 lowercase, 1 digit.
- bcrypt cost factor 12.
- Login lockout: 15 minutes after 5 failed attempts (rolling window).
- Anti-enumeration: constant-time dummy bcrypt on invalid usernames.

### 5.3 Device Fingerprinting

- Optional. Client sends user_agent + stable device attributes.
- Server computes `SHA-256(salt + user_agent + sorted_attributes)`.
- Stored unique per (account, fingerprint_hash).

---

## 6. Authorization & RBAC

### 6.1 Two-Layer Authorization

1. **Route-level (middleware):** `RequireRoles()` middleware restricts endpoint access by role. Applied at the Echo group level.
2. **Object-level (service):** Service methods enforce fine-grained access — course membership, resource visibility, self-access checks.

### 6.2 Role Permissions Matrix

| Capability | Admin | Scheduler | Instructor | Reviewer | Finance | Auditor |
|---|:---:|:---:|:---:|:---:|:---:|:---:|
| Account CRUD | W | - | - | - | - | - |
| Season/Match scheduling | W | W | - | - | - | - |
| Course/Resource management | W | - | W* | - | - | - |
| Moderation dictionaries | W | - | - | - | - | - |
| Moderation reviews | W | - | - | W | - | - |
| Reports handling | W | - | - | W | - | - |
| Review workflow decisions | W | - | - | W | - | - |
| Payment ledger | W | - | - | - | W | - |
| Audit logs (read) | W | - | - | - | - | R |
| Audit write ops | W | - | - | - | - | - |

*Instructor: requires course staff membership for write operations.

### 6.3 Object-Level Authorization (Courses)

- **GetCourse:** Published courses visible to all authenticated users. Draft/Archived require course membership or admin.
- **CreateResource/UpdateResource/UploadVersion:** Require course staff membership.
- **ListResources/SearchResources:** Admin/staff see all; enrolled see only Enrolled-visibility.
- **ListMembers/RemoveMember:** Require course staff or admin.

---

## 7. Scheduling & Match Lifecycle

### 7.1 Match Status Workflow

```
Draft → Scheduled → In-Progress → Final
  ↓        ↓            ↓
Canceled  Canceled    Canceled
```

Final and Canceled are terminal states.

### 7.2 Scheduling Validations

| Rule | Enforcement |
|---|---|
| Home/away balance | No team may exceed 3 consecutive home or away games |
| Round integrity | No duplicate pairings (either direction) within a round |
| Venue conflicts | No overlapping matches within 90 minutes at same venue |
| Override | Any violation requires explicit `override_reason` stored for audit |

### 7.3 Schedule Generation

Round-robin algorithm: N-1 rounds for N teams, bye handling for odd counts, round-robin venue assignment, configurable interval and start time. Each generated match passes through the same validation pipeline as manual creation.

### 7.4 Assignment Locking

Referee/staff assignments locked once match reaches In-Progress. Reassignment before that requires documented reason (Scheduler role only).

---

## 8. Courses & Teaching Resources

### 8.1 Course Outline

Tree structure via `course_outline_nodes` with `parent_id` self-referential FK and `order_index`. Node types: Chapter, Unit. Returned as nested JSON tree.

### 8.2 Resource Versioning

Immutable versions. Each upload creates a new `resource_version` with auto-incremented version_number. `latest_version_id` pointer updated atomically. SHA-256 deduplication: if same hash exists, storage path is reused.

### 8.3 Full-Text Search

PostgreSQL tsvector with weighted search: title (A), description (B), extracted text from PDF/DOCX (C). GIN-indexed. Visibility-aware: enrolled users only search Enrolled-visibility resources.

### 8.4 MIME Allowlist & Text Extraction

- 20+ allowed MIME types (documents, videos, images).
- `extracted_text` only accepted for `application/pdf` and DOCX. Rejected with 400 for other types.

---

## 9. Content Moderation & Risk Control

### 9.1 Sensitive Word Checking

Locally managed dictionaries (CRUD by Administrator). Case-insensitive matching with position-based highlighting. Returns match start/end positions and 30-character context windows.

### 9.2 Manual Review States

Pending → Approved / Rejected. Terminal states are final. Reason required for all decisions. All logged to audit trail.

### 9.3 Reports

Categories: Spam, Harassment, Inappropriate Content, Policy Violation, Other. Status: Open → Under Review → Resolved/Dismissed. Evidence attachments (local files with SHA-256). Progress notes with author tracking.

### 9.4 Anti-Spam

- 60 write operations/minute per account (WriteLimiter middleware).
- 10 reports/day per reporter (service-level enforcement).

---

## 10. Review Workflow

### 10.1 Configurable Levels

1-3 approval levels per review type, configured via `review_configs`. CHECK constraint enforces range.

### 10.2 Decision Flow

- Level approval advances to next level (or final approval if last).
- Rejection at any level terminates the entire request.
- Return-for-supplement pauses the flow; submitter resubmits with follow-up records.
- Assignee enforcement: if set, only that person can decide.

### 10.3 Disposition Write-Back

Registered callbacks at composition root. On final approval: course → Published, resource → Enrolled visibility, match → Scheduled. Unregistered entity types are no-op.

---

## 11. Payments & Reconciliation

### 11.1 Offline Ledger

Entirely offline. Records obligations and settlements in USD. Channels: Cash, Check, Wire Transfer, Internal Transfer, Journal Entry.

### 11.2 Idempotency

Client-supplied key unique per account for 24 hours. Dedicated `idempotency_keys` table with deterministic window columns. Duplicate within window returns original (HTTP 200). After window expiry, key can be reused.

### 11.3 Posting Verification

No external callbacks. Finance Clerk signs postings explicitly (`PUT /:id/sign`), setting finance_clerk_id and settled_at.

### 11.4 Failed Settlement Retries

Up to 3 retries. Failed → retry moves back to Obligation for re-signing.

### 11.5 Reconciliation

Daily summary aggregation. CSV export with all ledger fields. Reports stored with totals snapshot.

---

## 12. Audit Logs & Compliance

### 12.1 Extended Audit Entries

operator, timestamp, source, reason, before/after snapshots (JSONB), content_hash (SHA-256), workstation/device info. Per-entry integrity verification.

### 12.2 Retention Tiers

| Tier | Retention | Auto-expiry |
|---|---|---|
| access | 30 days | expires_at computed on creation |
| operation | 180 days | expires_at computed on creation |
| audit | 7 years (~2555 days) | expires_at computed on creation |

### 12.3 Tamper-Evident Hash Chain

Daily batch: `SHA-256(previous_hash | batch_hash)`. Verifiable by recomputing from entries. Entry count cross-checked.

### 12.4 Auth Path Logging

Login success/failure, token refresh, logout all logged with TierAccess. No sensitive secrets in logs.

---

## 13. Observability

### 13.1 Health Endpoint

`GET /health/detailed` — DB connectivity, connection pool stats, disk check, goroutine count, uptime.

### 13.2 Metrics Endpoint

`GET /metrics` — request count, error count, active sessions, p50/p95/p99 latency, DB pool stats, memory allocation. Ring buffer of last 1000 requests for percentile calculation.

### 13.3 Structured Logging

Echo logger configured for JSON output: timestamp, method, URI, status, latency, bytes, remote IP.

---

## 14. Deployment & Performance

### 14.1 Docker Compose

`docker compose up` starts PostgreSQL 16 + API. DB healthcheck gates API startup. Migrations run automatically. Volumes: `pgdata` (database), `filestorage` (uploads, CSVs, exports).

### 14.2 Performance Targets

- p95 < 300ms for typical queries under 200 concurrent sessions.
- Database connection retry: 30 attempts, 2s apart.
- Graceful shutdown on SIGINT/SIGTERM with 10s timeout.

### 14.3 Build

Multi-stage Docker build: Go 1.22-alpine builder → Alpine 3.19 runtime. CGO_ENABLED=0 for static binary.
