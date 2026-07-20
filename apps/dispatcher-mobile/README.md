# NADAA Dispatcher Mobile

Expo/React Native triage app for Phase 2 dispatchers.

## Scope

- Dispatcher shell with NADAA brand, MFA-aware agency session, and bottom navigation.
- Incident queue with hazard/severity/status/time filters.
- Selected incident detail with timeline and duplicate candidates.
- Status update, assignment handoff, and timeline-note actions.
- Hospital capacity lookup near the selected incident.
- Shared contracts from `@nadaa/shared-types` and brand tokens from `@nadaa/brand`.
- Offline primitives for incident cache, session, and capacity persistence
  (AsyncStorage); empty storage hydrates an honest signed-out, empty-queue
  state — never fixture data.
- Auth expiry (401/403) clears the stored session, stops queue polling, and
  routes back to sign-in.
- Foreground queue polling for critical incident escalation; push registration
  reports an honest not-configured state until notification-service exposes a
  device-token endpoint.

## Local Checks

```bash
pnpm --filter @nadaa/dispatcher-mobile typecheck
pnpm smoke:dispatcher-mobile
```

## Expo Runtime

The Expo runtime packages (`expo`, `react`, `react-native`, `@expo/vector-icons`, `expo-location`, `expo-notifications`, and `@react-native-async-storage/async-storage`) are pinned dependencies of this workspace. To run on a device/simulator:

```bash
pnpm --filter @nadaa/dispatcher-mobile start
```

## Configuration

Copy `.env.example` when testing against local services:

- `EXPO_PUBLIC_AUTH_API_URL`
- `EXPO_PUBLIC_INCIDENT_API_URL`
- `EXPO_PUBLIC_SHELTER_API_URL`
- `EXPO_PUBLIC_PUSH_PROVIDER`
