package handlers

import (
	"net/http"

	"github.com/stanleyHayes/nadaa/services/alert-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/alert-service/internal/utils"
)

func (s *Server) previewTargetHandler(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAuthority(w, r, utils.DraftRoles); !ok {
		return
	}

	var target models.AlertTarget
	if err := utils.DecodeJSON(r, &target); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized := utils.NormalizeTarget(target)
	if code, message := utils.ValidateTarget(normalized); code != "" {
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.TargetPreviewResponse{
		Target:   normalized,
		Summary:  utils.TargetSummary(normalized),
		Warnings: utils.TargetWarnings(normalized),
	})
}
