# NADAA Citizen Mobile

Expo/React Native foundation for the Phase 2 citizen mobile app.

## Scope

- Native citizen shell with NADAA brand, emergency 112 call action, and bottom navigation.
- Current alert feed, area risk, incident report draft/submission, offline guides, shelter/recovery support, and session/permission setup screens.
- Shared contracts from `@nadaa/shared-types` and brand tokens from `@nadaa/brand`.
- Offline primitives for guide cache, report draft, and session persistence (AsyncStorage).
- Citizen OTP sign-in against auth-service; the citizen Bearer token is attached to volunteer registration and task endpoints.
- Push registration obtains the real Expo push token and honestly reports when the server has no registration endpoint.

## Local Checks

```bash
pnpm --filter @nadaa/citizen-mobile typecheck
pnpm smoke:citizen-mobile
```

## Expo Runtime

Expo runtime packages (`expo`, `react`, `react-native`, `@expo/vector-icons`, `expo-location`, `expo-image-picker`, `expo-notifications`, `@react-native-async-storage/async-storage`) are pinned dependencies of this workspace. To run on a device/simulator:

```bash
pnpm --filter @nadaa/citizen-mobile start
```

## Configuration

Copy `.env.example` when testing against local services:

- `EXPO_PUBLIC_AUTH_API_URL`
- `EXPO_PUBLIC_INCIDENT_API_URL`
- `EXPO_PUBLIC_RISK_API_URL`
- `EXPO_PUBLIC_NOTIFICATION_API_URL`
- `EXPO_PUBLIC_GUIDE_API_URL`
- `EXPO_PUBLIC_SHELTER_API_URL`
- `EXPO_PUBLIC_PUSH_PROVIDER`
