package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/stanleyHayes/nadaa/services/integration-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/integration-service/internal/utils"
)

func (s *server) importRoadClosureHandler(w http.ResponseWriter, r *http.Request) {
	var request models.RoadClosureImportRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	request.Source = strings.TrimSpace(strings.ToLower(request.Source))
	request.SourceRef = strings.TrimSpace(request.SourceRef)
	request.RoadName = strings.TrimSpace(request.RoadName)
	request.Status = strings.TrimSpace(strings.ToLower(request.Status))
	request.Reason = strings.TrimSpace(request.Reason)
	request.Geometry = strings.TrimSpace(request.Geometry)
	request.Detour = strings.TrimSpace(request.Detour)

	if request.Source == "" {
		utils.WriteError(w, http.StatusBadRequest, "missing_source", "source is required")
		return
	}
	if request.RoadName == "" {
		utils.WriteError(w, http.StatusBadRequest, "missing_road_name", "roadName is required")
		return
	}
	if request.Status == "" {
		utils.WriteError(w, http.StatusBadRequest, "missing_status", "status is required")
		return
	}
	if request.Status != "active" && request.Status != "scheduled" && request.Status != "lifted" && request.Status != "cancelled" {
		utils.WriteError(w, http.StatusBadRequest, "invalid_status", "status must be active, scheduled, lifted, or cancelled")
		return
	}
	if request.Geometry == "" {
		utils.WriteError(w, http.StatusBadRequest, "missing_geometry", "geometry is required")
		return
	}
	if code, message := utils.ValidateWKTLineString(request.Geometry); code != "" {
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}
	if request.ValidFrom.IsZero() {
		utils.WriteError(w, http.StatusBadRequest, "missing_valid_from", "validFrom is required")
		return
	}
	if request.ValidTo != nil && request.ValidTo.Before(request.ValidFrom) {
		utils.WriteError(w, http.StatusBadRequest, "invalid_valid_to", "validTo must be after validFrom")
		return
	}

	if err := s.forwardRoadClosureToService(r, request); err != nil {
		var downErr *downstreamError
		if errors.As(err, &downErr) {
			log.Printf("WARN integration-service road_closure_import downstream_rejected status=%d code=%s", downErr.status, utils.SanitizeLogValue(downErr.code))
			utils.WriteError(w, downErr.status, downErr.code, downErr.message)
			return
		}
		log.Printf("WARN integration-service road_closure_import forward_failed error=%v", err)
		utils.WriteError(w, http.StatusBadGateway, "road_closure_service_unavailable", "road closure service could not accept the import")
		return
	}

	record := s.store.ImportRoadClosure(request)
	log.Printf("INFO integration-service road_closure_import accepted id=%s source=%s roadName=%s", record.ID, utils.SanitizeLogValue(record.Source), utils.SanitizeLogValue(record.RoadName))
	utils.WriteJSON(w, http.StatusAccepted, models.RoadClosureImportResponse{Imported: 1, Record: record, AcceptedAt: record.ImportedAt})
}

// downstreamError carries a client-facing rejection from road-closure-service so
// the caller sees the real status and error code instead of a blanket 502.
type downstreamError struct {
	status  int
	code    string
	message string
}

func (e *downstreamError) Error() string {
	return fmt.Sprintf("road closure service returned %d: %s", e.status, e.message)
}

func (s *server) forwardRoadClosureToService(r *http.Request, request models.RoadClosureImportRequest) error {
	payload, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("marshal road closure request: %w", err)
	}

	target, err := url.JoinPath(s.roadClosureAPIURL, "/api/v1/road-closures/imports/adapter")
	if err != nil {
		return fmt.Errorf("build road closure service URL: %w", err)
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, target, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create road closure service request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if value := r.Header.Get("X-NADAA-Request-ID"); value != "" {
		req.Header.Set("X-NADAA-Request-ID", value)
	}
	// Forward the caller's bearer token so road-closure-service verifies and
	// attributes the real actor. Legacy actor headers are only passed through
	// for local dev when mock actors are explicitly enabled; they are never
	// fabricated here.
	if authorization := strings.TrimSpace(r.Header.Get("Authorization")); authorization != "" {
		req.Header.Set("Authorization", authorization)
	} else if s.allowMockActors {
		for _, header := range []string{
			"X-NADAA-Actor-ID",
			"X-NADAA-Actor-Role",
			"X-NADAA-Agency-ID",
			"X-NADAA-MFA-Completed",
		} {
			if value := r.Header.Get(header); value != "" {
				req.Header.Set(header, value)
			}
		}
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("road closure service request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		if isClientErrorStatus(resp.StatusCode) {
			return downstreamClientError(resp.StatusCode, body)
		}
		return fmt.Errorf("road closure service returned %d: %s", resp.StatusCode, utils.SanitizeLogValue(string(body)))
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}

// isClientErrorStatus reports whether a downstream status is a client error the
// caller can act on (400/401/403/404) rather than an availability failure.
func isClientErrorStatus(status int) bool {
	switch status {
	case http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound:
		return true
	}
	return false
}

// downstreamClientError maps a downstream client-error response onto a
// downstreamError, propagating the downstream error code when it is present.
func downstreamClientError(status int, body []byte) error {
	var apiError models.APIError
	if err := json.Unmarshal(body, &apiError); err == nil && apiError.Error.Code != "" {
		return &downstreamError{status: status, code: apiError.Error.Code, message: apiError.Error.Message}
	}
	return &downstreamError{status: status, code: "road_closure_service_rejected", message: "road closure service rejected the import"}
}

func (s *server) listRoadClosureImportsHandler(w http.ResponseWriter, r *http.Request) {
	source := utils.NormalizeQueryValue(r.URL.Query().Get("source"))
	utils.WriteJSON(w, http.StatusOK, map[string]any{"imports": s.store.ListRoadClosureImports(source), "generatedAt": time.Now().UTC()})
}
