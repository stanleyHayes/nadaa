-- Remaining service alignment fixes from the 2026-07-20 full code review.
-- 007 aligned alerts, notification enums, and the missing-person tables; this
-- migration closes the two gaps it left behind:
--   (a) UUID/string-ID misalignment in the central tables. The services mint
--       prefixed string IDs (usr_*, inc_*, road_closure_*, relief_*, aid_*)
--       that can never cast to UUID, so the remaining UUID primary keys and
--       users(id) foreign keys would reject every service-written row.
--       Standardize the affected columns on TEXT, drop the gen_random_uuid()
--       defaults, and drop the FK constraints that no longer fit — matching
--       007: cross-service referential integrity stays application-level.
--   (b) incident-service persists abuse-review and status-lifecycle fields
--       (IncidentRecord in services/incident-service/internal/models) that
--       have no columns yet.
-- All statements are idempotent (IF [NOT] EXISTS / DO blocks) so the file can
-- be re-applied safely.

-- (a1) Drop the FK constraints that no longer fit. That includes the
-- parent/child FKs into the converted PKs (incident_media.incident_id,
-- incident_timeline_events.incident_id, relief_point_stock_history.
-- relief_point_id, aid_pledges.aid_request_id): PostgreSQL does not widen the
-- referencing column together with the PK — ALTER COLUMN ... TYPE TEXT fails
-- with "incompatible types" while the FK exists. Both sides are standardized
-- on TEXT below and, per 007, referential integrity stays application-level.
ALTER TABLE IF EXISTS incidents DROP CONSTRAINT IF EXISTS incidents_reported_by_fkey;
ALTER TABLE IF EXISTS incidents DROP CONSTRAINT IF EXISTS incidents_verified_by_fkey;
ALTER TABLE IF EXISTS incidents DROP CONSTRAINT IF EXISTS incidents_duplicate_of_incident_id_fkey;
ALTER TABLE IF EXISTS incident_timeline_events DROP CONSTRAINT IF EXISTS incident_timeline_events_incident_id_fkey;
ALTER TABLE IF EXISTS incident_timeline_events DROP CONSTRAINT IF EXISTS incident_timeline_events_actor_user_id_fkey;
ALTER TABLE IF EXISTS incident_media DROP CONSTRAINT IF EXISTS incident_media_incident_id_fkey;
ALTER TABLE IF EXISTS incident_media DROP CONSTRAINT IF EXISTS incident_media_uploaded_by_fkey;
ALTER TABLE IF EXISTS audit_logs DROP CONSTRAINT IF EXISTS audit_logs_actor_user_id_fkey;
ALTER TABLE IF EXISTS road_closures DROP CONSTRAINT IF EXISTS road_closures_created_by_fkey;
ALTER TABLE IF EXISTS road_closures DROP CONSTRAINT IF EXISTS road_closures_updated_by_fkey;
ALTER TABLE IF EXISTS relief_points DROP CONSTRAINT IF EXISTS relief_points_created_by_fkey;
ALTER TABLE IF EXISTS relief_points DROP CONSTRAINT IF EXISTS relief_points_updated_by_fkey;
ALTER TABLE IF EXISTS relief_point_stock_history DROP CONSTRAINT IF EXISTS relief_point_stock_history_relief_point_id_fkey;
ALTER TABLE IF EXISTS relief_point_stock_history DROP CONSTRAINT IF EXISTS relief_point_stock_history_changed_by_fkey;
ALTER TABLE IF EXISTS aid_requests DROP CONSTRAINT IF EXISTS aid_requests_created_by_fkey;
ALTER TABLE IF EXISTS aid_requests DROP CONSTRAINT IF EXISTS aid_requests_approved_by_fkey;
ALTER TABLE IF EXISTS aid_requests DROP CONSTRAINT IF EXISTS aid_requests_source_relief_point_id_fkey;
ALTER TABLE IF EXISTS aid_pledges DROP CONSTRAINT IF EXISTS aid_pledges_aid_request_id_fkey;
ALTER TABLE IF EXISTS aid_pledges DROP CONSTRAINT IF EXISTS aid_pledges_reviewed_by_fkey;

-- (a2) incidents.id: incident-service assigns IDs itself ('inc_*'), so drop
-- the UUID default and the type. reported_by/verified_by hold auth-service
-- actor IDs (usr_*) or stay empty for anonymous reports; duplicate_of_incident_id
-- holds an incident-service ID, so its self-FK goes as well.
ALTER TABLE IF EXISTS incidents ALTER COLUMN id DROP DEFAULT;
ALTER TABLE IF EXISTS incidents ALTER COLUMN id TYPE TEXT;
ALTER TABLE IF EXISTS incidents ALTER COLUMN reported_by TYPE TEXT;
ALTER TABLE IF EXISTS incidents ALTER COLUMN verified_by TYPE TEXT;
ALTER TABLE IF EXISTS incidents ALTER COLUMN duplicate_of_incident_id TYPE TEXT;

