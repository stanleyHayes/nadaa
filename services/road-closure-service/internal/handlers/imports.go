package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/stanleyHayes/nadaa/services/road-closure-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/road-closure-service/internal/utils"
)

func (s *Server) importAdapterHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r, closureUpdateRoles)
	if !ok {
		return
	}

	var request models.AdapterImportRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		log.Printf("WARN road-closure-service adapter_import invalid_json actor=%s error=%v", ctx.ActorUserID, err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeAdapterImport(request)
	if code != "" {
		log.Printf("WARN road-closure-service adapter_import validation_failed actor=%s code=%s", ctx.ActorUserID, code)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	closures := s.store.ImportAdapter(normalized, ctx, s.now().UTC())
	log.Printf("INFO road-closure-service adapter_import completed actor=%s source=%s imported=%d", ctx.ActorUserID, normalized.Source, len(closures))
	utils.WriteJSON(w, http.StatusOK, models.AdapterImportResponse{
		Imported:    len(closures),
		Closures:    closures,
		GeneratedAt: s.now().UTC(),
		Source:      normalized.Source,
	})
}

func normalizeAdapterImport(request models.AdapterImportRequest) (models.AdapterImportRequest, string, string) {
	request.Source = utils.NormalizeQueryValue(request.Source)
	request.SourceRef = strings.TrimSpace(request.SourceRef)
	request.RoadName = strings.TrimSpace(request.RoadName)
	request.Status = utils.NormalizeQueryValue(request.Status)
	request.Reason = strings.TrimSpace(request.Reason)
	request.Detour = strings.TrimSpace(request.Detour)

	if request.Source == "" {
		request.Source = "adapter"
	}
	if len(request.Source) > 80 || utils.UnsafeText(request.Source) {
		return request, "invalid_source", "source must be 80 safe characters or fewer"
	}
	if len(request.SourceRef) > 120 || utils.UnsafeText(request.SourceRef) {
		return request, "invalid_source_ref", "sourceRef must be 120 safe characters or fewer"
	}
	if request.RoadName == "" || len(request.RoadName) > 200 || utils.UnsafeText(request.RoadName) {
		return request, "invalid_road_name", "roadName is required and must be 200 safe characters or fewer"
	}
	if !allowedClosureStatuses[request.Status] {
		return request, "invalid_status", "status must be active, scheduled, lifted, or cancelled"
	}
	geometry, err := utils.ParseWKTLineString(request.Geometry)
	if err != nil {
		return request, "invalid_geometry", err.Error()
	}
	if errCode, errMsg := utils.ValidateGeometry(geometry); errCode != "" {
		return request, errCode, errMsg
	}
	request.Geometry = utils.FormatWKTLineString(geometry)
	if len(request.Reason) > 200 || utils.UnsafeText(request.Reason) {
		return request, "invalid_reason", "reason must be 200 safe characters or fewer"
	}
	if len(request.Detour) > 500 || utils.UnsafeText(request.Detour) {
		return request, "invalid_detour", "detour must be 500 safe characters or fewer"
	}
	if request.ValidFrom.IsZero() {
		return request, "missing_valid_from", "validFrom is required"
	}
	if request.ValidTo != nil && request.ValidTo.Before(request.ValidFrom) {
		return request, "invalid_valid_to", "validTo must be after validFrom"
	}
	return request, "", ""
}
