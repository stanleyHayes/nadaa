CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS pgcrypto;

DO $$
BEGIN
  CREATE TYPE user_role AS ENUM (
    'citizen',
    'agency_viewer',
    'dispatcher',
    'responder',
    'nadmo_officer',
    'district_officer',
    'agency_admin',
    'system_admin'
  );
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

DO $$
BEGIN
  CREATE TYPE agency_type AS ENUM (
    'nadmo',
    'district_assembly',
    'police',
    'fire',
    'ambulance',
    'meteorological',
    'hydrological',
    'hospital',
    'utility',
    'ngo',
    'other'
  );
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

DO $$
BEGIN
  CREATE TYPE hazard_type AS ENUM (
    'flood',
    'fire',
    'road_crash',
    'building_collapse',
    'medical_emergency',
    'security_incident',
    'disease_outbreak',
    'electrical_hazard',
    'blocked_drain',
    'landslide',
    'marine_accident',
    'storm',
    'tidal_wave',
    'other'
  );
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

DO $$
BEGIN
  CREATE TYPE risk_level AS ENUM ('low', 'moderate', 'high', 'severe', 'emergency');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

DO $$
BEGIN
  CREATE TYPE incident_status AS ENUM (
    'reported',
    'under_review',
    'verified',
    'assigned',
    'response_en_route',
    'on_scene',
    'contained',
    'recovery_ongoing',
    'closed',
    'false_report'
  );
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

DO $$
BEGIN
  CREATE TYPE alert_severity AS ENUM ('advisory', 'watch', 'warning', 'severe_warning', 'emergency');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

DO $$
BEGIN
  CREATE TYPE alert_status AS ENUM ('draft', 'submitted', 'approved', 'rejected', 'published', 'expired', 'cancelled');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

DO $$
BEGIN
  CREATE TYPE guide_stage AS ENUM ('before', 'during', 'after', 'recovery');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS agencies (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  type agency_type NOT NULL,
  region TEXT NOT NULL,
  district TEXT NOT NULL,
  contact_number TEXT,
  service_area_geometry geometry(MultiPolygon, 4326),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  phone TEXT UNIQUE,
  email TEXT UNIQUE,
  role user_role NOT NULL DEFAULT 'citizen',
  agency_id UUID REFERENCES agencies(id),
  preferred_language TEXT NOT NULL DEFAULT 'en',
  home_location geometry(Point, 4326),
  contact_permission BOOLEAN NOT NULL DEFAULT false,
  mfa_required BOOLEAN NOT NULL DEFAULT false,
  mfa_enabled BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT users_authority_agency_required CHECK (
    role = 'citizen' OR agency_id IS NOT NULL
  )
);

