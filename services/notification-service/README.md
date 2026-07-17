# Notification Service

The notification service owns the citizen alert feed, push/SMS provider abstraction, reviewed voice alert assets, telecom cell broadcast, inclusive access webhooks, and delivery attempt logs.

## Endpoints

- `GET /healthz`
- `GET /api/v1/notifications/alerts`
- `POST /api/v1/notifications/alerts/{id}/deliver`
- `GET /api/v1/notifications/delivery-logs`
- `POST /api/v1/notifications/voice-alerts`
- `GET /api/v1/notifications/voice-alerts`
- `POST /api/v1/notifications/voice-alerts/{id}/review`
- `POST /api/v1/notifications/voice-alerts/{id}/deliver`
- `POST /api/v1/notifications/cell-broadcasts`
- `GET /api/v1/notifications/cell-broadcasts`
- `GET /api/v1/notifications/cell-broadcasts/{id}/preview`
- `POST /api/v1/notifications/cell-broadcasts/{id}/review`
- `POST /api/v1/notifications/cell-broadcasts/{id}/deliver`

`GET /api/v1/notifications/alerts?includeExpired=true` returns current and expired citizen alert feed items. The service attempts to read approved/published alerts from `NADAA_ALERT_SERVICE_URL` and keeps fixture fallback alerts available for local development.

`POST /api/v1/notifications/alerts/{id}/deliver` accepts:

```json
{
  "recipientId": "usr_demo_citizen",
  "phone": "+233200000000",
  "pushToken": "ExponentPushToken-demo",
  "channels": ["push", "sms"],
  "language": "en"
}
```

### Delivery providers (dependency injection)

Each channel resolves a `models.NotificationProvider` through the
`handlers.BuildProviders` seam, so a real backend can be swapped in per channel
purely through configuration. Every channel **defaults to the sandbox (mock)
provider** so the platform runs end-to-end before real credentials arrive, and
selection **fails safe**: choosing a real provider without its credentials
disables that channel with a clear reason rather than silently mocking a live
selection or crashing (mirroring the cell-broadcast default).

| Channel | `NADAA_*_PROVIDER` | Options | Default |
| --- | --- | --- | --- |
| SMS | `NADAA_SMS_PROVIDER` | `sandbox`, `arkesel`, `disabled` | `sandbox` |
| Push | `NADAA_PUSH_PROVIDER` | `sandbox`, `expo`, `disabled` | `sandbox` |
| Voice | `NADAA_VOICE_PROVIDER` | `sandbox`, `disabled` (`arkesel` reserved) | `sandbox` |

The legacy `NADAA_SMS_ENABLED` / `NADAA_PUSH_ENABLED` / `NADAA_VOICE_ENABLED`
flags still work — setting one to `false` forces that channel to `disabled`.

**Arkesel (SMS)** — the confirmed value-for-money SMS backend for Ghana (direct
MTN / Telecel / AirtelTigo routes). Select with `NADAA_SMS_PROVIDER=arkesel` and set:

- `NADAA_ARKESEL_API_KEY` — the account API key (required; missing key disables the channel).
- `NADAA_ARKESEL_SENDER` — the approved sender id (default `NADAA`).
- `NADAA_ARKESEL_BASE_URL` — override the API host (default `https://sms.arkesel.com`).

**Expo (push)** — the value-for-money push backend for the Expo mobile apps
(free, delivers to APNs + FCM behind one token). Select with
`NADAA_PUSH_PROVIDER=expo`. Optionally set `NADAA_EXPO_ACCESS_TOKEN` (enhanced
push security) and `NADAA_EXPO_BASE_URL` (override, default `https://exp.host`).

**Voice** — the provider research recommends **Arkesel Voice** as the
value-for-money choice for Ghana (GHS 0.15/min, per-second, answered-only, with
Twi/Hausa/Ewe text-to-speech — roughly 40× cheaper than Twilio), with Africa's
Talking Voice as the fallback if DTMF-over-API must be proven on day one. The
live Arkesel Voice path is not wired yet (its campaign/DTMF API needs a paid
pilot), so `voice` stays on `sandbox`; selecting `arkesel` disables the channel
with a clear reason until it is integrated.

Real providers are fail-safe: a `dryRun` request never touches the network
(recorded `simulated`), a missing phone/push token is `skipped`, and an upstream
error is `failed` with the provider's reason — never a silent success.

