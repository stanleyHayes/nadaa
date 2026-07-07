CREATE TABLE IF NOT EXISTS road_closures (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  road_name TEXT NOT NULL,
  reason TEXT,
  status TEXT NOT NULL CHECK (status IN ('active', 'scheduled', 'lifted', 'cancelled')),
  severity TEXT NOT NULL CHECK (severity IN ('low', 'moderate', 'high', 'severe', 'emergency')),
  source TEXT NOT NULL,
  source_ref TEXT,
  geometry geometry(LineString, 4326) NOT NULL,
  valid_from TIMESTAMPTZ NOT NULL DEFAULT now(),
  valid_to TIMESTAMPTZ,
  detour_note TEXT,
  created_by UUID REFERENCES users(id),
  updated_by UUID REFERENCES users(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_road_closures_geometry ON road_closures USING GIST (geometry);
CREATE INDEX IF NOT EXISTS idx_road_closures_status_valid ON road_closures (status, valid_from, valid_to);
