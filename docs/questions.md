# Collegiate Athletics & Learning Ops API — Open Questions

This document captures unresolved business-logic questions for the Collegiate Athletics & Learning Ops backend API (Go / Echo + sqlx, PostgreSQL, single-node Docker, fully offline). Each entry records the ambiguity, our working assumption, and the solution implemented unless stakeholders direct otherwise.

---

## 1. Match Scheduling — Venue Conflict Window Duration

**Question:** The 90-minute venue conflict window prevents overlapping matches. Is 90 minutes sufficient for all venue types (indoor arena vs outdoor field)? Should the window be configurable per venue?

**My Understanding:** 90 minutes is a reasonable default that accounts for match duration plus changeover time. Different venue types may need different windows, but a single fixed value keeps the scheduling logic simpler and more predictable for the initial release.

**Solution:** We implement a fixed 90-minute conflict window in the SQL query (`ABS(EXTRACT(EPOCH FROM (scheduled_at - $2::timestamptz))) < 5400`). The value is a constant in the repository layer. If per-venue configuration is needed in the future, a `conflict_window_minutes` column can be added to the `venues` table and the query can reference it. For now, the fixed window satisfies the stated requirement without introducing unnecessary complexity.

---

## 2. Schedule Generation — Odd Number of Teams

**Question:** When a season has an odd number of teams, the round-robin algorithm needs to handle byes. Should bye rounds be recorded as matches, or simply skipped? How does a bye affect the consecutive home/away tracking?

**My Understanding:** Byes should be invisible at the database level — no "bye match" records exist. The team that has a bye in a given round simply has no match that round, which naturally resets any consecutive home/away streak since the streak is calculated from actual scheduled matches.

**Solution:** The `GenerateSchedule` method adds a virtual nil team when the team count is odd. Pairings involving the nil team are silently skipped — no match record is created. Consecutive home/away balance checks query actual matches only, so a bye week does not count as either home or away. The response reports only the actually created matches.

---

## 3. Idempotency Key — Reuse After 24 Hours

**Question:** After the 24-hour idempotency window expires, should the system allow the same key to be reused for a completely different payment? What happens if the original payment is still in Obligation status when the window expires?

**My Understanding:** The idempotency key is a deduplication mechanism, not a permanent identifier. After 24 hours the key should be freely reusable regardless of the original payment's status. The 24-hour window covers retry scenarios; any reuse beyond that is intentional.

**Solution:** We use a dedicated `idempotency_keys` table with explicit `window_start` and `window_end` columns and a unique index on `(account_id, idempotency_key)`. Expired keys are cleaned up by `DeleteExpiredIdempotencyKeys`. After cleanup, the unique index slot is freed and the same key can be inserted again. The original payment remains unaffected — it is a standalone ledger entry. The service layer's `FindActiveIdempotencyKey` uses a time-bounded query (`WHERE window_end > $now`) so even if cleanup hasn't run yet, an expired key won't match.

---

## 4. Review Workflow — Disposition Write-Back Failure

**Question:** When a review reaches final approval and the disposition callback fires (e.g., publishing a course), what happens if the target entity has been deleted or is in an unexpected state? Should the review decision be rolled back?

**My Understanding:** The disposition write-back is a side effect of the review decision, not a precondition. Rolling back the review because the target entity is in an unexpected state would create confusing semantics. The review decision itself is the authoritative record; the write-back is best-effort.

**Solution:** The disposition callback executes outside the review decision's database transaction. If the callback fails (target deleted, constraint violation, etc.), the review decision stands as committed. The failure is silently swallowed (same pattern as audit logging). Administrators can inspect the review's `final_decision` and the target entity's state independently. For unregistered entity types, the callback map returns nothing — a no-op by design, not an error. This avoids coupling the review workflow's correctness to the state of external entities.

---

## 5. Course Membership — Self-Enrollment

**Question:** Can users enroll themselves in a course, or must a course staff member add them? Should there be an enrollment request workflow?

**My Understanding:** For a governed educational environment, self-enrollment without staff approval creates access control risks. Course staff should explicitly manage membership. An enrollment request workflow could be built on top of the review system in the future.

