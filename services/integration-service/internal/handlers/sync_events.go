package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/stanleyHayes/nadaa/services/integration-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/integration-service/internal/utils"
)

func (s *server) createSyncEventHandler(w http.ResponseWriter, r *http.Request) {
	var request models.SyncRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	request.Type = utils.NormalizeQueryValue(request.Type)

	if code, message := validateSyncRequest(request); code != "" {
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	event := s.store.CreateSyncEvent(request, time.Now().UTC())
	utils.WriteJSON(w, http.StatusAccepted, event)
}

func (s *server) listSyncEventsHandler(w http.ResponseWriter, r *http.Request) {
	eventType := utils.NormalizeQueryValue(r.URL.Query().Get("type"))
	if eventType != "" && eventType != "incident" && eventType != "alert" {
		utils.WriteError(w, http.StatusBadRequest, "invalid_type", "type must be incident or alert")
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.SyncEventListResponse{Events: s.store.ListSyncEvents(eventType)})
}

func validateSyncRequest(request models.SyncRequest) (string, string) {
	if request.Type != "incident" && request.Type != "alert" {
		return "invalid_type", "type must be incident or alert"
	}
	if strings.TrimSpace(request.SourceID) == "" {
		return "missing_source_id", "sourceId is required"
	}
	if strings.TrimSpace(request.Reference) == "" {
		return "missing_reference", "reference is required"
	}
	if strings.TrimSpace(request.HazardType) == "" {
		return "missing_hazard_type", "hazardType is required"
	}
	if len(request.TargetAgencyIDs) == 0 {
		return "missing_target_agencies", "at least one targetAgencyId is required"
	}
	if strings.TrimSpace(request.CorrelationID) == "" {
		return "missing_correlation_id", "correlationId is required for idempotent sync"
	}
	return "", ""
}
