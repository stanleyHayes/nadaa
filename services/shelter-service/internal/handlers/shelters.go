package handlers

import (
	"net/http"

	"github.com/stanleyHayes/nadaa/services/shelter-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/shelter-service/internal/utils"
)

func (s *server) listSheltersHandler(w http.ResponseWriter, _ *http.Request) {
	utils.WriteJSON(w, http.StatusOK, models.ShelterListResponse{Shelters: s.store.ListShelters(), GeneratedAt: s.now().UTC()})
}

func (s *server) nearbySheltersHandler(w http.ResponseWriter, r *http.Request) {
	location, ok := utils.ParseLocation(w, r)
	if !ok {
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.NearbyShelterResponse{
		Shelters:        s.store.NearbyShelters(location, utils.DefaultNearbyLimit),
		RecoverySupport: s.store.NearbyRecoverySupport(location, utils.DefaultNearbyLimit),
		GeneratedAt:     s.now().UTC(),
	})
}

func (s *server) nearbyRecoverySupportHandler(w http.ResponseWriter, r *http.Request) {
	location, ok := utils.ParseLocation(w, r)
	if !ok {
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.RecoverySupportResponse{
		RecoverySupport: s.store.NearbyRecoverySupport(location, utils.DefaultNearbyLimit),
		GeneratedAt:     s.now().UTC(),
	})
}

func (s *server) updateShelterOccupancyHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r, utils.ShelterUpdateRoles)
	if !ok {
		return
	}

	var request models.OccupancyUpdateRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := utils.NormalizeOccupancyUpdate(request)
	if code != "" {
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	shelter, code, message := s.store.UpdateShelter(r.PathValue("id"), normalized, ctx, s.now().UTC())
	if code != "" {
		utils.WriteError(w, utils.StatusForCode(code), code, message)
		return
	}
	utils.WriteJSON(w, http.StatusOK, models.ShelterUpdateResponse{Shelter: shelter})
}
