import type {
  AgencyMFASetupResponse,
  AgencyMFAVerifyResponse,
  LoginAgencyResponse,
} from "@nadaa/shared-types";
import { AUTH_API_BASE } from "@/app/config";

/**
 * Error raised for any non-OK auth-service response. `code` carries the
 * service's machine-readable reason (`invalid_credentials`, `mfa_required`,
 * `mfa_setup_required`, `too_many_attempts`, ...) so the sign-in screen can
 * branch on it; `message` is the service's human-readable detail.
 */
export class AuthApiError extends Error {
  readonly code: string;

  constructor(code: string, message: string) {
    super(message);
    this.name = "AuthApiError";
    this.code = code;
  }
}

/** Thrown when the auth service cannot be reached at all. */
export class AuthUnavailableError extends Error {
  constructor() {
    super("The authentication service is unavailable. Try again shortly.");
    this.name = "AuthUnavailableError";
  }
}

const JSON_HEADERS = { "Content-Type": "application/json" };

async function parseError(response: Response): Promise<AuthApiError> {
  let code = `http_${response.status}`;
  let message = `The authentication service rejected the request (${response.status}).`;
  try {
    const payload = (await response.json()) as {
      error?: { code?: string; message?: string };
    };
    if (payload.error?.code) {
      code = payload.error.code;
    }
    if (payload.error?.message) {
      message = payload.error.message;
    }
  } catch {
    // Non-JSON error body; keep the status-based defaults.
  }
  return new AuthApiError(code, message);
}

async function post<T>(path: string, body: Record<string, string>): Promise<T> {
  let response: Response;
  try {
    response = await fetch(`${AUTH_API_BASE}${path}`, {
      method: "POST",
      headers: JSON_HEADERS,
      body: JSON.stringify(body),
    });
  } catch {
    throw new AuthUnavailableError();
  }
  if (!response.ok) {
    throw await parseError(response);
  }
  return (await response.json()) as T;
}

/**
 * Authenticate an agency (authority) user. With no `mfaCode` the service
 * answers 401 `mfa_required` for enrolled accounts and 403
 * `mfa_setup_required` for accounts that must enrol first.
 */
export function loginAgency(
  email: string,
  password: string,
  mfaCode = "",
): Promise<LoginAgencyResponse> {
  return post<LoginAgencyResponse>("/auth/agency/login", {
    email,
    password,
    mfaCode,
  });
}

/**
 * Start MFA enrolment for a freshly provisioned agency user. The user id comes
 * from the provisioning administrator; the password entered at sign-in is the
 * account's temporary password at this point.
 */
export function setupAgencyMfa(
  userId: string,
  email: string,
  temporaryPassword: string,
): Promise<AgencyMFASetupResponse> {
  return post<AgencyMFASetupResponse>(
    `/auth/agency-users/${encodeURIComponent(userId)}/mfa/setup`,
    { email, temporaryPassword },
  );
}

/** Complete MFA enrolment by confirming the challenge code. */
export function verifyAgencyMfa(
  userId: string,
  email: string,
  temporaryPassword: string,
  code: string,
): Promise<AgencyMFAVerifyResponse> {
  return post<AgencyMFAVerifyResponse>(
    `/auth/agency-users/${encodeURIComponent(userId)}/mfa/verify`,
    { email, temporaryPassword, code },
  );
}
