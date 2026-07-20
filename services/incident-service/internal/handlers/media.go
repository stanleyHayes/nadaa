package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/stanleyHayes/nadaa/services/incident-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/incident-service/internal/utils"
)

func (s *server) initiateMediaUploadHandler(w http.ResponseWriter, r *http.Request) {
	clientID := utils.ClientIdentifier(r)
	if !s.rateLimiter.Allow(clientID) {
		utils.WriteError(w, http.StatusTooManyRequests, "rate_limited", "too many media upload requests; please wait before trying again")
		return
	}

	var request models.InitiateMediaUploadRequest
	if err := utils.DecodeJSON(w, r, &request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeMediaUploadRequest(request)
	if code != "" {
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	// Initiation stays open to anonymous reporters, but when the caller carries
	// a verified bearer token its subject becomes the uploader of record so the
	// byte upload endpoint can require the same identity later.
	if token, ok := bearerToken(r); ok {
		if claims, err := verifyAuthToken(s.tokenSecret, s.now, token); err == nil && strings.TrimSpace(claims.UserID) != "" {
			normalized.UploadedBy = strings.TrimSpace(claims.UserID)
		}
	}

	record := s.store.CreateMediaUpload(normalized, s.publicBaseURL, s.now())
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

// putMediaContentHandler stores the uploaded bytes for a pending media record
// on disk, capped at the content type's allowed size.
func (s *server) putMediaContentHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	record, found := s.store.MediaByID(id)
	if !found {
		utils.WriteError(w, http.StatusNotFound, "not_found", "media record was not found")
		return
	}
	if !s.requireMediaContentActor(w, r, record) {
		return
	}
	if record.Status == "linked" {
		utils.WriteError(w, http.StatusBadRequest, "media_already_linked", "media is already linked to an incident and cannot be replaced")
		return
	}
	if !s.now().Before(record.ExpiresAt) {
		utils.WriteError(w, http.StatusBadRequest, "upload_expired", "upload URL has expired; initiate a new media upload")
		return
	}

	maxSize, ok := utils.AllowedMediaTypes[record.ContentType]
	if !ok {
		utils.WriteError(w, http.StatusBadRequest, "unsupported_media_type", "media record has an unsupported content type")
		return
	}

	data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxSize))
	if err != nil {
		if strings.Contains(err.Error(), "too large") {
			utils.WriteError(w, http.StatusRequestEntityTooLarge, "payload_too_large", fmt.Sprintf("file exceeds %d byte limit for %s", maxSize, record.ContentType))
			return
		}
		// #nosec G706 -- media id is sanitized with utils.SafeLogValue.
		log.Printf("ERROR incident-service media_content_read_failed id=%s error=%v", utils.SafeLogValue(id), err)
		utils.WriteError(w, http.StatusInternalServerError, "read_failed", "could not read uploaded file")
		return
	}

	if err := os.MkdirAll(s.mediaStoragePath, 0o750); err != nil {
		log.Printf("ERROR incident-service media_content_mkdir_failed path=%s error=%v", utils.SafeLogValue(s.mediaStoragePath), err)
		utils.WriteError(w, http.StatusInternalServerError, "write_failed", "could not save uploaded file")
		return
	}
	storagePath := filepath.Join(s.mediaStoragePath, record.ID)
	if err := os.WriteFile(storagePath, data, 0o600); err != nil {
		// #nosec G706 -- media id and path are sanitized with utils.SafeLogValue.
		log.Printf("ERROR incident-service media_content_write_failed id=%s path=%s error=%v", utils.SafeLogValue(id), utils.SafeLogValue(storagePath), err)
		utils.WriteError(w, http.StatusInternalServerError, "write_failed", "could not save uploaded file")
		return
	}

	record, code, message := s.store.CompleteMediaUpload(id, int64(len(data)))
	if code != "" {
		_ = os.Remove(storagePath)
		utils.WriteError(w, statusForCode(code), code, message)
		return
	}
	// #nosec G706 -- media id is sanitized with utils.SafeLogValue.
	log.Printf("INFO incident-service media_content_uploaded id=%s size=%d", utils.SafeLogValue(id), len(data))
	utils.WriteJSON(w, http.StatusOK, record)
}

// getMediaContentHandler returns the stored bytes for a media record.
func (s *server) getMediaContentHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	record, found := s.store.MediaByID(id)
	if !found {
		utils.WriteError(w, http.StatusNotFound, "not_found", "media record was not found")
		return
	}
	if !s.requireMediaContentActor(w, r, record) {
		return
	}

	file, err := os.Open(filepath.Join(s.mediaStoragePath, record.ID))
	if err != nil {
		// #nosec G706 -- media id is sanitized with utils.SafeLogValue.
		log.Printf("WARN incident-service media_content_missing id=%s error=%v", utils.SafeLogValue(id), err)
		utils.WriteError(w, http.StatusNotFound, "file_not_found", "stored media file was not found")
		return
	}
	defer func() { _ = file.Close() }()

	stat, err := file.Stat()
	if err != nil {
		// #nosec G706 -- media id is sanitized with utils.SafeLogValue.
		log.Printf("ERROR incident-service media_content_stat_failed id=%s error=%v", utils.SafeLogValue(id), err)
		utils.WriteError(w, http.StatusInternalServerError, "file_stat_failed", "could not stat stored media file")
		return
	}

	w.Header().Set("Content-Type", record.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", record.FileName))
	w.Header().Set("Content-Length", strconv.FormatInt(stat.Size(), 10))
	if _, err := io.Copy(w, file); err != nil {
		// #nosec G706 -- media id is sanitized with utils.SafeLogValue.
		log.Printf("ERROR incident-service media_content_copy_failed id=%s error=%v", utils.SafeLogValue(id), err)
	}
}

// requireMediaContentActor authorizes media byte access: an authority user with
// an incident read role (MFA completed), or the verified citizen whose token
// subject matches the uploader that initiated the media record.
func (s *server) requireMediaContentActor(w http.ResponseWriter, r *http.Request, record models.MediaRecord) bool {
	if s.allowMockActors && hasMockActorHeaders(r) {
		_, ok := s.requireMockAuthority(w, r, incidentReadRoles)
		return ok
	}

	token, ok := bearerToken(r)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "missing_token", "Bearer token is required")
		return false
	}
	claims, err := verifyAuthToken(s.tokenSecret, s.now, token)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "invalid_token", "token is invalid or expired")
		return false
	}

	if claims.UserType == "agency" {
		ctx := authorityContextFromClaims(claims, r)
		if ctx.ActorUserID == "" || ctx.ActorRole == "" {
			utils.WriteError(w, http.StatusUnauthorized, "missing_authority_context", "token must carry an authority user id and role")
			return false
		}
		if !ctx.MFACompleted {
			utils.WriteError(w, http.StatusForbidden, "mfa_required", "MFA must be completed for media content access")
			return false
		}
		if !incidentReadRoles[ctx.ActorRole] {
			utils.WriteError(w, http.StatusForbidden, "forbidden", "actor role is not allowed to access media content")
			return false
		}
		return true
	}

	if record.UploadedBy != "" && strings.TrimSpace(claims.UserID) == record.UploadedBy {
		return true
	}
	utils.WriteError(w, http.StatusForbidden, "forbidden", "only the initiating uploader or an authority user can access media content")
	return false
}

func (s *server) listMediaHandler(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireAuthority(w, r, incidentReadRoles); !ok {
		return
	}
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
