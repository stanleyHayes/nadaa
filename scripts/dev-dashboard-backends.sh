#!/usr/bin/env bash
# Build + run the remaining dashboard-facing services on the free 94xx block
# (extends dev-citizen-backends.sh which runs 9410-9421). CORS allow-all when
# NADAA_ALLOWED_ORIGINS is empty.
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="/tmp/nadaa-svc-bin"
LOG="/tmp/nadaa-svc-logs"
mkdir -p "$BIN" "$LOG"

# Shared dev auth posture for every service: verify tokens signed with the dev
# secret, honor self-asserted X-NADAA-Actor-* headers from the demo dashboards,
# and share the internal service-to-service token (ml-service requires it).
export NADAA_ENV=development
export NADAA_AUTH_ALLOW_MOCK_ACTORS=true
export NADAA_AUTH_TOKEN_SECRET=dev-secret-change-me
export NADAA_INTERNAL_SERVICE_TOKEN=dev-internal-service-token

# Dev agency admin bootstrapped into auth-service. Local dev only: the
# password clears the 12-character bootstrap minimum and the MFA code must be
# an explicit 6 digits — the service no longer falls back to a default code.
ADMIN_EMAIL="admin@nadaa.local"
ADMIN_PASSWORD="change-me-locally"
ADMIN_MFA_CODE="123456"

echo "=== NADAA dashboard dev backends (root: $ROOT) ==="
echo "dev agency admin login: $ADMIN_EMAIL / $ADMIN_PASSWORD (MFA code: $ADMIN_MFA_CODE)"
echo "dev auth: NADAA_ENV=development, mock actor headers on, token secret dev-secret-change-me"

stop() {
  # Only kill processes running our own built binaries — never whatever else
  # happens to be listening on the port.
  local name="$1" pids
  pids=$(pgrep -f "^$BIN/$name$" 2>/dev/null || true)
  [ -n "$pids" ] && kill $pids 2>/dev/null
}

start() {
  # $1 service dir  $2 shortname  $3 port-env-var  $4 port  $5+ optional KEY=VALUE env
  local dir="$1" name="$2" penv="$3" port="$4"
  shift 4
  stop "$name"
  echo "[build] $name"
  if ! (cd "$ROOT/services/$dir" && go build -o "$BIN/$name" ./cmd/server) 2>"$LOG/$name.build.log"; then
    echo "[BUILD-FAIL] $name — see $LOG/$name.build.log"
    return
  fi
  echo "[run]   $name :$port"
  env "$penv=:$port" NADAA_ALLOWED_ORIGINS="" "$@" nohup "$BIN/$name" >"$LOG/$name.log" 2>&1 </dev/null &
  disown
}

start alert-service        alert        NADAA_ALERT_ADDR        9422
# ml-service resolves its model artifacts relative to the working directory;
# pin them to the repo so the script works from any cwd.
start ml-service           ml           NADAA_ML_ADDR           9423  NADAA_ML_MODEL_DIR="$ROOT/data/flood-risk/models"
start school-service       school       PORT                    9424
start imagery-service      imagery      PORT                    9425
# auth-service alone needs the insecure-secret bypass: the shared dev signing
# secret is deliberately short, and its Validate() fails closed without it.
start auth-service         auth         NADAA_AUTH_ADDR         9426 \
  NADAA_AUTH_ALLOW_INSECURE_SECRET=true \
  NADAA_AUTH_BOOTSTRAP_ADMIN_EMAIL="$ADMIN_EMAIL" \
  NADAA_AUTH_BOOTSTRAP_ADMIN_PASSWORD="$ADMIN_PASSWORD" \
  NADAA_AUTH_BOOTSTRAP_ADMIN_MFA_CODE="$ADMIN_MFA_CODE"
start integration-service  integration  NADAA_INTEGRATION_ADDR  9427

sleep 6
echo "=== listening check ==="
for p in 9422 9423 9424 9425 9426 9427; do
  if lsof -nP -iTCP:$p -sTCP:LISTEN >/dev/null 2>&1; then echo "$p UP"; else echo "$p DOWN"; fi
done
