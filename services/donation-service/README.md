# Donation Service

The donation service coordinates donors, aid catalog items, aid requests, and pledges for the NADAA platform.

## Current endpoints

Public endpoints:

- `GET /healthz`
- `GET /api/v1/aid-catalog`
- `GET /api/v1/aid-requests?status=&category=&region=&priority=`
- `POST /api/v1/donors`
- `POST /api/v1/aid-requests/{id}/pledges`
- `POST /api/v1/donations`
- `GET /api/v1/donations/{reference}`
- `POST /api/v1/webhooks/paystack`

Authority endpoints (require `X-NADAA-Actor-ID`, `X-NADAA-Actor-Role`, `X-NADAA-Agency-ID`, `X-NADAA-MFA-Completed: true`, `X-NADAA-Request-ID`):

- `GET /api/v1/donors?type=&q=`
- `POST /api/v1/donors`
- `GET /api/v1/donors/{id}`
- `PATCH /api/v1/donors/{id}`
- `POST /api/v1/aid-requests`
- `GET /api/v1/aid-requests/{id}`
- `PATCH /api/v1/aid-requests/{id}`
- `GET /api/v1/aid-requests/{id}/pledges`
- `GET /api/v1/pledges?status=`
- `PATCH /api/v1/pledges/{id}`
- `POST /api/v1/aid-requests/{id}/allocate`
- `GET /api/v1/donations?status=&campaign=`

Allowed authority roles are `system_admin`, `agency_admin`, `agency_viewer`, `nadmo_officer`, `district_officer`, `dispatcher`, and `ngo`.

## Monetary donations (payment gateway)

Cash donations (mobile money + cards) run through a `models.PaymentProvider`
resolved by the `handlers.BuildPaymentProvider` seam, so the gateway can be
swapped purely through configuration. It **defaults to the sandbox provider** so
the flow runs end-to-end before real credentials arrive, and it **fails safe**:
selecting a real provider without its key yields a disabled provider with a
clear reason rather than a broken live path.

Flow:

1. `POST /api/v1/donations` records a `pending` donation and returns the
   gateway `authorizationUrl` the donor is redirected to. Body:
   `{"donorName","email","amount","currency":"GHS","campaign","message"}`
   (`email` and an `amount` ≥ GHS 1.00 are required; amounts are stored in
   pesewas).
2. The donor pays on the gateway page.
3. The donation is credited **only after a server-side verification** — never
   from a webhook payload alone. `POST /api/v1/webhooks/paystack` verifies the
   `x-paystack-signature` (HMAC-SHA512), then re-verifies the transaction via
   the gateway before marking it paid. `GET /api/v1/donations/{reference}` does
   the same verification on demand. All transitions are idempotent, and a
   verified-amount mismatch is rejected as `amount_mismatch`, so replayed or
   tampered webhooks cannot double-credit.

**Paystack** is the confirmed value-for-money gateway for Ghana (direct MTN
MoMo, Telecel Cash, AirtelTigo Money + cards, T+1 direct-to-MoMo settlement,
HMAC-SHA512 webhooks). Select with `NADAA_PAYMENT_PROVIDER=paystack` and set
`NADAA_PAYSTACK_SECRET_KEY` (required), `NADAA_PAYSTACK_CALLBACK_URL`
(post-payment redirect), and optionally `NADAA_PAYSTACK_BASE_URL`
(default `https://api.paystack.co`).

## Run

```bash
go run ./cmd/server
```

The service listens on `:8100` by default. Override with `PORT`.

## Test

```bash
go test ./...
```

## Build

```bash
go build ./cmd/server
```

## Environment variables

| Variable | Default | Description |
| --- | --- | --- |
| `PORT` | `:8100` | HTTP listen address. |
| `NADAA_ALLOWED_ORIGINS` | `*` | Comma-separated allowed CORS origins. Use `*` to allow all. |
| `NADAA_PAYMENT_PROVIDER` | `sandbox` | Payment gateway: `sandbox`, `paystack`, or `disabled`. |
| `NADAA_PAYSTACK_SECRET_KEY` | — | Paystack secret key (required when provider is `paystack`). |
| `NADAA_PAYSTACK_CALLBACK_URL` | — | Post-payment redirect URL. |
| `NADAA_PAYSTACK_BASE_URL` | `https://api.paystack.co` | Override the Paystack API host. |

## Notes

The current implementation uses an in-memory fixture store so citizen and authority surfaces can integrate against a stable API contract. Persistence can replace the store later without changing the endpoint shapes.