**Solution:** Only users with course staff membership (or Administrators) can add members via `POST /api/courses/:course_id/members`. There is no self-enrollment endpoint. The review workflow can be used to model enrollment requests if needed: a user submits a review request of type "enrollment" targeting the course, and staff approve/reject. This is a policy-level decision that can be implemented without API changes.

---

## 6. Resource Visibility — Transitioning from Staff to Enrolled

**Question:** When a resource's visibility changes from Staff to Enrolled, should existing enrolled users be notified? Should the change be reversible?

**My Understanding:** Visibility changes are straightforward metadata updates. No notification system exists in the current scope, and reversibility is naturally supported since the visibility field can be set back to Staff at any time.

**Solution:** Visibility is a mutable field on the resource. Staff can change it freely via `PUT /api/resources/:id`. The disposition write-back from the review workflow sets visibility to Enrolled on final approval. There is no notification mechanism — this is outside scope for a backend-only API. Reverting visibility back to Staff is permitted at any time by course staff. All visibility changes are audit-logged with before/after snapshots.

---

## 7. Audit Log Retention — Tier Assignment for Domain Events

**Question:** Which tier should domain events fall into? For example, is a match status transition an "operation" or an "audit" event? Should the tier be configurable per event type?

**My Understanding:** The three tiers map to sensitivity and compliance requirements: access (auth events, short retention), operation (routine CRUD, medium retention), audit (compliance-critical, long retention). Most domain events are operations; events with compliance significance (payments, reviews, status transitions with financial or legal impact) are audit-tier.

**Solution:** The default tier for all `Log()` calls is `TierOperation` (180-day retention). Auth events (login/logout/refresh) explicitly use `TierAccess` (30-day). Payment signing, course publication via review disposition, and any operation using `LogExtended()` with explicit `Tier: TierAudit` get 7-year retention. The tier is set at the call site, not configurable per event type at runtime. If per-event configurability is needed, a lookup table mapping (entity_type, action) to tier could be added without changing the logging API.

---

## 8. Sensitive Word Matching — Unicode and Multi-Language Support

**Question:** The current implementation uses case-insensitive string matching via `strings.ToLower`. Does the system need to handle Unicode normalization, diacritics, or non-Latin scripts?

**My Understanding:** For a US-based collegiate athletics department, English is the primary language. Basic case-insensitive matching via Go's `strings.ToLower` handles ASCII and most common Unicode codepoints correctly. Full Unicode normalization (NFD/NFC) and script-specific matching are out of scope.

**Solution:** Words are lowercased on storage and searched with `strings.ToLower` on the input text. This handles English and common Western European characters. Words are matched longest-first (sorted by length descending) to avoid partial matches of longer phrases. No Unicode normalization or stemming is applied. If multi-language support is needed, the matching logic can be replaced with a proper NLP tokenizer without changing the dictionary management API.

---

## 9. Reconciliation CSV — Concurrent Generation

**Question:** If two Finance Clerks generate reconciliation reports for the same date simultaneously, do they get separate reports or should the operation be idempotent per date?

**My Understanding:** Reconciliation is a snapshot operation — generating twice for the same date should produce two separate reports, since the ledger state may have changed between the two calls (e.g., a payment was signed in between). Each report captures a point-in-time view.

**Solution:** Each `POST /api/reconciliation/reports` call creates a new `reconciliation_reports` row and a new CSV file, even for the same date. There is no uniqueness constraint on `report_date`. This allows Finance Clerks to generate updated snapshots as needed. The `ListReconciliationReports` endpoint shows all generated reports sorted by creation date, so auditors can see the full history of reconciliation snapshots.

---

## 10. Password Change — Session Invalidation Scope

**Question:** When a user changes their password, should only their other sessions be invalidated, or all sessions including the current one?

**My Understanding:** All sessions should be invalidated on password change. This is the safest approach — if the password was changed because of a suspected compromise, all existing sessions (including potentially compromised ones) should be force-expired.

**Solution:** `ChangePassword` calls `RevokeAllForAccount` which sets `revoked = TRUE` on all refresh tokens for the account. This invalidates all sessions on all devices. The current session's access token remains valid until its 30-minute expiry (since JWTs are stateless), but the refresh token cannot be used to obtain a new access token. The user must re-authenticate on all devices with the new password.
