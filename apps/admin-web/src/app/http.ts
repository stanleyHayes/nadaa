/**
 * Raised when a request is rejected because the session is no longer valid.
 * Callers can let it propagate — the session has already been cleared, so the
 * app returns to the sign-in screen on its own.
 */
export class SessionExpiredError extends Error {
  constructor() {
    super("Your session has expired. Sign in again to continue.");
    this.name = "SessionExpiredError";
  }
}

/**
 * Turn a 401 into a re-authentication instead of an error banner. A 401 means
 * the actor is no longer authorized, so we clear the session — which re-renders
 * the app to the sign-in gate — and throw so the in-flight request unwinds.
 * Call this before treating a non-OK response as a data error.
 */
export function handleUnauthorized(response: Response): void {
  if (response.status === 401) {
    // The global 401 guard (installUnauthorizedRedirect) already clears the
    // session for every fetch; just unwind this request.
    throw new SessionExpiredError();
  }
}
