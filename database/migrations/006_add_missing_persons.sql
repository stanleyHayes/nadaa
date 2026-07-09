-- Missing persons and family reunification workflow for NADAA-132.
-- Public visibility is controlled by authority review; reporter details and
-- closure notes are sensitive and must not be exposed in public search.

CREATE TABLE IF NOT EXISTS missing_person_reports (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  reference TEXT NOT NULL UNIQUE,
  person_name TEXT NOT NULL,
  age INTEGER CHECK (age >= 0 AND age <= 120),
  gender TEXT CHECK (gender IN ('female', 'male', 'non_binary', 'unknown')),
  description TEXT NOT NULL,
  photo_url TEXT,
  last_seen_at TIMESTAMPTZ NOT NULL,
  last_seen_label TEXT NOT NULL,
  region TEXT NOT NULL,
  district TEXT NOT NULL,
  geometry geometry(Point, 4326),
  related_incident_id UUID REFERENCES incidents(id),
  reporter_name TEXT NOT NULL,
  reporter_phone TEXT NOT NULL,
  reporter_email TEXT,
  reporter_relationship TEXT NOT NULL,
  consent_to_contact BOOLEAN NOT NULL DEFAULT true,
  consent_to_public_share BOOLEAN NOT NULL DEFAULT false,
  status TEXT NOT NULL CHECK (status IN ('pending_review', 'active', 'located', 'reunited', 'closed', 'rejected')) DEFAULT 'pending_review',
  review_status TEXT NOT NULL CHECK (review_status IN ('pending', 'approved', 'rejected')) DEFAULT 'pending',
  public_visibility TEXT NOT NULL CHECK (public_visibility IN ('private', 'public')) DEFAULT 'private',
  public_summary TEXT,
  review_notes TEXT,
  closure_type TEXT CHECK (closure_type IN ('reunited', 'located_safe', 'duplicate', 'withdrawn', 'deceased', 'other')),
  closure_notes TEXT,
  created_by TEXT NOT NULL DEFAULT 'public',
  reviewed_by UUID REFERENCES users(id),
  reviewed_at TIMESTAMPTZ,
  closed_by UUID REFERENCES users(id),
  closed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_missing_person_reports_geometry ON missing_person_reports USING GIST (geometry);
CREATE INDEX IF NOT EXISTS idx_missing_person_reports_public_status ON missing_person_reports (public_visibility, review_status, status, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_missing_person_reports_district ON missing_person_reports (district, status, review_status);

CREATE TABLE IF NOT EXISTS missing_person_audit_entries (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  missing_person_report_id UUID NOT NULL REFERENCES missing_person_reports(id) ON DELETE CASCADE,
  action TEXT NOT NULL,
  actor_user_id UUID REFERENCES users(id),
  actor_agency_id UUID REFERENCES agencies(id),
  actor_role TEXT,
  notes TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_missing_person_audit_entries_report ON missing_person_audit_entries (missing_person_report_id, created_at);
