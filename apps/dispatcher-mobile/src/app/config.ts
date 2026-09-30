/**
 * API origins, resolved at build time from EXPO_PUBLIC_* variables.
 *
 * The localhost fallbacks exist only for `expo start` against a local stack.
 * A release build must never inherit them silently: this app's primary job is
 * reporting an emergency, so a wrong origin means the report fails at the exact
 * moment it matters, with nothing shown to the user. `requireApiBase` therefore
 * throws on launch in a release build when a variable is missing or is not
 * HTTPS. A crash caught in QA or TestFlight is recoverable; a silent failure in
 * someone's hands during an emergency is not. iOS App Transport Security and
 * Android's cleartext policy (API 28+) would reject plain HTTP regardless.
 */
function requireApiBase(
  name: string,
  value: string | undefined,
  devFallback: string,
): string {
  const url = value?.trim();
  if (!url) {
    if (__DEV__) {
      return devFallback;
    }
    throw new Error(
      `${name} is not set. Release builds must define every EXPO_PUBLIC_*_API_URL; refusing to start against ${devFallback}.`,
    );
  }
  if (!__DEV__ && !url.startsWith("https://")) {
    throw new Error(
      `${name} must use https:// in a release build (received "${url}"). iOS App Transport Security and Android cleartext policy block plain HTTP.`,
    );
  }
  return url;
}

export const AUTH_API_BASE = requireApiBase(
  "EXPO_PUBLIC_AUTH_API_URL",
  process.env.EXPO_PUBLIC_AUTH_API_URL,
  "http://localhost:8080/api/v1",
);

export const INCIDENT_API_BASE = requireApiBase(
  "EXPO_PUBLIC_INCIDENT_API_URL",
  process.env.EXPO_PUBLIC_INCIDENT_API_URL,
  "http://localhost:8084/api/v1",
);

export const SHELTER_API_BASE = requireApiBase(
  "EXPO_PUBLIC_SHELTER_API_URL",
  process.env.EXPO_PUBLIC_SHELTER_API_URL,
  "http://localhost:8093/api/v1",
);

export const PUSH_PROVIDER = process.env.EXPO_PUBLIC_PUSH_PROVIDER ?? "sandbox";

export const SESSION_KEY = "nadaa.dispatcher.session.v1";
export const INCIDENT_CACHE_KEY = "nadaa.dispatcher.incidents.v1";
export const CAPACITY_CACHE_KEY = "nadaa.dispatcher.capacity.v1";
export const SELECTED_INCIDENT_KEY = "nadaa.dispatcher.selected-incident.v1";
