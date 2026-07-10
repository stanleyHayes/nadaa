#!/usr/bin/env bash
# Build and run the citizen-facing NADAA backends on a free 94xx port block.
# The default 80xx/90xx ports are squatted by unrelated apps on this machine,
# so the citizen app is pointed here via apps/citizen-web/.env.local.
# CORS: services allow-all when NADAA_ALLOWED_ORIGINS is empty.
ROOT="/Users/shayford/Desktop/Dev/Projects/nadaa"
BIN="/tmp/nadaa-svc-bin"
LOG="/tmp/nadaa-svc-logs"
mkdir -p "$BIN" "$LOG"

start() {
  # $1 service dir  $2 shortname  $3 port-env-var  $4 port
  local dir="$1" name="$2" penv="$3" port="$4"
  local pid
  pid=$(lsof -nP -iTCP:"$port" -sTCP:LISTEN -t 2>/dev/null)
  [ -n "$pid" ] && kill "$pid" 2>/dev/null
  echo "[build] $name"
  if ! (cd "$ROOT/services/$dir" && go build -o "$BIN/$name" ./cmd/server) 2>"$LOG/$name.build.log"; then
    echo "[BUILD-FAIL] $name — see $LOG/$name.build.log"
    return
  fi
  echo "[run]   $name :$port"
  env "$penv=:$port" NADAA_ALLOWED_ORIGINS="" nohup "$BIN/$name" >"$LOG/$name.log" 2>&1 </dev/null &
  disown
}

start risk-service            risk          NADAA_RISK_ADDR          9410
start notification-service    notification  NADAA_NOTIFICATION_ADDR  9411
start guide-service           guide         NADAA_GUIDE_ADDR         9412
start incident-service        incident      NADAA_INCIDENT_ADDR      9413
start road-closure-service    roadclosure   NADAA_ROAD_CLOSURE_ADDR  9414
start shelter-service         shelter       NADAA_SHELTER_ADDR       9415
start route-service           route         PORT                     9416
start damage-claim-service    damage        PORT                     9417
start donation-service        donation      PORT                     9418
start missing-person-service  missingperson PORT                     9419
start open-data-service       opendata      PORT                     9420
start campaign-service        campaign      PORT                     9421

sleep 6
echo "=== listening check ==="
for p in 9410 9411 9412 9413 9414 9415 9416 9417 9418 9419 9420 9421; do
  if lsof -nP -iTCP:$p -sTCP:LISTEN >/dev/null 2>&1; then echo "$p UP"; else echo "$p DOWN"; fi
done
