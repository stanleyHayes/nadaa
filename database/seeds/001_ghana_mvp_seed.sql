INSERT INTO agencies (id, name, type, region, district, contact_number, service_area_geometry)
VALUES
  (
    '00000000-0000-0000-0000-000000000101',
    'NADMO Accra Metro',
    'nadmo',
    'Greater Accra',
    'Accra Metropolitan',
    '112',
    ST_Multi(ST_GeomFromText('POLYGON((-0.250 5.520, -0.110 5.520, -0.110 5.650, -0.250 5.650, -0.250 5.520))', 4326))
  ),
  (
    '00000000-0000-0000-0000-000000000102',
    'Ghana National Fire Service Accra',
    'fire',
    'Greater Accra',
    'Accra Metropolitan',
    '112',
    ST_Multi(ST_GeomFromText('POLYGON((-0.260 5.510, -0.090 5.510, -0.090 5.660, -0.260 5.660, -0.260 5.510))', 4326))
  ),
  (
    '00000000-0000-0000-0000-000000000103',
    'National Ambulance Service Accra',
    'ambulance',
    'Greater Accra',
    'Accra Metropolitan',
    '112',
    ST_Multi(ST_GeomFromText('POLYGON((-0.280 5.500, -0.080 5.500, -0.080 5.670, -0.280 5.670, -0.280 5.500))', 4326))
  )
ON CONFLICT (id) DO NOTHING;

INSERT INTO users (id, name, phone, email, role, agency_id, preferred_language, contact_permission, mfa_required, mfa_enabled)
VALUES
  (
    '00000000-0000-0000-0000-000000000201',
    'NADAA System Admin',
    '+233200000001',
    'admin@nadaa.local',
    'system_admin',
    '00000000-0000-0000-0000-000000000101',
    'en',
    false,
    true,
    true
  ),
  (
    '00000000-0000-0000-0000-000000000202',
    'Accra Dispatcher',
    '+233200000002',
    'dispatcher@nadaa.local',
    'dispatcher',
    '00000000-0000-0000-0000-000000000101',
    'en',
    false,
    true,
    true
  ),
  (
    '00000000-0000-0000-0000-000000000203',
    'Ama Mensah',
    '+233200000003',
    'ama@example.local',
    'citizen',
    NULL,
    'en',
    true,
    false,
    false
  )
ON CONFLICT (id) DO NOTHING;

INSERT INTO shelters (id, name, location_geometry, region, district, capacity, current_occupancy, contact, facilities, status)
VALUES
  (
    '00000000-0000-0000-0000-000000000301',
    'Accra Metro Assembly Shelter',
    ST_GeomFromText('POINT(-0.200 5.560)', 4326),
    'Greater Accra',
    'Accra Metropolitan',
    450,
    116,
    '112',
    '["water", "first_aid", "accessible_entry", "family_area"]'::jsonb,
    'open'
  ),
  (
    '00000000-0000-0000-0000-000000000302',
    'Osu Community Hall',
    ST_GeomFromText('POINT(-0.180 5.550)', 4326),
    'Greater Accra',
    'Accra Metropolitan',
    220,
    34,
    '112',
    '["water", "first_aid"]'::jsonb,
    'open'
  )
ON CONFLICT (id) DO NOTHING;

INSERT INTO risk_zones (id, hazard_type, risk_level, geometry, source, valid_from, valid_to, explanation)
VALUES
  (
    '00000000-0000-0000-0000-000000000401',
    'flood',
    'severe',
    ST_Multi(ST_GeomFromText('POLYGON((-0.230 5.530, -0.160 5.530, -0.160 5.590, -0.230 5.590, -0.230 5.530))', 4326)),
    'development_fixture',
    now(),
    NULL,
    'Low-lying Accra sample zone with historical flood reports and rainfall sensitivity.'
  ),
  (
    '00000000-0000-0000-0000-000000000402',
    'fire',
    'moderate',
    ST_Multi(ST_GeomFromText('POLYGON((-0.210 5.540, -0.140 5.540, -0.140 5.610, -0.210 5.610, -0.210 5.540))', 4326)),
    'development_fixture',
    now(),
    NULL,
    'Dense commercial area sample zone.'
  )
ON CONFLICT (id) DO NOTHING;

