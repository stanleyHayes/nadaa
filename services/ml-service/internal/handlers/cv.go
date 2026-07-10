package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/stanleyHayes/nadaa/services/ml-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/ml-service/internal/utils"
)

func (s *server) createCVAnalysisHandler(w http.ResponseWriter, r *http.Request) {
	var request models.CVAnalysisRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid_request", "request body must be valid JSON")
		return
	}

	result, err := s.store.AnalyzeImage(request, s.now().UTC())
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "cv_analysis_failed", err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.CVAnalysisResponse{
		Result: result,
		Safety: models.SafetyPolicy{
			HumanReviewRequired: result.HumanReviewRequired,
			AutoPublishAllowed:  false,
			Message:             "CV output is decision support only and cannot trigger alerts or public actions without authority review and approval.",
		},
	})
}

func (s *server) getCVResultHandler(w http.ResponseWriter, r *http.Request) {
	imageID := r.PathValue("imageId")
	if imageID == "" {
		utils.WriteError(w, http.StatusBadRequest, "invalid_request", "imageId path parameter is required")
		return
	}

	result, ok := s.store.GetCVResult(imageID)
	if !ok {
		utils.WriteError(w, http.StatusNotFound, "not_found", "CV result was not found for the given imageId")
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.CVResultDetailResponse{Result: result})
}

func (s *server) listCVResultsHandler(w http.ResponseWriter, _ *http.Request) {
	results := s.store.ListCVResults()
	utils.WriteJSON(w, http.StatusOK, models.CVResultListResponse{Results: results})
}
