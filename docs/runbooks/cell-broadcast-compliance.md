# Cell Broadcast Compliance & Operational Runbook (NADAA-163)

**Owner:** NADMO National Alerting Desk · **Service:** `notification-service` ·
**Status:** Sandbox-ready; live telecom path pending official agreement.

Cell Broadcast (CB) pushes an approved emergency alert to every handset within a
geographic scope on a reserved message-identifier channel, independent of any
subscriber list or data connection. This runbook covers the approval controls,
compliance requirements, operational procedure, and failure handling for the
NADAA cell broadcast capability.

> **Be Aware. Be Prepared. Be Safe.** Cell broadcast is the highest-reach, most
> intrusive alerting channel NADAA operates. Every broadcast is human-approved
> before it can leave the platform. There is no automated send path.

---

## 1. Scope and standards

- **Standards:** 3GPP TS 23.041 (Cell Broadcast Service), aligned to the CMAS /
  Wireless Emergency Alerts (WEA) channel model and CAP (Common Alerting Protocol)
  classification. Ghana deployment maps onto the operator's Cell Broadcast Centre
  (CBC) per the national emergency-alerting agreement.
- **Channels (message identifiers):**
  | Identifier | Label        | Handset category   | Source severity                   | Override               |
  | ---------: | ------------ | ------------------ | --------------------------------- | ---------------------- |
  |       4370 | presidential | Presidential Alert | extreme / critical / catastrophic | yes — bypasses opt-out |
  |       4371 | extreme      | Extreme Alert      | severe / high                     | no                     |
  |       4373 | severe       | Severe Alert       | everything else eligible          | no                     |
- **Reserved use:** CB is reserved for severe-and-above hazards. Lower-severity
  advisories use push/SMS/voice, not cell broadcast.
- **Languages:** en, tw (Twi), ga (Ga), ee (Ewe), dag (Dagbani), ha (Hausa). One
  page-bounded segment is generated per language.
- **Encoding limits:** each CB page is 82 octets → 93 GSM-7 characters or 40 UCS-2
  characters, up to 15 concatenated pages. The service computes the data coding
  scheme, character count, page count, and truncation flag per segment.

---

## 2. Approval and override controls (must remain enforced)

These controls are enforced in code and must not be bypassed:

1. **Source is an already-approved alert.** A cell broadcast can only be generated
   from an alert present in the citizen feed (human-approved upstream). Expired
   alerts are rejected (`alert_not_deliverable`).
2. **Two-step human review.** Generation produces a `pending_review` set. A named
   reviewer must `approve` it before delivery. Delivery of a set that is not
   `approved`/`approved` returns `409 cell_broadcast_not_approved`.
3. **Per-language review.** Reviewers may approve or reject individual language
   segments; only `approved` segments are broadcast, others are recorded as
   `skipped`. A mixed outcome is surfaced as `partial_review`.
4. **Emergency override is explicit.** The presidential channel (4370) sets
   `emergencyOverride = true`, which bypasses subscriber opt-out. It is reserved
   for `extreme`-severity alerts and is visible on the message and every dispatch
   record for audit.
5. **No generic path.** `cell_broadcast` is deliberately excluded from the generic
   `/alerts/{id}/deliver` channel allowlist so it can never be sent outside the
   approval-gated flow. It appears in the delivery-log allowlist for audit only.

---

## 3. Adapter isolation and configuration

The telecom integration is isolated behind the `CellBroadcastAdapter` interface so
the official path, a sandbox simulator, or a disabled no-op can be swapped without
touching approval or audit logic. `NADAA_CELL_BROADCAST_MODE` selects the adapter:

| Mode                 | Adapter        | Behavior                                                                              |
| -------------------- | -------------- | ------------------------------------------------------------------------------------- |
| `disabled` (default) | `disabled_cbc` | Records every dispatch as `skipped`. Safe default until the live agreement is active. |
| `sandbox`            | `sandbox_cbc`  | In-process simulator. Live sends record `broadcast`; dry runs record `simulated`.     |

