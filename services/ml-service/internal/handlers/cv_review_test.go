package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stanleyHayes/nadaa/services/ml-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/ml-service/internal/utils"
)

// analyzeImageSecured posts a CV analyze request against the secured test
// server using the service-token credential.
func analyzeImageSecured(t *testing.T, srv *server, imageID, imageName string) models.CVAnalysisResult {
	t.Helper()
	body, _ := json.Marshal(models.CVAnalysisRequest{ImageID: imageID, ImageName: imageName})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cv/analyze", bytes.NewReader(body))
	req.Header.Set(serviceTokenHeader, testServiceToken)
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("analyze: expected 200 got %d: %s", rr.Code, rr.Body.String())
	}
	var resp models.CVAnalysisResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode analyze: %v", err)
	}
	return resp.Result
}

func agencyReviewToken(t *testing.T) string {
	t.Helper()
	claims := securedTestClaims
	claims.ExpiresAt = testNow.Add(time.Hour).Unix()
	return signTestToken(t, testTokenSecret, claims)
}

func reviewCV(t *testing.T, srv *server, id string, body any, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/cv/results/"+id+"/review", bytes.NewReader(raw))
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, req)
	return rr
}

func TestCVReviewApprovePersistsDecision(t *testing.T) {
	srv := newSecuredTestServer(t)
	result := analyzeImageSecured(t, srv, "img_review_001", "flooded-road.jpg")

	// The review endpoint accepts the result ID as well as the image ID.
	rr := reviewCV(t, srv, result.ID, models.CVReviewRequest{Decision: "approved", Note: "confirmed by NADMO"},
		map[string]string{"Authorization": "Bearer " + agencyReviewToken(t)})
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200 got %d: %s", rr.Code, rr.Body.String())
	}
	var resp models.CVResultDetailResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode review: %v", err)
	}
	if resp.Result.ReviewStatus != "approved" {
		t.Fatalf("expected approved review status, got %q", resp.Result.ReviewStatus)
	}
	if resp.Result.ReviewedBy != securedTestClaims.UserID {
		t.Fatalf("expected reviewer %q from verified claims, got %q", securedTestClaims.UserID, resp.Result.ReviewedBy)
	}
	if resp.Result.ReviewedAt == "" {
		t.Fatal("expected reviewedAt to be set")
	}
	if resp.Result.ReviewNote != "confirmed by NADMO" {
		t.Fatalf("expected review note, got %q", resp.Result.ReviewNote)
	}

	// The decision must persist on the stored result.
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/cv/results/img_review_001", nil)
	getReq.Header.Set(serviceTokenHeader, testServiceToken)
	getRR := httptest.NewRecorder()
	srv.Routes().ServeHTTP(getRR, getReq)
	var got models.CVResultDetailResponse
	if err := json.Unmarshal(getRR.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode get: %v", err)
	}
	if got.Result.ReviewStatus != "approved" || got.Result.ReviewedBy != securedTestClaims.UserID || got.Result.ReviewedAt == "" {
		t.Fatalf("expected persisted review, got %+v", got.Result)
	}
}

func TestCVReviewRejectPersistsDecision(t *testing.T) {
	srv := newSecuredTestServer(t)
	analyzeImageSecured(t, srv, "img_review_002", "unclear-test.jpg")

	rr := reviewCV(t, srv, "img_review_002", models.CVReviewRequest{Decision: "rejected"},
		map[string]string{"Authorization": "Bearer " + agencyReviewToken(t)})
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200 got %d: %s", rr.Code, rr.Body.String())
	}
	var resp models.CVResultDetailResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode review: %v", err)
	}
	if resp.Result.ReviewStatus != "rejected" {
		t.Fatalf("expected rejected review status, got %q", resp.Result.ReviewStatus)
	}
}

func TestCVReviewRejectsInvalidDecision(t *testing.T) {
	srv := newSecuredTestServer(t)
	analyzeImageSecured(t, srv, "img_review_003", "flood-scene.jpg")

	rr := reviewCV(t, srv, "img_review_003", models.CVReviewRequest{Decision: "maybe"},
		map[string]string{"Authorization": "Bearer " + agencyReviewToken(t)})
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 got %d: %s", rr.Code, rr.Body.String())
	}
	var payload models.APIError
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if payload.Error.Code != "invalid_decision" {
		t.Fatalf("expected invalid_decision, got %q", payload.Error.Code)
	}
}

func TestCVReviewUnknownIDReturns404(t *testing.T) {
	srv := newSecuredTestServer(t)
	rr := reviewCV(t, srv, "cv_does_not_exist", models.CVReviewRequest{Decision: "approved"},
		map[string]string{"Authorization": "Bearer " + agencyReviewToken(t)})
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404 got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestCVReviewRequiresVerifiedAgencyBearer(t *testing.T) {
	srv := newSecuredTestServer(t)
	analyzeImageSecured(t, srv, "img_review_004", "flood-scene.jpg")

	// A service token alone cannot review: decisions must be attributable to
	// a verified agency actor.
	if rr := reviewCV(t, srv, "img_review_004", models.CVReviewRequest{Decision: "approved"},
		map[string]string{serviceTokenHeader: testServiceToken}); rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 without a bearer token got %d", rr.Code)
	}

	// A verified citizen token is authenticated but not authorized.
	citizenClaims := utils.TokenClaims{
		UserID:    "citizen_001",
		UserType:  "citizen",
		Role:      "citizen",
		ExpiresAt: testNow.Add(time.Hour).Unix(),
	}
	citizenToken := signTestToken(t, testTokenSecret, citizenClaims)
	if rr := reviewCV(t, srv, "img_review_004", models.CVReviewRequest{Decision: "approved"},
		map[string]string{
			serviceTokenHeader: testServiceToken,
			"Authorization":    "Bearer " + citizenToken,
		}); rr.Code != http.StatusForbidden {
		t.Fatalf("expected status 403 with a citizen token got %d", rr.Code)
	}
}