CREATE TABLE IF NOT EXISTS incidents (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  reference TEXT UNIQUE NOT NULL,
  type hazard_type NOT NULL,
  severity risk_level NOT NULL DEFAULT 'moderate',
  status incident_status NOT NULL DEFAULT 'reported',
  description TEXT NOT NULL,
  location_geometry geometry(Point, 4326) NOT NULL,
  reported_by UUID REFERENCES users(id),
  anonymous BOOLEAN NOT NULL DEFAULT false,
  contact_permission BOOLEAN NOT NULL DEFAULT false,
  verified_by UUID REFERENCES users(id),
  assigned_agency_id UUID REFERENCES agencies(id),
  people_affected INTEGER NOT NULL DEFAULT 0 CHECK (people_affected >= 0),
  injuries_reported BOOLEAN NOT NULL DEFAULT false,
  accessibility_needs TEXT,
  suspicious_score NUMERIC(5, 4) NOT NULL DEFAULT 0 CHECK (suspicious_score >= 0 AND suspicious_score <= 1),
  duplicate_of_incident_id UUID REFERENCES incidents(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS incident_media (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  incident_id UUID NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
  media_url TEXT NOT NULL,
  media_type TEXT NOT NULL,
  uploaded_by UUID REFERENCES users(id),
  content_type TEXT,
  size_bytes BIGINT CHECK (size_bytes IS NULL OR size_bytes >= 0),
  checksum TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS incident_timeline_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  incident_id UUID NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
  actor_user_id UUID REFERENCES users(id),
  actor_agency_id UUID REFERENCES agencies(id),
  event_type TEXT NOT NULL,
  note TEXT,
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS alerts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  title TEXT NOT NULL,
  message TEXT NOT NULL,
  hazard_type hazard_type NOT NULL,
  severity alert_severity NOT NULL,
  target_geometry geometry(MultiPolygon, 4326) NOT NULL,
  target_label TEXT NOT NULL,
  issued_by UUID REFERENCES users(id),
  approved_by UUID REFERENCES users(id),
  agency_id UUID REFERENCES agencies(id),
  starts_at TIMESTAMPTZ NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  status alert_status NOT NULL DEFAULT 'draft',
  recommended_action TEXT,
  evacuation_required BOOLEAN NOT NULL DEFAULT false,
  shelter_links JSONB NOT NULL DEFAULT '[]'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT alerts_expiry_after_start CHECK (expires_at > starts_at)
);

CREATE TABLE IF NOT EXISTS risk_zones (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  hazard_type hazard_type NOT NULL,
  risk_level risk_level NOT NULL,
  geometry geometry(MultiPolygon, 4326) NOT NULL,
  source TEXT NOT NULL,
  valid_from TIMESTAMPTZ NOT NULL DEFAULT now(),
  valid_to TIMESTAMPTZ,
  explanation TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS shelters (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  location_geometry geometry(Point, 4326) NOT NULL,
  region TEXT NOT NULL,
  district TEXT NOT NULL,
  capacity INTEGER CHECK (capacity IS NULL OR capacity >= 0),
  current_occupancy INTEGER NOT NULL DEFAULT 0 CHECK (current_occupancy >= 0),
  contact TEXT,
  facilities JSONB NOT NULL DEFAULT '[]'::jsonb,
  status TEXT NOT NULL DEFAULT 'open',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT shelters_occupancy_within_capacity CHECK (
    capacity IS NULL OR current_occupancy <= capacity
  )
);

CREATE TABLE IF NOT EXISTS emergency_guides (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  hazard_type hazard_type NOT NULL,
  stage guide_stage NOT NULL,
  title TEXT NOT NULL,
  body TEXT NOT NULL,
  language TEXT NOT NULL DEFAULT 'en',
  offline_available BOOLEAN NOT NULL DEFAULT true,
  sort_order INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS ml_predictions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  hazard_type hazard_type NOT NULL,
  model_version TEXT NOT NULL,
  prediction_time TIMESTAMPTZ NOT NULL,
  target_time TIMESTAMPTZ NOT NULL,
  geometry geometry(MultiPolygon, 4326) NOT NULL,
  probability NUMERIC(5, 4) NOT NULL CHECK (probability >= 0 AND probability <= 1),
  severity risk_level NOT NULL,
  confidence TEXT NOT NULL,
  explanation JSONB NOT NULL DEFAULT '[]'::jsonb,
  input_feature_set_version TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS weather_observations (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  source TEXT NOT NULL,
  observed_at TIMESTAMPTZ NOT NULL,
  valid_from TIMESTAMPTZ NOT NULL,
  valid_to TIMESTAMPTZ,
  location_geometry geometry(Point, 4326) NOT NULL,
  rainfall_mm NUMERIC(10, 2),
  water_level_m NUMERIC(10, 2),
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS audit_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  actor_user_id UUID REFERENCES users(id),
  actor_agency_id UUID REFERENCES agencies(id),
  actor_role user_role,
  action TEXT NOT NULL,
  target_type TEXT NOT NULL,
  target_id TEXT,
  request_id TEXT,
  ip_address INET,
  user_agent TEXT,
  before JSONB,
  after JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_agencies_service_area_geometry ON agencies USING GIST (service_area_geometry);
CREATE INDEX IF NOT EXISTS idx_users_home_location ON users USING GIST (home_location);
CREATE INDEX IF NOT EXISTS idx_users_role ON users (role);
CREATE INDEX IF NOT EXISTS idx_incidents_location_geometry ON incidents USING GIST (location_geometry);
CREATE INDEX IF NOT EXISTS idx_incidents_status_created_at ON incidents (status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_incidents_type_severity ON incidents (type, severity);
CREATE INDEX IF NOT EXISTS idx_incident_media_incident_id ON incident_media (incident_id);
CREATE INDEX IF NOT EXISTS idx_timeline_incident_created_at ON incident_timeline_events (incident_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_alerts_target_geometry ON alerts USING GIST (target_geometry);
CREATE INDEX IF NOT EXISTS idx_alerts_status_expires_at ON alerts (status, expires_at);
CREATE INDEX IF NOT EXISTS idx_risk_zones_geometry ON risk_zones USING GIST (geometry);
CREATE INDEX IF NOT EXISTS idx_risk_zones_hazard_level ON risk_zones (hazard_type, risk_level);
CREATE INDEX IF NOT EXISTS idx_shelters_location_geometry ON shelters USING GIST (location_geometry);
CREATE INDEX IF NOT EXISTS idx_guides_hazard_stage_language ON emergency_guides (hazard_type, stage, language);
CREATE INDEX IF NOT EXISTS idx_ml_predictions_geometry ON ml_predictions USING GIST (geometry);
CREATE INDEX IF NOT EXISTS idx_ml_predictions_hazard_target_time ON ml_predictions (hazard_type, target_time DESC);
CREATE INDEX IF NOT EXISTS idx_weather_observations_location ON weather_observations USING GIST (location_geometry);
CREATE INDEX IF NOT EXISTS idx_weather_observations_source_time ON weather_observations (source, observed_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_created_at ON audit_logs (actor_user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_target ON audit_logs (target_type, target_id);
