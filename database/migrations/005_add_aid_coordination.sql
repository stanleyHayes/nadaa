-- Donation and aid coordination for NADAA-123.
-- Aid requests are operational recovery needs; pledges are partner commitments
-- tracked separately from incident status and dispatch state.

CREATE TABLE IF NOT EXISTS aid_requests (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  title TEXT NOT NULL,
  category TEXT NOT NULL CHECK (category IN ('food', 'water', 'medical', 'hygiene', 'shelter', 'logistics', 'cash', 'equipment', 'volunteers', 'other')),
  priority TEXT NOT NULL CHECK (priority IN ('low', 'medium', 'high', 'urgent')),
  status TEXT NOT NULL CHECK (status IN ('pending_review', 'approved', 'open', 'partially_matched', 'fulfilled', 'paused', 'closed', 'rejected')) DEFAULT 'pending_review',
  region TEXT,
  district TEXT,
  geometry geometry(Point, 4326) NOT NULL,
  receiving_organization TEXT NOT NULL,
  contact TEXT,
  quantity_needed INTEGER NOT NULL CHECK (quantity_needed > 0),
  quantity_unit TEXT NOT NULL,
  quantity_pledged INTEGER NOT NULL DEFAULT 0 CHECK (quantity_pledged >= 0),
  description TEXT NOT NULL,
  needed_by TIMESTAMPTZ NOT NULL,
  visibility TEXT NOT NULL CHECK (visibility IN ('public', 'partners_only')) DEFAULT 'public',
  source_relief_point_id UUID REFERENCES relief_points(id),
  created_by UUID REFERENCES users(id),
  approved_by UUID REFERENCES users(id),
  approval_notes TEXT,
  anti_fraud_notes TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_aid_requests_geometry ON aid_requests USING GIST (geometry);
CREATE INDEX IF NOT EXISTS idx_aid_requests_public_status ON aid_requests (visibility, status, priority, needed_by);
CREATE INDEX IF NOT EXISTS idx_aid_requests_category_district ON aid_requests (category, district);

CREATE TABLE IF NOT EXISTS aid_pledges (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  aid_request_id UUID NOT NULL REFERENCES aid_requests(id) ON DELETE CASCADE,
  donor_name TEXT NOT NULL,
  donor_type TEXT NOT NULL CHECK (donor_type IN ('individual', 'business', 'ngo', 'faith_group', 'diaspora', 'government', 'other')),
  contact TEXT NOT NULL,
  quantity INTEGER NOT NULL CHECK (quantity > 0),
  unit TEXT NOT NULL,
  note TEXT,
  status TEXT NOT NULL CHECK (status IN ('pledged', 'accepted', 'received', 'cancelled', 'flagged')) DEFAULT 'pledged',
  review_status TEXT NOT NULL CHECK (review_status IN ('pending_review', 'cleared', 'flagged')) DEFAULT 'pending_review',
  fraud_review_notes TEXT,
  reviewed_by UUID REFERENCES users(id),
  pledged_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_aid_pledges_request_status ON aid_pledges (aid_request_id, status, review_status);
CREATE INDEX IF NOT EXISTS idx_aid_pledges_review_status ON aid_pledges (review_status, pledged_at DESC);
