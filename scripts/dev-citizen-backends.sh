#!/usr/bin/env bash
# Build and run the citizen-facing NADAA backends on a free 94xx port block.
# The default 80xx/90xx ports are squatted by unrelated apps on this machine,
# so the citizen app is pointed here via apps/citizen-web/.env.local.
# CORS: services allow-all when NADAA_ALLOWED_ORIGINS is empty.
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="/tmp/nadaa-svc-bin"
LOG="/tmp/nadaa-svc-logs"
mkdir -p "$BIN" "$LOG"

# Shared dev auth posture for every service: verify tokens signed with the dev
# secret, honor self-asserted X-NADAA-Actor-* headers from the demo apps, and
# share the internal service-to-service token. auth-service itself is run by
# dev-dashboard-backends.sh, which also prints the dev agency admin login.
export NADAA_ENV=development
export NADAA_AUTH_ALLOW_MOCK_ACTORS=true
export NADAA_AUTH_TOKEN_SECRET=dev-secret-change-me
export NADAA_INTERNAL_SERVICE_TOKEN=dev-internal-service-token

echo "=== NADAA citizen dev backends (root: $ROOT) ==="
echo "dev auth: NADAA_ENV=development, mock actor headers on, token secret dev-secret-change-me"

stop() {
  # Only kill processes running our own built binaries — never whatever else
  # happens to be listening on the port.
  local name="$1" pids
  pids=$(pgrep -f "^$BIN/$name$" 2>/dev/null || true)
  [ -n "$pids" ] && kill $pids 2>/dev/null
}

start() {
  # $1 service dir  $2 shortname  $3 port-env-var  $4 port
  local dir="$1" name="$2" penv="$3" port="$4"
  stop "$name"
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
