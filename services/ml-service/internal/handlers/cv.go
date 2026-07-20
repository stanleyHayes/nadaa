package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

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

func (s *server) listCVResultsHandler(w http.ResponseWriter, r *http.Request) {
	limit, offset, ok := parsePagination(w, r)
	if !ok {
		return
	}
	results, total := s.store.ListCVResults(limit, offset)
	utils.WriteJSON(w, http.StatusOK, models.CVResultListResponse{
		Results: results,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
	})
}

// reviewCVResultHandler records a human review decision on a CV result. The
// reviewer identity always comes from a verified agency bearer token — never
// from service tokens or request-body fields — so the audit trail is
// attributable.
func (s *server) reviewCVResultHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.requireAgencyReviewer(w, r)
	if !ok {
		return
	}

	var request models.CVReviewRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid_request", "request body must be valid JSON")
		return
	}
	request.Decision = strings.ToLower(strings.TrimSpace(request.Decision))
	if request.Decision != "approved" && request.Decision != "rejected" {
		utils.WriteError(w, http.StatusBadRequest, "invalid_decision", "decision must be approved or rejected")
		return
	}

	result, found := s.store.ReviewCVResult(r.PathValue("id"), request, claims.UserID, s.now().UTC())
	if !found {
		utils.WriteError(w, http.StatusNotFound, "not_found", "CV result was not found for the given id")
		return
	}
	utils.WriteJSON(w, http.StatusOK, models.CVResultDetailResponse{Result: result})
}

// requireAgencyReviewer verifies the caller's bearer token carries an
// agency/dispatcher role and returns its claims.
func (s *server) requireAgencyReviewer(w http.ResponseWriter, r *http.Request) (utils.TokenClaims, bool) {
	token, present := utils.BearerToken(r)
	if !present {
		utils.WriteError(w, http.StatusUnauthorized, "unauthorized", "a verified agency bearer token is required to review CV results")
		return utils.TokenClaims{}, false
	}
	claims, err := utils.VerifyToken(token, s.config.TokenSecret, s.now())
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "unauthorized", "a verified agency bearer token is required to review CV results")
		return utils.TokenClaims{}, false
	}
	if !internalAccessRoles[normalizeRole(claims.Role)] {
		utils.WriteError(w, http.StatusForbidden, "forbidden", "the verified token does not carry an agency or dispatcher role")
		return utils.TokenClaims{}, false
	}
	return claims, true
}
