package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/stanleyHayes/nadaa/services/incident-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/incident-service/internal/utils"
)

func (s *server) initiateMediaUploadHandler(w http.ResponseWriter, r *http.Request) {
	var request models.InitiateMediaUploadRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeMediaUploadRequest(request)
	if code != "" {
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	record := s.store.CreateMediaUpload(normalized, s.now())
	utils.WriteJSON(w, http.StatusCreated, models.MediaUploadResponse{
		MediaID:   record.ID,
		UploadURL: record.UploadURL,
		Method:    "PUT",
		Headers: map[string]string{
			"Content-Type": record.ContentType,
		},
		ExpiresAt:    record.ExpiresAt,
		MaxSizeBytes: utils.AllowedMediaTypes[record.ContentType],
		Access:       record.Access,
	})
}

func (s *server) listMediaHandler(w http.ResponseWriter, _ *http.Request) {
	utils.WriteJSON(w, http.StatusOK, models.MediaListResponse{Media: s.store.ListMedia()})
}

func normalizeMediaUploadRequest(request models.InitiateMediaUploadRequest) (models.InitiateMediaUploadRequest, string, string) {
	request.Purpose = strings.TrimSpace(strings.ToLower(request.Purpose))
	request.FileName = strings.TrimSpace(request.FileName)
	request.ContentType = strings.TrimSpace(strings.ToLower(request.ContentType))
	request.UploadedBy = strings.TrimSpace(request.UploadedBy)

	if request.Purpose == "" {
		request.Purpose = "incident_media"
	}
	if request.Purpose != "incident_media" {
		return request, "invalid_purpose", "purpose must be incident_media"
	}

	if !utils.ValidFileName(request.FileName) {
		return request, "invalid_file_name", "fileName must be 1 to 180 safe characters without path separators"
	}

	maxSize, ok := utils.AllowedMediaTypes[request.ContentType]
	if !ok {
		return request, "unsupported_media_type", "contentType must be a supported image, video, or audio type"
	}

	if request.SizeBytes <= 0 || request.SizeBytes > maxSize {
		return request, "invalid_file_size", fmt.Sprintf("sizeBytes must be between 1 and %d for %s", maxSize, request.ContentType)
	}

	if request.UploadedBy != "" && !utils.MediaRefPattern.MatchString(request.UploadedBy) {
		return request, "invalid_uploaded_by", "uploadedBy must be a safe user reference when supplied"
	}

	return request, "", ""
}
