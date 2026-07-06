const requiredWebTargets = [
  ["citizen-web", "STAGING_CITIZEN_URL", "NADAA Citizen"],
  ["authority-dashboard", "STAGING_AUTHORITY_URL", "NADAA Authority Dashboard"],
  ["dispatcher-web", "STAGING_DISPATCHER_URL", "NADAA Dispatch Command"],
  ["admin-web", "STAGING_ADMIN_URL", "NADAA Admin Console"],
];

const optionalServiceTargets = [
  ["alert-service", "STAGING_ALERT_SERVICE_URL"],
  ["auth-service", "STAGING_AUTH_SERVICE_URL"],
  ["incident-service", "STAGING_INCIDENT_SERVICE_URL"],
  ["guide-service", "STAGING_GUIDE_SERVICE_URL"],
  ["integration-service", "STAGING_INTEGRATION_SERVICE_URL"],
  ["ml-service", "STAGING_ML_SERVICE_URL"],
  ["notification-service", "STAGING_NOTIFICATION_SERVICE_URL"],
  ["risk-service", "STAGING_RISK_SERVICE_URL"],
  ["shelter-service", "STAGING_SHELTER_SERVICE_URL"],
];

for (const [name, envKey, expectedTitle] of requiredWebTargets) {
  const url = requiredURL(envKey);
  const response = await fetch(url);
  if (!response.ok) {
    throw new Error(
      `${name} staging smoke failed: ${response.status} ${response.statusText}`,
    );
  }

  const html = await response.text();
  if (!html.includes(`<title>${expectedTitle}</title>`)) {
    throw new Error(`${name} staging smoke reached the wrong app at ${url}`);
  }

  console.log(`${name} staging OK ${response.status}`);
}

for (const [name, envKey] of optionalServiceTargets) {
  const baseURL = optionalURL(envKey);
  if (!baseURL) {
    console.log(`${name} staging SKIP ${envKey} not set`);
    continue;
  }

  const healthURL = new URL("/healthz", baseURL);
  const response = await fetch(healthURL);
  if (!response.ok) {
    throw new Error(
      `${name} staging health failed: ${response.status} ${response.statusText}`,
    );
  }

  console.log(`${name} staging OK ${response.status}`);
}

function requiredURL(envKey) {
  const value = optionalURL(envKey);
  if (!value) {
    throw new Error(`${envKey} is required for staging smoke tests`);
  }
  return value;
}

function optionalURL(envKey) {
  const value = process.env[envKey]?.trim();
  if (!value) {
    return "";
  }

  try {
    return new URL(value).toString();
  } catch {
    throw new Error(`${envKey} must be an absolute URL`);
  }
}
