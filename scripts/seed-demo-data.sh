#!/usr/bin/env bash
# Seed the incident-service (which starts empty) with realistic Ghana demo
# incidents via its public POST intake, so the dashboards show real data.
# Other services (missing-persons, aid, shelters, campaigns, alerts) self-seed.
export PATH="/usr/bin:/bin:/usr/sbin:/sbin:/opt/homebrew/bin"
INCIDENTS="http://localhost:9413/api/v1/incidents"

post() {
  curl -s -o /dev/null -w "  incident: %{http_code}\n" -X POST "$INCIDENTS" \
    -H "Content-Type: application/json" -d "$1"
}

echo "Seeding incidents -> $INCIDENTS"
post '{"type":"road_crash","description":"Three-vehicle crash blocking the Tema motorway shoulder; responders needed.","location":{"lat":5.6698,"lng":-0.0166},"peopleAffected":6,"injuriesReported":true,"urgency":"high","anonymous":false,"contactPermission":true,"media":[],"reporter":{"userId":"usr_demo_citizen"}}'
post '{"type":"fire","description":"Market fire spreading through stalls at Kaneshie; people trapped inside.","location":{"lat":5.5717,"lng":-0.2360},"peopleAffected":30,"injuriesReported":true,"urgency":"life_threatening","anonymous":false,"contactPermission":true,"media":[],"reporter":{"userId":"usr_demo_citizen"}}'
post '{"type":"flood","description":"Odaw river overflowing into homes around Alajo; families evacuating.","location":{"lat":5.5850,"lng":-0.2100},"peopleAffected":120,"injuriesReported":false,"urgency":"high","anonymous":false,"contactPermission":true,"media":[],"reporter":{"userId":"usr_demo_citizen"}}'
post '{"type":"road_crash","description":"Motorbike and taxi collision on the Spintex road; one rider down.","location":{"lat":5.6280,"lng":-0.1160},"peopleAffected":2,"injuriesReported":true,"urgency":"moderate","anonymous":true,"contactPermission":false,"media":[],"reporter":{"userId":"usr_demo_citizen"}}'
post '{"type":"flood","description":"Flash flood in a low-lying area of Kumasi after heavy rain.","location":{"lat":6.6885,"lng":-1.6244},"peopleAffected":45,"injuriesReported":false,"urgency":"moderate","anonymous":false,"contactPermission":true,"media":[],"reporter":{"userId":"usr_demo_citizen"}}'
post '{"type":"other","description":"Partial building collapse near Makola; unknown number of people trapped.","location":{"lat":5.5470,"lng":-0.2100},"peopleAffected":8,"injuriesReported":true,"urgency":"life_threatening","anonymous":false,"contactPermission":true,"media":[],"reporter":{"userId":"usr_demo_citizen"}}'
post '{"type":"flood","description":"Coastal flooding in Ada affecting shoreline homes and a school.","location":{"lat":5.7850,"lng":0.6330},"peopleAffected":60,"injuriesReported":false,"urgency":"high","anonymous":false,"contactPermission":true,"media":[],"reporter":{"userId":"usr_demo_citizen"}}'
post '{"type":"electrical_hazard","description":"Downed live power line across a residential street in Madina.","location":{"lat":5.6820,"lng":-0.1660},"peopleAffected":0,"injuriesReported":false,"urgency":"moderate","anonymous":true,"contactPermission":false,"media":[],"reporter":{"userId":"usr_demo_citizen"}}'
post '{"type":"flood","description":"Drainage overflow flooding the Circle underpass; vehicles stalling.","location":{"lat":5.5710,"lng":-0.1970},"peopleAffected":15,"injuriesReported":false,"urgency":"high","anonymous":false,"contactPermission":true,"media":[],"reporter":{"userId":"usr_demo_citizen"}}'

echo "Done. Current count:"
curl -s -H "X-NADAA-Actor-ID:usr_sys" -H "X-NADAA-Agency-ID:agc_1" -H "X-NADAA-Actor-Role:system_admin" -H "X-NADAA-MFA-Completed:true" "$INCIDENTS" \
  | grep -o '"id"' | grep -c '"id"' | sed 's/^/  incidents in store: /'
