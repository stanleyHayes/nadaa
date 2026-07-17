-- Service alignment fixes from the 2026-07-17 full code review (findings #110, #111).
-- The Go services write values the original schema rejected: notification-service
-- emits voice and cell_broadcast delivery attempts with simulated/broadcast/sent
-- outcomes, alert-service persists the full approval workflow, incident-service
-- records urgency and an abuse score, and several services identify rows with
-- prefixed string IDs (usr_*, alert_*, inc_*) or the literal 'public' instead of
-- UUIDs. This migration brings the schema in line with what the services write.
-- All statements are idempotent (IF [NOT] EXISTS / DO blocks) so the file can be
-- re-applied safely.

-- (a) notification_delivery_logs channels written by notification-service
-- (allowedChannels/allowedLogChannels in internal/handlers/alerts.go):
-- push, sms, voice, cell_broadcast.
ALTER TYPE notification_channel ADD VALUE IF NOT EXISTS 'voice';
ALTER TYPE notification_channel ADD VALUE IF NOT EXISTS 'cell_broadcast';

-- (b) notification_delivery_logs statuses written by notification-service
-- providers and the cell broadcast adapter (internal/models/providers.go,
-- internal/store/cellbroadcast.go): delivered, failed, skipped, simulated,
-- sent, broadcast. 'queued' already exists and stays.
ALTER TYPE notification_delivery_status ADD VALUE IF NOT EXISTS 'broadcast';
ALTER TYPE notification_delivery_status ADD VALUE IF NOT EXISTS 'simulated';
ALTER TYPE notification_delivery_status ADD VALUE IF NOT EXISTS 'sent';

-- (c) alert-service persists the full alert workflow (AuthorityAlert in
-- services/alert-service/internal/models) but 001 only had approved_by.
ALTER TABLE IF EXISTS alerts
  ADD COLUMN IF NOT EXISTS submitted_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS approved_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS rejected_by TEXT,
  ADD COLUMN IF NOT EXISTS rejected_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS status_reason TEXT,
  ADD COLUMN IF NOT EXISTS emergency_override BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS source_prediction_id TEXT;
-- source_prediction_id links to the originating ML prediction. It is TEXT with
-- no FK: ml-service prediction IDs are service-owned strings (fixture IDs and
-- ml_log_* log IDs), not guaranteed UUIDs or rows in ml_predictions.

-- incident-service persists urgency (allowedUrgencies in
-- services/incident-service/internal/handlers/incidents.go) and names the
-- suspicious-report score abuseScore (IncidentRecord in internal/models).
ALTER TABLE IF EXISTS incidents
  ADD COLUMN IF NOT EXISTS urgency TEXT NOT NULL DEFAULT 'moderate';

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'incidents_urgency_allowed') THEN
    ALTER TABLE incidents
      ADD CONSTRAINT incidents_urgency_allowed
      CHECK (urgency IN ('low', 'moderate', 'high', 'life_threatening'));
  END IF;
END $$;

DO $$
BEGIN
  IF EXISTS (
      SELECT 1 FROM information_schema.columns
      WHERE table_name = 'incidents' AND column_name = 'suspicious_score'
    )
    AND NOT EXISTS (
      SELECT 1 FROM information_schema.columns
      WHERE table_name = 'incidents' AND column_name = 'abuse_score'
    ) THEN
    ALTER TABLE incidents RENAME COLUMN suspicious_score TO abuse_score;
  END IF;
END $$;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'incidents_suspicious_score_check')
    AND NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'incidents_abuse_score_check') THEN
    ALTER TABLE incidents RENAME CONSTRAINT incidents_suspicious_score_check TO incidents_abuse_score_check;
  END IF;
END $$;

-- (d) UUID/string-ID incompatibilities. The services identify users, alerts,
-- and incidents with prefixed string IDs (usr_*, alert_*, inc_*) and the
-- missing-person audit actor can be the literal 'public' (006 already
-- standardized created_by on TEXT). Standardize the affected columns on TEXT
-- and drop the FK constraints that no longer fit: they cannot reference the
-- UUID PKs because the services never mint UUIDs for these values.
-- Cross-service referential integrity (e.g. delivery log -> alert) stays
-- application-level, matching how the services are deployed today.

-- alerts.id: alert-service assigns IDs itself ('alert_000001'), so drop the
-- UUID default and the type. notification_delivery_logs.alert_id was already
-- TEXT; no FK is added to alerts(id) because notification-service writes
-- delivery attempts for alert-service-owned IDs that may not exist in this
-- database, so a database FK would reject legitimate rows.
ALTER TABLE IF EXISTS alerts ALTER COLUMN id DROP DEFAULT;
ALTER TABLE IF EXISTS alerts ALTER COLUMN id TYPE TEXT;

-- alerts.issued_by / alerts.approved_by hold auth-service actor IDs (usr_*),
-- not users(id) UUIDs.
ALTER TABLE IF EXISTS alerts DROP CONSTRAINT IF EXISTS alerts_issued_by_fkey;
ALTER TABLE IF EXISTS alerts ALTER COLUMN issued_by TYPE TEXT;
ALTER TABLE IF EXISTS alerts DROP CONSTRAINT IF EXISTS alerts_approved_by_fkey;
ALTER TABLE IF EXISTS alerts ALTER COLUMN approved_by TYPE TEXT;

-- missing_person_reports.related_incident_id holds incident-service IDs
-- ('inc_accra_flood_0241'), not incidents(id) UUIDs.
ALTER TABLE IF EXISTS missing_person_reports DROP CONSTRAINT IF EXISTS missing_person_reports_related_incident_id_fkey;
ALTER TABLE IF EXISTS missing_person_reports ALTER COLUMN related_incident_id TYPE TEXT;

-- missing-person reviewer/closer and audit actors are auth-service usr_* IDs
-- or the literal 'public', not users(id) UUIDs.
ALTER TABLE IF EXISTS missing_person_reports DROP CONSTRAINT IF EXISTS missing_person_reports_reviewed_by_fkey;
ALTER TABLE IF EXISTS missing_person_reports ALTER COLUMN reviewed_by TYPE TEXT;
ALTER TABLE IF EXISTS missing_person_reports DROP CONSTRAINT IF EXISTS missing_person_reports_closed_by_fkey;
ALTER TABLE IF EXISTS missing_person_reports ALTER COLUMN closed_by TYPE TEXT;
ALTER TABLE IF EXISTS missing_person_audit_entries DROP CONSTRAINT IF EXISTS missing_person_audit_entries_actor_user_id_fkey;
ALTER TABLE IF EXISTS missing_person_audit_entries ALTER COLUMN actor_user_id TYPE TEXT;
