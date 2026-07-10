#!/usr/bin/env bash
# Build + run the remaining dashboard-facing services on the free 94xx block
# (extends dev-citizen-backends.sh which runs 9410-9421). CORS allow-all when
# NADAA_ALLOWED_ORIGINS is empty.
ROOT="/Users/shayford/Desktop/Dev/Projects/nadaa"
BIN="/tmp/nadaa-svc-bin"
LOG="/tmp/nadaa-svc-logs"
mkdir -p "$BIN" "$LOG"

start() {
  local dir="$1" name="$2" penv="$3" port="$4" extra="$5"
  local pid
  pid=$(lsof -nP -iTCP:"$port" -sTCP:LISTEN -t 2>/dev/null)
  [ -n "$pid" ] && kill "$pid" 2>/dev/null
  echo "[build] $name"
  if ! (cd "$ROOT/services/$dir" && go build -o "$BIN/$name" ./cmd/server) 2>"$LOG/$name.build.log"; then
    echo "[BUILD-FAIL] $name — see $LOG/$name.build.log"
    return
  fi
  echo "[run]   $name :$port"
  # $extra carries optional per-service KEY=VALUE env (word-split intentionally).
  env "$penv=:$port" NADAA_ALLOWED_ORIGINS="" $extra nohup "$BIN/$name" >"$LOG/$name.log" 2>&1 </dev/null &
  disown
}

start alert-service        alert        NADAA_ALERT_ADDR        9422
start ml-service           ml           NADAA_ML_ADDR           9423
start school-service       school       PORT                    9424
start imagery-service      imagery      PORT                    9425
# auth-service alone does real token auth; the demo dashboards send the shared
# X-NADAA-* actor headers, so opt this dev instance into mock-actor parity.
start auth-service         auth         NADAA_AUTH_ADDR         9426  NADAA_AUTH_ALLOW_MOCK_ACTORS=true
start integration-service  integration  NADAA_INTEGRATION_ADDR  9427

sleep 6
echo "=== listening check ==="
for p in 9422 9423 9424 9425 9426 9427; do
  if lsof -nP -iTCP:$p -sTCP:LISTEN >/dev/null 2>&1; then echo "$p UP"; else echo "$p DOWN"; fi
done
