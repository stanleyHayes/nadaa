package handlers

import (
	"net/http"

	"github.com/stanleyHayes/nadaa/services/incident-service/internal/utils"
)

// Routes returns the configured HTTP handler with middleware applied.
func (s *server) Routes(allowedOrigins map[string]bool) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthHandler)
	mux.HandleFunc("POST /api/v1/incidents", s.createIncidentHandler)
	mux.HandleFunc("GET /api/v1/incidents", s.listIncidentsHandler)
	mux.HandleFunc("GET /api/v1/incidents/{id}", s.getIncidentHandler)
	mux.HandleFunc("GET /api/v1/incidents/{id}/duplicates", s.duplicateReviewHandler)
	mux.HandleFunc("GET /api/v1/incidents/{id}/triage", s.suggestTriageHandler)
	mux.HandleFunc("POST /api/v1/incidents/{id}/triage-review", s.reviewTriageHandler)
	mux.HandleFunc("GET /api/v1/incidents/audit", s.listIncidentAuditHandler)
	mux.HandleFunc("POST /api/v1/incidents/{id}/merge", s.mergeIncidentHandler)
	mux.HandleFunc("POST /api/v1/incidents/{id}/abuse-review", s.reviewAbuseHandler)
	mux.HandleFunc("PATCH /api/v1/incidents/{id}/status", s.updateIncidentStatusHandler)
	mux.HandleFunc("POST /api/v1/incidents/{id}/verify", s.verifyIncidentHandler)
	mux.HandleFunc("POST /api/v1/incidents/{id}/assignments", s.assignIncidentHandler)
	mux.HandleFunc("POST /api/v1/incidents/{id}/volunteer-tasks", s.assignVolunteerTaskHandler)
	mux.HandleFunc("POST /api/v1/media/uploads", s.initiateMediaUploadHandler)
	mux.HandleFunc("PUT /api/v1/media/{id}/content", s.putMediaContentHandler)
	mux.HandleFunc("GET /api/v1/media/{id}/content", s.getMediaContentHandler)
	mux.HandleFunc("GET /api/v1/media", s.listMediaHandler)
	mux.HandleFunc("POST /api/v1/volunteers", s.registerVolunteerHandler)
	mux.HandleFunc("GET /api/v1/volunteers", s.listVolunteersHandler)
	mux.HandleFunc("POST /api/v1/volunteers/{id}/verify", s.verifyVolunteerHandler)
	mux.HandleFunc("GET /api/v1/volunteers/{id}/tasks", s.listVolunteerTasksHandler)
	mux.HandleFunc("PATCH /api/v1/volunteer-tasks/{id}/status", s.updateVolunteerTaskStatusHandler)
	mux.HandleFunc("POST /api/v1/volunteer-tasks/{id}/observations", s.submitVolunteerObservationHandler)
	return utils.WithCORS(allowedOrigins, mux)
}
