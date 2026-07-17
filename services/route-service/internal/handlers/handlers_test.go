package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/nadaa/services/route-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/route-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/route-service/internal/store"
	"github.com/stanleyHayes/nadaa/services/route-service/internal/utils"
)

func newTestConfig() *config.Config {
	return &config.Config{
		Addr:                  ":8096",
		RoadClosureServiceURL: "http://closures.test",
		ShelterServiceURL:     "http://shelters.test",
		RiskServiceURL:        "http://risk.test",
		AllowedOrigins:        nil,
	}
}

func newTestServer(responses map[string]mockResponse) *Server {
	srv, _ := newTestServerWithTransport(responses)
	return srv
}

func newTestServerWithTransport(responses map[string]mockResponse) (*Server, *mockTransport) {
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	cfg := newTestConfig()
	srv := NewServer(store.NewMemoryStore(now), func() time.Time { return now }, cfg)
	transport := &mockTransport{responses: responses, authorization: map[string]string{}}
	srv.httpClient = &http.Client{Transport: transport}
	return srv, transport
}

type mockResponse struct {
	status int
	body   string
}

type mockTransport struct {
	responses map[string]mockResponse
	// authorization records the Authorization header seen per request path.
	authorization map[string]string
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	path := req.URL.Path
	if m.authorization != nil {
		m.authorization[path] = req.Header.Get("Authorization")
	}
	response, ok := m.responses[path]
	if !ok {
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(strings.NewReader(`{"error":"not found"}`)),
			Header:     http.Header{"Content-Type": []string{"application/json"}},
		}, nil
	}
	return &http.Response{
		StatusCode: response.status,
		Body:       io.NopCloser(strings.NewReader(response.body)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}, nil
}

func TestHealthHandler(t *testing.T) {
	srv := newTestServer(nil)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health", nil)

	srv.healthHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
}

func TestOptionsHandler(t *testing.T) {
	srv := newTestServer(nil)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/routes/options", nil)

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var payload models.OptionsResponse
	decodeResponse(t, response, &payload)
	if len(payload.WaypointTypes) != 3 {
		t.Fatalf("expected 3 waypoint types, got %#v", payload.WaypointTypes)
	}
}

