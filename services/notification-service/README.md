# Notification Service

The notification service owns the citizen alert feed, push/SMS provider abstraction, reviewed voice alert assets, inclusive access webhooks, and delivery attempt logs.

## Endpoints

- `GET /healthz`
- `GET /api/v1/notifications/alerts`
- `POST /api/v1/notifications/alerts/{id}/deliver`
- `GET /api/v1/notifications/delivery-logs`
- `POST /api/v1/notifications/voice-alerts`
- `GET /api/v1/notifications/voice-alerts`
- `POST /api/v1/notifications/voice-alerts/{id}/review`
- `POST /api/v1/notifications/voice-alerts/{id}/deliver`

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

Development providers are mock providers by default:

- `push` uses `mock_push`.
- `sms` uses `mock_sms`.
- Set `NADAA_SMS_ENABLED=false` to log SMS attempts as `skipped`.
- Set `NADAA_PUSH_ENABLED=false` to log push attempts as `skipped`.

Delivery attempts are stored in the in-memory log for the MVP service and represented in the core database schema by `notification_delivery_logs`.

Voice delivery uses a separate approval gate:

1. Generate variants with `POST /api/v1/notifications/voice-alerts`.
2. Review them with `POST /api/v1/notifications/voice-alerts/{id}/review`.
3. Deliver only approved assets with `POST /api/v1/notifications/voice-alerts/{id}/deliver`.

The sandbox voice provider writes `mock_voice` delivery logs. Set `NADAA_VOICE_ENABLED=false` to log voice attempts as `skipped`.

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
