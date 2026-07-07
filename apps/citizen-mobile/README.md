# NADAA Citizen Mobile

Expo/React Native foundation for the Phase 2 citizen mobile app.

## Scope

- Native citizen shell with NADAA brand, emergency 112 call action, and bottom navigation.
- Current alert feed, area risk, incident report draft/submission, offline guides, shelter/recovery support, and session/permission setup screens.
- Shared contracts from `@nadaa/shared-types` and brand tokens from `@nadaa/brand`.
- Offline primitives for guide cache, report draft, and session persistence.
- Push registration abstraction with sandbox provider state.

## Local Checks

```bash
pnpm --filter @nadaa/citizen-mobile typecheck
pnpm smoke:citizen-mobile
```

## Expo Runtime

The foundation keeps Expo runtime packages as optional peers so the monorepo CI can validate contracts without installing the full native toolchain on every run. To run on a device/simulator, install the native peers for this workspace, then run:

```bash
pnpm --filter @nadaa/citizen-mobile start
```

Expected native peers: `expo`, `react`, `react-native`, `@expo/vector-icons`, `expo-location`, `expo-image-picker`, `expo-notifications`, and `@react-native-async-storage/async-storage`.

## Configuration

Copy `.env.example` when testing against local services:

- `EXPO_PUBLIC_INCIDENT_API_URL`
- `EXPO_PUBLIC_RISK_API_URL`
- `EXPO_PUBLIC_NOTIFICATION_API_URL`
- `EXPO_PUBLIC_GUIDE_API_URL`
- `EXPO_PUBLIC_SHELTER_API_URL`
- `EXPO_PUBLIC_PUSH_PROVIDER`