INSERT INTO emergency_guides (id, hazard_type, stage, title, body, language, offline_available, sort_order)
VALUES
  (
    '00000000-0000-0000-0000-000000000501',
    'flood',
    'before',
    'Prepare before flooding',
    'Know your nearest shelter, avoid dumping refuse in drains, keep documents dry, and prepare an emergency bag.',
    'en',
    true,
    10
  ),
  (
    '00000000-0000-0000-0000-000000000502',
    'flood',
    'during',
    'Stay safe during flooding',
    'Move to higher ground, avoid walking or driving through floodwater, turn off electricity if safe, and call 112 for emergencies.',
    'en',
    true,
    20
  ),
  (
    '00000000-0000-0000-0000-000000000503',
    'flood',
    'after',
    'Return safely after flooding',
    'Wait for official guidance, avoid contaminated water, photograph damage before cleanup, and report blocked drains or damaged utilities.',
    'en',
    true,
    30
  ),
  (
    '00000000-0000-0000-0000-000000000504',
    'fire',
    'during',
    'Fire safety response',
    'Leave immediately, warn people nearby, stay low under smoke, never use lifts, and call 112 for Ghana National Fire Service support.',
    'en',
    true,
    40
  ),
  (
    '00000000-0000-0000-0000-000000000505',
    'road_crash',
    'during',
    'Road crash first response',
    'Move to a safe place, switch on hazard lights if possible, do not move injured people unless there is immediate danger, and call 112.',
    'en',
    true,
    50
  ),
  (
    '00000000-0000-0000-0000-000000000506',
    'electrical_hazard',
    'during',
    'Electrical hazard safety',
    'Stay away from fallen wires, flooded electrical equipment, and sparking poles. Keep others clear and call 112 or the utility emergency line.',
    'en',
    true,
    60
  ),
  (
    '00000000-0000-0000-0000-000000000507',
    'disease_outbreak',
    'before',
    'Disease prevention basics',
    'Wash hands often, isolate when symptomatic, follow Ghana Health Service guidance, keep medicine supplies ready, and protect vulnerable family members.',
    'en',
    true,
    70
  ),
  (
    '00000000-0000-0000-0000-000000000508',
    'other',
    'during',
    'Safe evacuation',
    'Take only essentials, follow official routes, help children and elderly people first, avoid floodwater or smoke, and tell relatives where you are going.',
    'en',
    true,
    80
  ),
  (
    '00000000-0000-0000-0000-000000000509',
    'other',
    'before',
    'Emergency bag checklist',
    'Pack water, food, torch, radio, power bank, first aid, medicine, copies of documents, cash, hygiene items, and child or disability-specific supplies.',
    'en',
    true,
    90
  ),
  (
    '00000000-0000-0000-0000-000000000510',
    'other',
    'before',
    'Family emergency plan',
    'Choose meeting points, store emergency contacts, teach children how to call 112, plan transport, and decide who checks on vulnerable relatives.',
    'en',
    true,
    100
  ),
  (
    '00000000-0000-0000-0000-000000000511',
    'other',
    'during',
    'Calling 112',
    'Call 112 for life-threatening emergencies. Share the hazard, exact location, people affected, injuries, and a safe callback number if available.',
    'en',
    true,
    110
  )
ON CONFLICT (id) DO UPDATE SET
  hazard_type = EXCLUDED.hazard_type,
  stage = EXCLUDED.stage,
  title = EXCLUDED.title,
  body = EXCLUDED.body,
  language = EXCLUDED.language,
  offline_available = EXCLUDED.offline_available,
  sort_order = EXCLUDED.sort_order,
  updated_at = now();

INSERT INTO incidents (
  id,
  reference,
  type,
  severity,
  status,
  description,
  location_geometry,
  reported_by,
  anonymous,
  contact_permission,
  people_affected,
  injuries_reported
)
VALUES (
  '00000000-0000-0000-0000-000000000601',
  'INC-0001',
  'flood',
  'high',
  'reported',
  'Sample report: water is rising near a low-lying road.',
  ST_GeomFromText('POINT(-0.1870 5.6037)', 4326),
  '00000000-0000-0000-0000-000000000203',
  false,
  true,
  4,
  false
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO alerts (
  id,
  title,
  message,
  hazard_type,
  severity,
  target_geometry,
  target_label,
  issued_by,
  approved_by,
  agency_id,
  starts_at,
  expires_at,
  status,
  recommended_action,
  evacuation_required,
  shelter_links
)
VALUES (
  '00000000-0000-0000-0000-000000000701',
  'Development Severe Flood Watch',
  'Heavy rainfall may cause flooding in low-lying communities. Avoid flooded roads and prepare to move to higher ground.',
  'flood',
  'severe_warning',
  ST_Multi(ST_GeomFromText('POLYGON((-0.250 5.520, -0.110 5.520, -0.110 5.650, -0.250 5.650, -0.250 5.520))', 4326)),
  'Accra Metropolitan sample zone',
  '00000000-0000-0000-0000-000000000202',
  '00000000-0000-0000-0000-000000000201',
  '00000000-0000-0000-0000-000000000101',
  now(),
  now() + interval '12 hours',
  'approved',
  'Prepare to evacuate if instructed by authorities.',
  false,
  '["00000000-0000-0000-0000-000000000301"]'::jsonb
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO ml_predictions (
  id,
  hazard_type,
  model_version,
  prediction_time,
  target_time,
  geometry,
  probability,
  severity,
  confidence,
  explanation,
  input_feature_set_version
)
VALUES (
  '00000000-0000-0000-0000-000000000801',
  'flood',
  'flood-rule-baseline-0.1.0',
  now(),
  now() + interval '3 hours',
  ST_Multi(ST_GeomFromText('POLYGON((-0.230 5.530, -0.160 5.530, -0.160 5.590, -0.230 5.590, -0.230 5.530))', 4326)),
  0.8200,
  'severe',
  'medium',
  '["heavy rainfall forecast", "low elevation", "historical flood zone"]'::jsonb,
  'development-fixture-0.1.0'
)
ON CONFLICT (id) DO NOTHING;
