# NADAA Dispatcher Mobile

Expo/React Native triage app for Phase 2 dispatchers.

## Scope

- Dispatcher shell with NADAA brand, MFA-aware agency session, and bottom navigation.
- Incident queue with hazard/severity/status/time filters.
- Selected incident detail with timeline and duplicate candidates.
- Status update, assignment handoff, and timeline-note actions.
- Hospital capacity lookup near the selected incident.
- Shared contracts from `@nadaa/shared-types` and brand tokens from `@nadaa/brand`.
- Offline primitives for incident cache, session, and capacity persistence.
- Sandbox push registration for critical incident escalation.

## Local Checks

```bash
pnpm --filter @nadaa/dispatcher-mobile typecheck
pnpm smoke:dispatcher-mobile
```

## Expo Runtime

The app keeps Expo runtime packages as optional peers so the monorepo CI can validate contracts without installing the full native toolchain on every run. To run on a device/simulator, install the native peers for this workspace, then run:

```bash
pnpm --filter @nadaa/dispatcher-mobile start
```

Expected native peers: `expo`, `react`, `react-native`, `@expo/vector-icons`, `expo-location`, `expo-notifications`, and `@react-native-async-storage/async-storage`.

## Configuration

Copy `.env.example` when testing against local services:

- `EXPO_PUBLIC_AUTH_API_URL`
- `EXPO_PUBLIC_INCIDENT_API_URL`
- `EXPO_PUBLIC_SHELTER_API_URL`
- `EXPO_PUBLIC_PUSH_PROVIDER`
