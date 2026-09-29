package handlers

import (
	"os"
	"strings"
)

// brandName is the product name used in outbound citizen copy — SMS, USSD,
// WhatsApp and voice.
//
// The Go services are 18 independent modules with no shared package, and they
// cannot import the TypeScript brand package that the web and mobile clients
// resolve their identity from. The name therefore arrives by environment, with
// the current identity as the default so a service that is deployed without the
// variable still sends correct copy rather than an empty string.
//
// SUBVENIO_BRAND_NAME is a NEW variable, deliberately not one of the existing
// NADAA_* ones: those are wire configuration that services authenticate each
// other with, and they must not move with the brand.
//
// Keep the default in step with ACTIVE_BRAND in packages/brand/src/identity.ts.
func brandName() string {
	if v := strings.TrimSpace(os.Getenv("SUBVENIO_BRAND_NAME")); v != "" {
		return v
	}
	return "SUBVENIO"
}
