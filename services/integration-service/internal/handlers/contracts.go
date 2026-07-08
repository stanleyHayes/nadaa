package handlers

import (
	"net/http"

	"github.com/stanleyHayes/nadaa/services/integration-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/integration-service/internal/utils"
)

var allowedDirections = map[string]bool{
	"inbound":       true,
	"outbound":      true,
	"bidirectional": true,
}

var allowedDomains = map[string]bool{
	"weather":           true,
	"hydrology":         true,
	"incident_sync":     true,
	"alert_sync":        true,
	"road_closure":      true,
	"hospital_capacity": true,
	"utility_outage":    true,
	"shelter_status":    true,
}

func (s *server) listContractsHandler(w http.ResponseWriter, r *http.Request) {
	domain := utils.NormalizeQueryValue(r.URL.Query().Get("domain"))
	direction := utils.NormalizeQueryValue(r.URL.Query().Get("direction"))
	partner := utils.NormalizeQueryValue(r.URL.Query().Get("partner"))

	if domain != "" && !allowedDomains[domain] {
		utils.WriteError(w, http.StatusBadRequest, "invalid_domain", "domain must be a supported integration domain")
		return
	}
	if direction != "" && !allowedDirections[direction] {
		utils.WriteError(w, http.StatusBadRequest, "invalid_direction", "direction must be inbound, outbound, or bidirectional")
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.ContractListResponse{Contracts: s.store.ListContracts(domain, direction, partner)})
}
