-- Relief distribution points for NADAA-122.
-- Relief points are managed by authorities and consumed by citizens and dispatchers
-- as a recovery-context layer alongside shelters and road closures.

CREATE TABLE IF NOT EXISTS relief_points (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  type TEXT NOT NULL CHECK (type IN ('food', 'water', 'medical', 'hygiene', 'blankets', 'cash', 'mixed')),
  region TEXT,
  district TEXT,
  address TEXT,
  geometry geometry(Point, 4326) NOT NULL,
  contact TEXT,
  operating_hours TEXT,
  eligibility TEXT,
  schedule TEXT,
  stock_categories JSONB NOT NULL DEFAULT '[]'::jsonb,
  status TEXT NOT NULL CHECK (status IN ('open', 'limited', 'closed', 'paused')) DEFAULT 'open',
  source TEXT NOT NULL DEFAULT 'manual',
  source_ref TEXT,
  created_by UUID REFERENCES users(id),
  updated_by UUID REFERENCES users(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_relief_points_geometry ON relief_points USING GIST (geometry);
CREATE INDEX IF NOT EXISTS idx_relief_points_status_type ON relief_points (status, type);

-- Stock history captures snapshots each time an authority updates stock quantities.
CREATE TABLE IF NOT EXISTS relief_point_stock_history (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  relief_point_id UUID NOT NULL REFERENCES relief_points(id) ON DELETE CASCADE,
  changed_by UUID REFERENCES users(id),
  changed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  note TEXT,
  stock_categories JSONB NOT NULL DEFAULT '[]'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_relief_point_stock_history_relief_point_id ON relief_point_stock_history (relief_point_id, changed_at DESC);
