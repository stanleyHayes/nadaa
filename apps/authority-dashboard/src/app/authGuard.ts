let installed = false;

/**
 * Send the user back to the login screen whenever their session is no longer
 * valid. Wraps window.fetch once: any 401 response clears the session via
 * `onUnauthorized`, which trips the app's auth gate so the sign-in screen shows.
 */
export function installUnauthorizedRedirect(onUnauthorized: () => void): void {
  if (installed || typeof window === "undefined") {
    return;
  }
  installed = true;
  const nativeFetch = window.fetch.bind(window);
  window.fetch = async (input: RequestInfo | URL, init?: RequestInit) => {
    const response = await nativeFetch(input, init);
    if (response.status === 401) {
      onUnauthorized();
    }
    return response;
  };
}