-- users.id: auth-service assigns IDs itself ('usr_*').
ALTER TABLE IF EXISTS users ALTER COLUMN id DROP DEFAULT;
ALTER TABLE IF EXISTS users ALTER COLUMN id TYPE TEXT;

-- Timeline/audit/media actors are auth-service usr_* IDs or the literal
-- 'public', not users(id) UUIDs; the timeline/media parent columns hold
-- incident-service inc_* IDs.
ALTER TABLE IF EXISTS incident_timeline_events ALTER COLUMN incident_id TYPE TEXT;
ALTER TABLE IF EXISTS incident_timeline_events ALTER COLUMN actor_user_id TYPE TEXT;
ALTER TABLE IF EXISTS incident_media ALTER COLUMN incident_id TYPE TEXT;
ALTER TABLE IF EXISTS incident_media ALTER COLUMN uploaded_by TYPE TEXT;
ALTER TABLE IF EXISTS audit_logs ALTER COLUMN actor_user_id TYPE TEXT;

-- road-closure-service assigns IDs itself ('road_closure_*'); created_by /
-- updated_by are auth-service usr_* IDs.
ALTER TABLE IF EXISTS road_closures ALTER COLUMN id DROP DEFAULT;
ALTER TABLE IF EXISTS road_closures ALTER COLUMN id TYPE TEXT;
ALTER TABLE IF EXISTS road_closures ALTER COLUMN created_by TYPE TEXT;
ALTER TABLE IF EXISTS road_closures ALTER COLUMN updated_by TYPE TEXT;

-- shelter-service assigns relief point IDs itself ('relief_*'); created_by /
-- updated_by / stock-history changed_by are auth-service usr_* IDs, and the
-- stock-history parent column holds a 'relief_*' ID.
ALTER TABLE IF EXISTS relief_points ALTER COLUMN id DROP DEFAULT;
ALTER TABLE IF EXISTS relief_points ALTER COLUMN id TYPE TEXT;
ALTER TABLE IF EXISTS relief_points ALTER COLUMN created_by TYPE TEXT;
ALTER TABLE IF EXISTS relief_points ALTER COLUMN updated_by TYPE TEXT;
ALTER TABLE IF EXISTS relief_point_stock_history ALTER COLUMN relief_point_id TYPE TEXT;
ALTER TABLE IF EXISTS relief_point_stock_history ALTER COLUMN changed_by TYPE TEXT;

-- Aid request IDs are service-assigned ('aid_*'); created_by / approved_by /
-- reviewed_by are auth-service usr_* IDs, source_relief_point_id holds a
-- shelter-service 'relief_*' ID, and the pledge parent column holds an
-- 'aid_*' ID.
ALTER TABLE IF EXISTS aid_requests ALTER COLUMN id DROP DEFAULT;
ALTER TABLE IF EXISTS aid_requests ALTER COLUMN id TYPE TEXT;
ALTER TABLE IF EXISTS aid_requests ALTER COLUMN created_by TYPE TEXT;
ALTER TABLE IF EXISTS aid_requests ALTER COLUMN approved_by TYPE TEXT;
ALTER TABLE IF EXISTS aid_requests ALTER COLUMN source_relief_point_id TYPE TEXT;
ALTER TABLE IF EXISTS aid_pledges ALTER COLUMN aid_request_id TYPE TEXT;
ALTER TABLE IF EXISTS aid_pledges ALTER COLUMN reviewed_by TYPE TEXT;

-- (b) incident-service persists abuse-review outcomes and status-lifecycle
-- fields on IncidentRecord (internal/models). Decisions come from
-- allowedAbuseReviewDecisions in internal/handlers/incidents.go. All optional
-- on the record, so the columns stay nullable except abuse_review_required,
-- which the service always writes. merged_into_id is TEXT with no FK for the
-- same reason as duplicate_of_incident_id above.
ALTER TABLE IF EXISTS incidents
  ADD COLUMN IF NOT EXISTS abuse_review_required BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS abuse_review_reason TEXT,
  ADD COLUMN IF NOT EXISTS abuse_review_decision TEXT,
  ADD COLUMN IF NOT EXISTS abuse_reviewed_by TEXT,
  ADD COLUMN IF NOT EXISTS abuse_reviewed_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS verified_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS status_updated_by TEXT,
  ADD COLUMN IF NOT EXISTS status_reason TEXT,
  ADD COLUMN IF NOT EXISTS resolution_notes TEXT,
  ADD COLUMN IF NOT EXISTS closed_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS merged_into_id TEXT,
  ADD COLUMN IF NOT EXISTS merged_by TEXT,
  ADD COLUMN IF NOT EXISTS merged_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS merge_reason TEXT;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'incidents_abuse_review_decision_allowed') THEN
    ALTER TABLE incidents
      ADD CONSTRAINT incidents_abuse_review_decision_allowed
      CHECK (abuse_review_decision IN ('clear', 'monitor', 'false_report'));
  END IF;
END $$;