Delivery attempts are stored in the in-memory log for the MVP service and represented in the core database schema by `notification_delivery_logs`.

Voice delivery uses a separate approval gate:

1. Generate variants with `POST /api/v1/notifications/voice-alerts`.
2. Review them with `POST /api/v1/notifications/voice-alerts/{id}/review`.
3. Deliver only approved assets with `POST /api/v1/notifications/voice-alerts/{id}/deliver`.

The sandbox voice provider writes `mock_voice` delivery logs. Set `NADAA_VOICE_ENABLED=false` to log voice attempts as `skipped`.

### Cell broadcast (telecom)

Cell Broadcast (3GPP CBS / CMAS / WEA) pushes an approved emergency alert to every
handset in a geographic scope on a reserved message-identifier channel, independent
of any subscriber list. It follows the same human-approval gate as voice:

1. Generate the review-gated message set from an approved citizen alert with
   `POST /api/v1/notifications/cell-broadcasts`. The service picks the CMAS/WEA
   channel from the alert severity (4370 presidential / 4371 extreme / 4373 severe),
   renders one page-bounded segment per language, and attaches CAP classification.
2. Preview the handset-accurate rendering with
   `GET /api/v1/notifications/cell-broadcasts/{id}/preview`.
3. Approve or reject with `POST /api/v1/notifications/cell-broadcasts/{id}/review`.
4. Broadcast only approved sets with `POST /api/v1/notifications/cell-broadcasts/{id}/deliver`.

The telecom integration is isolated behind a `CellBroadcastAdapter`. `NADAA_CELL_BROADCAST_MODE`
selects it:

- `disabled` (default) — a safe no-op that records every dispatch as `skipped`; used
  until an official telecom agreement and Cell Broadcast Entity are in place.
- `sandbox` — an in-process simulator so the end-to-end flow (including dry runs, which
  record `simulated`) can be exercised without a live network.

Every dispatch is written to the unified delivery log as channel `cell_broadcast`
(`GET /api/v1/notifications/delivery-logs?channel=cell_broadcast`). See
[docs/runbooks/cell-broadcast-compliance.md](../../docs/runbooks/cell-broadcast-compliance.md)
for the compliance and operational runbook.

### Authority and webhook authentication

The delivery endpoints that spend real money or emit mass alerts require a
verified authority actor (bearer token issued by auth-service, MFA completed,
allowed role):

- `POST /api/v1/notifications/alerts/{id}/deliver`
- `POST /api/v1/notifications/voice-alerts/{id}/deliver`
- `POST /api/v1/notifications/cell-broadcasts/{id}/review` (reviewer is the
  verified actor; any `reviewer` string in the request body is ignored)
- `POST /api/v1/notifications/cell-broadcasts/{id}/deliver`

Configuration:

- `NADAA_AUTH_TOKEN_SECRET` — HMAC secret verifying `nadaa.<payload>.<sig>`
  bearer tokens. Empty secret → authority requests are rejected with 401.
- `NADAA_AUTH_ALLOW_MOCK_ACTORS=true` — local development and smoke tests only:
  honors legacy `X-NADAA-Actor-ID` / `X-NADAA-Agency-ID` / `X-NADAA-Actor-Role` /
  `X-NADAA-MFA-Completed` headers when no bearer token is present.

The inbound provider webhooks (`POST /api/v1/notifications/ussd`,
`/sms/inbound`, `/whatsapp/inbound`, `/whatsapp/webhook`) stay public by default
for local development; setting `NADAA_SMS_WEBHOOK_SECRET`,
`NADAA_USSD_WEBHOOK_SECRET`, or `NADAA_WHATSAPP_WEBHOOK_SECRET` makes the
matching webhook require the same value in the `X-NADAA-Webhook-Secret` header
(constant-time comparison, 401 otherwise). Channels without a configured secret
log a one-time WARN at startup. (`NADAA_VOICE_WEBHOOK_SECRET` is read for
symmetry; no voice webhook route exists yet.)

## Local Development

```bash
cd services/notification-service
go run .
```

The service listens on `:8090` by default. Override with `NADAA_NOTIFICATION_ADDR`.

Run tests:

```bash
go test ./...
```

Run smoke checks with the service on port `8090`:

```bash
pnpm smoke:notification
pnpm smoke:voice-alerts
```

## Story Coverage

- NADAA-052
- NADAA-110
- NADAA-111
- NADAA-112
- NADAA-163
