# Donation Service

The donation service coordinates donors, aid catalog items, aid requests, and pledges for the NADAA platform.

## Current endpoints

Public endpoints:

- `GET /healthz`
- `GET /api/v1/aid-catalog`
- `GET /api/v1/aid-requests?status=&category=&region=&priority=`
- `POST /api/v1/donors`
- `POST /api/v1/aid-requests/{id}/pledges` (for public callers the pledge must
  be bound to a registered donor: `donorId` must reference an existing donor
  and `contactEmail` must match that donor's registered email
  case-insensitively, otherwise the request is rejected with 403. A verified
  authority caller pledges on behalf of the donor — the email match is skipped
  and the donor's registered email is inherited when the form sends none.
  Only delivered pledge quantities count toward a request's
  `quantityFulfilled`; pledged-but-undelivered quantities never mark a request
  `fulfilled`)
- `POST /api/v1/donations` (rate limited per client — see
  `NADAA_DONATION_RATE_LIMIT` below — because every initialization fires an
  outbound gateway call)
- `POST /api/v1/webhooks/paystack`

Authority endpoints (require an `Authorization: Bearer nadaa.<payload>.<sig>`
token issued by auth-service with an authority role and `mfa: true`; legacy
`X-NADAA-Actor-*` headers are honored only when
`NADAA_AUTH_ALLOW_MOCK_ACTORS=true` for local dev and smoke tests):

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
- `POST /api/v1/aid-requests/{id}/allocate` (records one delivered tranche:
  delivered quantities accumulate, a tranche beyond the remaining pledged
  quantity is rejected with 400, and the pledge flips to `delivered` only once
  fully allocated)
- `GET /api/v1/donations?status=&campaign=`
- `GET /api/v1/donations/{reference}`

Allowed authority roles are `system_admin`, `agency_admin`, `agency_viewer`, `nadmo_officer`, `district_officer`, `dispatcher`, and `ngo`.

## Monetary donations (payment gateway)

Cash donations (mobile money + cards) run through a `models.PaymentProvider`
resolved by the `handlers.BuildPaymentProvider` seam, so the gateway can be
swapped purely through configuration. It **defaults to the sandbox provider** so
the flow runs end-to-end before real credentials arrive, and it **fails safe**:
selecting a real provider without its key yields a disabled provider with a
clear reason rather than a broken live path. The sandbox provider only credits
donations (simulated "paid" verification and webhook acceptance) when
`NADAA_ENV=development`; anywhere else it leaves donations pending, rejects
webhooks, and its activation is WARN-logged at startup.

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
   the same verification on demand. All transitions are idempotent, a
   verified-amount mismatch is rejected as `amount_mismatch`, and a
   verified-currency mismatch as `currency_mismatch`, so replayed or tampered
   webhooks cannot double-credit. Transient gateway errors (timeouts, 5xx)
   leave the donation `pending` for a later re-check rather than failing it.

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
| `NADAA_ENV` | — | Set to `development` to enable sandbox payment crediting and localhost CORS bypass. |
| `NADAA_AUTH_TOKEN_SECRET` | — | HMAC-SHA256 secret verifying auth-service bearer tokens; empty rejects authority requests unless mock actors are allowed. |
| `NADAA_AUTH_ALLOW_MOCK_ACTORS` | `false` | When `true`, honor legacy `X-NADAA-Actor-*` headers (local dev and smoke tests). Rejected at startup unless `NADAA_ENV=development`. |
| `NADAA_DONATION_RATE_LIMIT` | `10` | Max `POST /api/v1/donations` payment initializations per client per window. |
| `NADAA_DONATION_RATE_WINDOW_SECONDS` | `60` | Rate-limit window length in seconds. |
| `NADAA_TRUST_PROXY_HEADERS` | `false` | When `true`, derive the rate-limit client IP from `X-Forwarded-For`/`X-Real-Ip` (only enable behind a trusted reverse proxy). |

## Notes

The current implementation uses an in-memory fixture store so citizen and authority surfaces can integrate against a stable API contract. Persistence can replace the store later without changing the endpoint shapes.