A real telecom integration registers its adapter in
`handlers.CellBroadcastAdapterFromMode` and is added to this table. **Do not enable
a live adapter without a signed operator agreement, an assigned CBC endpoint, and a
completed end-to-end test in the operator's sandbox.**

---

## 4. Operational procedure (activation)

Pre-conditions: an approved severe+ alert exists; on-call reviewer and NADMO duty
officer are available; `NADAA_CELL_BROADCAST_MODE` is set to the intended adapter.

1. **Generate.** `POST /api/v1/notifications/cell-broadcasts`
   with `{ "alertId": "...", "languages": [...], "areas": [...] }`.
   Confirm the returned channel, `emergencyOverride`, and per-segment page counts
   are as expected.
2. **Preview.** `GET /api/v1/notifications/cell-broadcasts/{id}/preview`. Read every
   language segment exactly as it will render on a handset. Confirm no segment is
   `truncated` unless intended, and that the emergency number (112) and target area
   are present.
3. **Review.** `POST /api/v1/notifications/cell-broadcasts/{id}/review`
   with `{ "action": "approve", "reviewer": "<name>", "note": "..." }`. Reject and
   regenerate if any segment is wrong. This step is mandatory and attributed.
4. **Dry run (recommended for live mode).** Deliver with `{ "dryRun": true }`.
   Every dispatch records `simulated`; no live broadcast is emitted. Confirm the
   channel, serial numbers, and areas.
5. **Broadcast.** `POST /api/v1/notifications/cell-broadcasts/{id}/deliver`
   with `{}` (or an `areas` override). Each language segment produces one dispatch
   with a unique CB serial number.
6. **Verify.** `GET /api/v1/notifications/delivery-logs?channel=cell_broadcast&alertId=<id>`.
   Confirm one audit entry per dispatched segment with the expected adapter and status.

---

## 5. Audit and observability

- **Audit log:** every dispatch (including `skipped`) writes a `cell_broadcast`
  entry to the unified delivery log, keyed by alert id, carrying the adapter,
  status, language, serial number, and target areas.
- **Structured logs:** generation, review, dispatch, and skip events emit
  structured logs (`cell broadcast message generated/reviewed`, `cell broadcast
segment dispatched/skipped`). Ship these to the observability stack (see
  NADAA-170) and alert on unexpected `failed`/`skipped` rates.
- **Metrics to watch:** dispatch count by channel and status, review latency
  (generate → approve), truncation rate, and adapter error rate.

---

## 6. Failure handling and rollback

- **Adapter unavailable / errors:** dispatches record `skipped` or `failed` with a
  reason. The alert is not silently dropped — re-attempt after confirming the CBC
  endpoint, or fall back to voice/SMS/push while the CB path is restored.
- **Wrong content broadcast:** CB has no per-handset recall. Mitigate by
  broadcasting a corrected/cancellation message on the same channel and geographic
  scope, and by updating the citizen feed. Record the incident.
- **Over-broad geographic scope:** re-broadcast with a corrected `areas` list; note
  that CB reaches transient devices in-scope (roaming, visitors) by design.
- **Disable quickly:** set `NADAA_CELL_BROADCAST_MODE=disabled` and redeploy to make
  the path a no-op without removing the feature.

---

## 7. Pre-go-live checklist (live telecom path)

- [ ] Signed operator / regulator agreement for CB channel use in Ghana.
- [ ] Assigned CBC endpoint, credentials, and geographic-scope (area) code mapping.
- [ ] Production `CellBroadcastAdapter` implemented and registered.
- [ ] End-to-end test completed in the operator's sandbox (all channels, all
      languages, dry run + live, override + non-override).
- [ ] Reviewer roster, escalation path, and duty-officer sign-off defined.
- [ ] Observability dashboards and alerts wired (NADAA-170).
- [ ] This runbook reviewed by NADMO and the operator's emergency-alerting team.

---

## 8. Related

- Service: [services/notification-service/README.md](../../services/notification-service/README.md)
- Integrations overview: [docs/integrations.md](../integrations.md)
- Resilience & ops: NADAA-170 (`docs/runbooks/` and observability dashboards)
- Voice alert flow (the analog approval-gated channel): notification-service voice endpoints