func TestPlanRouteValidReturnsRoute(t *testing.T) {
	responses := map[string]mockResponse{
		"/api/v1/road-closures": {status: http.StatusOK, body: `{"closures":[],"generatedAt":"2026-07-06T12:00:00Z"}`},
		"/api/v1/risk":          {status: http.StatusOK, body: `{"overallRisk":"low"}`},
	}
	srv := newTestServer(responses)

	body := models.RoutePlanRequest{
		Origin:       models.Coordinates{Lat: 5.6037, Lng: -0.1870},
		Destination:  &models.Coordinates{Lat: 5.6100, Lng: -0.1800},
		WaypointType: "manual",
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/routes/plan", jsonBody(body))
	request.Header.Set("Content-Type", "application/json")

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.RoutePlanResponse
	decodeResponse(t, response, &payload)
	if len(payload.Route) < 2 {
		t.Fatalf("expected route with at least origin and destination, got %#v", payload.Route)
	}
	if payload.DistanceMeters <= 0 {
		t.Fatalf("expected positive distance, got %d", payload.DistanceMeters)
	}
	if !payload.DecisionSupport {
		t.Fatalf("expected decisionSupport to be true")
	}
	if payload.Disclaimer == "" {
		t.Fatalf("expected disclaimer")
	}
}

func TestPlanRouteInvalidRequestReturns400(t *testing.T) {
	srv := newTestServer(nil)

	body := models.RoutePlanRequest{
		Origin:       models.Coordinates{Lat: 5.6037, Lng: -0.1870},
		WaypointType: "invalid_type",
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/routes/plan", jsonBody(body))
	request.Header.Set("Content-Type", "application/json")

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestPlanRouteMissingDestinationTargetsShelter(t *testing.T) {
	responses := map[string]mockResponse{
		"/api/v1/shelters/nearby": {
			status: http.StatusOK,
			body: `{
				"shelters":[
					{"id":"shelter_001","name":"Tema Community Shelter","location":{"lat":5.6500,"lng":-0.1700},"status":"open"}
				],
				"generatedAt":"2026-07-06T12:00:00Z"
			}`,
		},
		"/api/v1/road-closures": {status: http.StatusOK, body: `{"closures":[],"generatedAt":"2026-07-06T12:00:00Z"}`},
		"/api/v1/risk":          {status: http.StatusOK, body: `{"overallRisk":"low"}`},
	}
	srv := newTestServer(responses)

	body := models.RoutePlanRequest{
		Origin:       models.Coordinates{Lat: 5.6037, Lng: -0.1870},
		WaypointType: "shelter",
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/routes/plan", jsonBody(body))
	request.Header.Set("Content-Type", "application/json")

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.RoutePlanResponse
	decodeResponse(t, response, &payload)
	if payload.TargetShelter == nil {
		t.Fatalf("expected target shelter, got nil")
	}
	if payload.TargetShelter.ID != "shelter_001" {
		t.Fatalf("expected shelter_001, got %s", payload.TargetShelter.ID)
	}
	if len(payload.Route) < 2 {
		t.Fatalf("expected route with at least two waypoints, got %#v", payload.Route)
	}
	last := payload.Route[len(payload.Route)-1]
	if last.Lat != payload.TargetShelter.Location.Lat || last.Lng != payload.TargetShelter.Location.Lng {
		t.Fatalf("expected route to end at shelter, got %#v", last)
	}
}

func TestPlanRouteShelterLookupFailureReturns502(t *testing.T) {
	responses := map[string]mockResponse{
		"/api/v1/shelters/nearby": {status: http.StatusInternalServerError, body: `{"error":"down"}`},
	}
	srv := newTestServer(responses)

	body := models.RoutePlanRequest{
		Origin:       models.Coordinates{Lat: 5.6037, Lng: -0.1870},
		WaypointType: "shelter",
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/routes/plan", jsonBody(body))
	request.Header.Set("Content-Type", "application/json")

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusBadGateway {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadGateway, response.Code, response.Body.String())
	}
	var payload models.APIError
	decodeResponse(t, response, &payload)
	if payload.Error.Code != "shelter_lookup_failed" {
		t.Fatalf("expected shelter_lookup_failed, got %s", payload.Error.Code)
	}
}

func TestPlanRouteNoOpenShelterReturns404(t *testing.T) {
	responses := map[string]mockResponse{
		"/api/v1/shelters/nearby": {
			status: http.StatusOK,
			body: `{
				"shelters":[
					{"id":"shelter_001","name":"Closed Shelter","location":{"lat":5.6500,"lng":-0.1700},"status":"closed"}
				],
				"generatedAt":"2026-07-06T12:00:00Z"
			}`,
		},
	}
	srv := newTestServer(responses)

	body := models.RoutePlanRequest{
		Origin:       models.Coordinates{Lat: 5.6037, Lng: -0.1870},
		WaypointType: "shelter",
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/routes/plan", jsonBody(body))
	request.Header.Set("Content-Type", "application/json")

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d: %s", http.StatusNotFound, response.Code, response.Body.String())
	}
	var payload models.APIError
	decodeResponse(t, response, &payload)
	if payload.Error.Code != "no_shelter_available" {
		t.Fatalf("expected no_shelter_available, got %s", payload.Error.Code)
	}
}

func TestPlanRouteAvoidsClosureCrossingSegmentMidpoint(t *testing.T) {
	// A long two-point closure crossing the route's midpoint: its vertices are
	// over a kilometre from the route line, so vertex-only distance checks
	// never flag it — only point-to-segment distance does.
	responses := map[string]mockResponse{
		"/api/v1/road-closures": {
			status: http.StatusOK,
			body: `{
				"closures":[
					{"id":"closure_001","status":"active","severity":"high",
					 "geometry":{"type":"LineString","coordinates":[[-0.18,5.59],[-0.18,5.61]]}}
				],
				"generatedAt":"2026-07-06T12:00:00Z"
			}`,
		},
		"/api/v1/risk": {status: http.StatusOK, body: `{"overallRisk":"low"}`},
	}
	srv := newTestServer(responses)

	body := models.RoutePlanRequest{
		Origin:       models.Coordinates{Lat: 5.6000, Lng: -0.1900},
		Destination:  &models.Coordinates{Lat: 5.6000, Lng: -0.1700},
		WaypointType: "manual",
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/routes/plan", jsonBody(body))
	request.Header.Set("Content-Type", "application/json")

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var payload models.RoutePlanResponse
	decodeResponse(t, response, &payload)
	if len(payload.AvoidedClosures) != 1 || payload.AvoidedClosures[0] != "closure_001" {
		t.Fatalf("expected closure_001 to be avoided, got %#v", payload.AvoidedClosures)
	}
	if len(payload.Route) != 3 {
		t.Fatalf("expected a detour waypoint around the closure, got %#v", payload.Route)
	}
}

func TestPlanRouteFlagsSampledRiskZones(t *testing.T) {
	responses := map[string]mockResponse{
		"/api/v1/road-closures": {status: http.StatusOK, body: `{"closures":[],"generatedAt":"2026-07-06T12:00:00Z"}`},
		"/api/v1/risk":          {status: http.StatusOK, body: `{"overallRisk":"severe"}`},
	}
	srv := newTestServer(responses)

	body := models.RoutePlanRequest{
		Origin:       models.Coordinates{Lat: 5.6000, Lng: -0.1900},
		Destination:  &models.Coordinates{Lat: 5.6000, Lng: -0.1700},
		WaypointType: "manual",
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/routes/plan", jsonBody(body))
	request.Header.Set("Content-Type", "application/json")

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var payload models.RoutePlanResponse
	decodeResponse(t, response, &payload)
	if len(payload.AvoidedRiskZones) == 0 {
		t.Fatalf("expected sampled severe risk to be flagged, got %#v", payload.AvoidedRiskZones)
	}
	if len(payload.Route) < 3 {
		t.Fatalf("expected a detour waypoint around the risk zone, got %#v", payload.Route)
	}
}

func TestPlanRouteRiskLookupFailureStillReturnsRoute(t *testing.T) {
	responses := map[string]mockResponse{
		"/api/v1/road-closures": {status: http.StatusOK, body: `{"closures":[],"generatedAt":"2026-07-06T12:00:00Z"}`},
		"/api/v1/risk":          {status: http.StatusInternalServerError, body: `{"error":"down"}`},
	}
	srv := newTestServer(responses)

	body := models.RoutePlanRequest{
		Origin:       models.Coordinates{Lat: 5.6037, Lng: -0.1870},
		Destination:  &models.Coordinates{Lat: 5.6100, Lng: -0.1800},
		WaypointType: "manual",
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/routes/plan", jsonBody(body))
	request.Header.Set("Content-Type", "application/json")

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var payload models.RoutePlanResponse
	decodeResponse(t, response, &payload)
	if len(payload.AvoidedRiskZones) != 0 {
		t.Fatalf("expected no avoided risk zones on degraded lookup, got %#v", payload.AvoidedRiskZones)
	}
}

func TestPlanRouteForwardsAuthorizationHeader(t *testing.T) {
	responses := map[string]mockResponse{
		"/api/v1/road-closures": {status: http.StatusOK, body: `{"closures":[],"generatedAt":"2026-07-06T12:00:00Z"}`},
		"/api/v1/risk":          {status: http.StatusOK, body: `{"overallRisk":"low"}`},
	}
	srv, transport := newTestServerWithTransport(responses)

	body := models.RoutePlanRequest{
		Origin:       models.Coordinates{Lat: 5.6037, Lng: -0.1870},
		Destination:  &models.Coordinates{Lat: 5.6100, Lng: -0.1800},
		WaypointType: "manual",
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/routes/plan", jsonBody(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer nadaa.test.token")

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	if got := transport.authorization["/api/v1/road-closures"]; got != "Bearer nadaa.test.token" {
		t.Fatalf("expected Authorization to be forwarded to road-closure-service, got %q", got)
	}
	if got := transport.authorization["/api/v1/risk"]; got != "Bearer nadaa.test.token" {
		t.Fatalf("expected Authorization to be forwarded to risk-service, got %q", got)
	}
}

func TestMinDistanceToLineStringMeasuresSegments(t *testing.T) {
	line := [][]float64{{-0.18, 5.59}, {-0.18, 5.61}} // GeoJSON [lng, lat]

	// On the segment midpoint: vertex-only distance would be ~1.1 km.
	if d := utils.MinDistanceToLineString(models.Coordinates{Lat: 5.60, Lng: -0.18}, line); d > 1 {
		t.Fatalf("expected near-zero distance at segment midpoint, got %.2f", d)
	}
	// Just off the midpoint: ~111 m perpendicular offset.
	if d := utils.MinDistanceToLineString(models.Coordinates{Lat: 5.60, Lng: -0.179}, line); d < 50 || d > 200 {
		t.Fatalf("expected ~111 m perpendicular distance, got %.2f", d)
	}
	// Beyond both vertices: nearest point is a vertex.
	if d := utils.MinDistanceToLineString(models.Coordinates{Lat: 5.65, Lng: -0.18}, line); d < 4000 {
		t.Fatalf("expected distance beyond segment to reach the vertex, got %.2f", d)
	}
	if d := utils.MinDistanceToLineString(models.Coordinates{Lat: 5.60, Lng: -0.18}, nil); d != math.MaxFloat64 {
		t.Fatalf("expected MaxFloat64 for empty geometry, got %.2f", d)
	}
}

func TestDistanceMetersClampsAntipodal(t *testing.T) {
	a := models.Coordinates{Lat: 5.6037, Lng: -0.1870}
	antipode := models.Coordinates{Lat: -5.6037, Lng: 179.8130}

	d := utils.DistanceMeters(a, antipode)
	if math.IsNaN(d) {
		t.Fatalf("expected no NaN for antipodal points")
	}
	if halfCircumference := math.Pi * utils.EarthRadiusMeters; d < halfCircumference*0.999 || d > halfCircumference*1.001 {
		t.Fatalf("expected near half-circumference distance, got %.2f", d)
	}
	if same := utils.DistanceMeters(a, a); same != 0 {
		t.Fatalf("expected zero distance for identical points, got %.2f", same)
	}
}

func TestCorridorBBoxCoversOriginAndDestination(t *testing.T) {
	origin := models.Coordinates{Lat: 5.60, Lng: -0.19}
	destination := models.Coordinates{Lat: 5.65, Lng: -0.15}

	parts := strings.Split(utils.CorridorBBox(origin, destination, 1000), ",")
	if len(parts) != 4 {
		t.Fatalf("expected 4 bbox values, got %#v", parts)
	}
	values := make([]float64, 4)
	for i, part := range parts {
		parsed, err := strconv.ParseFloat(part, 64)
		if err != nil {
			t.Fatalf("bbox value %q is not a float: %v", part, err)
		}
		values[i] = parsed
	}
	minLng, minLat, maxLng, maxLat := values[0], values[1], values[2], values[3]
	if minLng >= origin.Lng || maxLng <= destination.Lng || minLat >= origin.Lat || maxLat <= destination.Lat {
		t.Fatalf("expected bbox to cover origin and destination with padding, got %v", values)
	}
}

func TestVerifyAuthToken(t *testing.T) {
	secret := "test-secret-with-at-least-32-bytes"
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)

	valid := signTestToken(secret, map[string]any{
		"sub":      "user_001",
		"typ":      "agency",
		"role":     "dispatcher",
		"agencyId": "agency_001",
		"district": "tema",
		"mfa":      true,
		"exp":      now.Add(time.Hour).Unix(),
	})

	claims, err := utils.VerifyAuthToken(secret, valid, now)
	if err != nil {
		t.Fatalf("expected valid token to verify, got %v", err)
	}
	if claims.UserID != "user_001" || claims.Role != "dispatcher" || claims.AgencyID != "agency_001" ||
		claims.District != "tema" || !claims.MFA {
		t.Fatalf("unexpected claims: %+v", claims)
	}

	if _, err := utils.VerifyAuthToken("wrong-secret", valid, now); err == nil {
		t.Fatalf("expected wrong secret to fail verification")
	}

	expired := signTestToken(secret, map[string]any{
		"sub": "user_001",
		"exp": now.Add(-time.Hour).Unix(),
	})
	if _, err := utils.VerifyAuthToken(secret, expired, now); err == nil {
		t.Fatalf("expected expired token to fail verification")
	}

	if _, err := utils.VerifyAuthToken(secret, "not-a-token", now); err == nil {
		t.Fatalf("expected malformed token to fail verification")
	}

	if _, err := utils.VerifyAuthToken("", valid, now); err == nil {
		t.Fatalf("expected empty secret to fail verification")
	}
}

func TestCORSLocalhostOriginRequiresDevelopmentEnv(t *testing.T) {
	allowed := map[string]bool{"https://citizen.nadaa.example": true}
	handler := utils.WithCORS(allowed, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	requestWithOrigin := func(origin string) *httptest.ResponseRecorder {
		request := httptest.NewRequest(http.MethodGet, "/routes/options", nil)
		request.Header.Set("Origin", origin)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		return response
	}

	// Allowlisted origins are always echoed.
	if got := requestWithOrigin("https://citizen.nadaa.example").Header().Get("Access-Control-Allow-Origin"); got != "https://citizen.nadaa.example" {
		t.Fatalf("expected allowlisted origin to be echoed, got %q", got)
	}
	// Localhost origins are not echoed when an allowlist is configured outside development.
	if got := requestWithOrigin("http://localhost:3000").Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected localhost origin to be rejected, got %q", got)
	}
	// In development mode the localhost exception applies.
	t.Setenv("NADAA_ENV", "development")
	if got := requestWithOrigin("http://localhost:3000").Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("expected localhost origin to be echoed in development, got %q", got)
	}
}

func signTestToken(secret string, claims map[string]any) string {
	payload, err := json.Marshal(claims)
	if err != nil {
		panic(err)
	}
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(encoded))
	return "nadaa." + encoded + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func jsonBody(value any) *bytes.Reader {
	body, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return bytes.NewReader(body)
}

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}
