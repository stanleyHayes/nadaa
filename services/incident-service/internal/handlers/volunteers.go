package handlers

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/stanleyHayes/nadaa/services/incident-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/incident-service/internal/store"
	"github.com/stanleyHayes/nadaa/services/incident-service/internal/utils"
)

var (
	allowedVolunteerAvailability = map[string]bool{
		"available": true,
		"busy":      true,
		"off_duty":  true,
	}
	allowedVolunteerVerificationDecisions = map[string]bool{
		"verify":  true,
		"reject":  true,
		"suspend": true,
	}
	allowedVolunteerTaskTypes = map[string]bool{
		"welfare_check":       true,
		"shelter_support":     true,
		"supply_distribution": true,
		"damage_observation":  true,
		"route_observation":   true,
		"community_alerting":  true,
	}
	allowedVolunteerTaskStatuses = map[string]bool{
		"accepted":         true,
		"en_route":         true,
		"on_scene":         true,
		"completed":        true,
		"cancelled":        true,
		"needs_escalation": true,
	}
	allowedVolunteerSafetyStatuses = map[string]bool{
		"safe":            true,
		"caution":         true,
		"unsafe":          true,
		"needs_authority": true,
	}
)

func (s *server) registerVolunteerHandler(w http.ResponseWriter, r *http.Request) {
	var request models.RegisterVolunteerRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		// #nosec G706 -- logged values are sanitized with utils.SafeLogValue.
		log.Printf("WARN incident-service volunteer_register invalid_json remote=%s error=%v", utils.SafeLogValue(utils.ClientIdentifier(r)), err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	// #nosec G706 -- logged values are sanitized with utils.SafeLogValue.
	log.Printf("INFO incident-service volunteer_register received citizenUserId=%s district=%s community=%s", utils.SafeLogValue(request.CitizenUserID), utils.SafeLogValue(request.District), utils.SafeLogValue(request.Community))

	normalized, code, message := normalizeVolunteerRegistrationRequest(request)
	if code != "" {
		// #nosec G706 -- logged values are sanitized with utils.SafeLogValue.
		log.Printf("WARN incident-service volunteer_register validation_failed citizenUserId=%s code=%s", utils.SafeLogValue(request.CitizenUserID), code)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	volunteer := s.store.RegisterVolunteer(normalized, s.now())
	log.Printf("INFO incident-service volunteer_register created volunteerId=%s groupId=%s verificationStatus=%s", volunteer.ID, volunteer.GroupID, volunteer.VerificationStatus)
	utils.WriteJSON(w, http.StatusCreated, models.VolunteerProfileResponse{Volunteer: volunteer})
}

func (s *server) listVolunteersHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r, assignmentRoles)
	if !ok {
		return
	}
	status := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("status")))
	district := strings.TrimSpace(r.URL.Query().Get("district"))
	volunteers := s.store.ListVolunteers(status, district)
	// #nosec G706 -- actor, role, status, and district are sanitized with utils.SafeLogValue.
	log.Printf("INFO incident-service volunteer_list actor=%s role=%s status=%s district=%s count=%d", utils.SafeLogValue(ctx.ActorUserID), utils.SafeLogValue(ctx.ActorRole), utils.SafeLogValue(status), utils.SafeLogValue(district), len(volunteers))
	utils.WriteJSON(w, http.StatusOK, models.VolunteerListResponse{Volunteers: volunteers})
}

func (s *server) verifyVolunteerHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r, assignmentRoles)
	if !ok {
		return
	}

	var request models.VerifyVolunteerRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		// #nosec G706 -- volunteer id and actor are sanitized with utils.SafeLogValue.
		log.Printf("WARN incident-service volunteer_verify invalid_json volunteerId=%s actor=%s error=%v", utils.SafeLogValue(r.PathValue("id")), utils.SafeLogValue(ctx.ActorUserID), err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeVolunteerVerifyRequest(request)
	if code != "" {
		// #nosec G706 -- volunteer id and actor are sanitized with utils.SafeLogValue.
		log.Printf("WARN incident-service volunteer_verify validation_failed volunteerId=%s actor=%s code=%s", utils.SafeLogValue(r.PathValue("id")), utils.SafeLogValue(ctx.ActorUserID), code)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	volunteer, code, message := s.store.VerifyVolunteer(r.PathValue("id"), normalized, ctx, s.now())
	if code != "" {
		// #nosec G706 -- volunteer id and actor are sanitized with utils.SafeLogValue.
		log.Printf("WARN incident-service volunteer_verify failed volunteerId=%s actor=%s code=%s", utils.SafeLogValue(r.PathValue("id")), utils.SafeLogValue(ctx.ActorUserID), code)
		utils.WriteError(w, statusForCode(code), code, message)
		return
	}
	// #nosec G706 -- decision and actor are sanitized with utils.SafeLogValue.
	log.Printf("INFO incident-service volunteer_verify completed volunteerId=%s decision=%s actor=%s", utils.SafeLogValue(volunteer.ID), utils.SafeLogValue(normalized.Decision), utils.SafeLogValue(ctx.ActorUserID))
	utils.WriteJSON(w, http.StatusOK, models.VolunteerProfileResponse{Volunteer: volunteer})
}

func (s *server) listVolunteerTasksHandler(w http.ResponseWriter, r *http.Request) {
	volunteerID := strings.TrimSpace(r.PathValue("id"))
	citizenUserID := ""
	if volunteer, found := s.store.VolunteerByID(volunteerID); found {
		citizenUserID = volunteer.CitizenUserID
	}
	if _, ok := s.requireVolunteerActor(w, r, citizenUserID); !ok {
		return
	}

	tasks, code, message := s.store.ListVolunteerTasks(volunteerID)
	if code != "" {
		// #nosec G706 -- logged values are sanitized with utils.SafeLogValue.
		log.Printf("WARN incident-service volunteer_tasks failed volunteerId=%s code=%s", utils.SafeLogValue(volunteerID), utils.SafeLogValue(code))
		utils.WriteError(w, statusForCode(code), code, message)
		return
	}
	// #nosec G706 -- logged values are sanitized with utils.SafeLogValue.
	log.Printf("INFO incident-service volunteer_tasks listed volunteerId=%s count=%d", utils.SafeLogValue(volunteerID), len(tasks))
	utils.WriteJSON(w, http.StatusOK, models.VolunteerTaskListResponse{Tasks: tasks})
}

func (s *server) assignVolunteerTaskHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r, assignmentRoles)
	if !ok {
		return
	}

	var request models.VolunteerTaskRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		// #nosec G706 -- incident id and actor are sanitized with utils.SafeLogValue.
		log.Printf("WARN incident-service volunteer_task_assign invalid_json incidentId=%s actor=%s error=%v", utils.SafeLogValue(r.PathValue("id")), utils.SafeLogValue(ctx.ActorUserID), err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	// #nosec G706 -- incident id, volunteer id, type, and actor are sanitized with utils.SafeLogValue.
	log.Printf("INFO incident-service volunteer_task_assign received incidentId=%s volunteerId=%s type=%s actor=%s", utils.SafeLogValue(r.PathValue("id")), utils.SafeLogValue(request.VolunteerID), utils.SafeLogValue(request.Type), utils.SafeLogValue(ctx.ActorUserID))

	normalized, code, message := normalizeVolunteerTaskRequest(request)
	if code != "" {
		// #nosec G706 -- incident id and volunteer id are sanitized with utils.SafeLogValue.
		log.Printf("WARN incident-service volunteer_task_assign validation_failed incidentId=%s volunteerId=%s code=%s", utils.SafeLogValue(r.PathValue("id")), utils.SafeLogValue(request.VolunteerID), code)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	task, code, message := s.store.AssignVolunteerTask(r.PathValue("id"), normalized, ctx, s.now())
	if code != "" {
		level := "WARN"
		if code == "store_error" {
			level = "ERROR"
		}
		// #nosec G706 -- incident id and volunteer id are sanitized with utils.SafeLogValue.
		log.Printf("%s incident-service volunteer_task_assign failed incidentId=%s volunteerId=%s code=%s", level, utils.SafeLogValue(r.PathValue("id")), utils.SafeLogValue(normalized.VolunteerID), code)
		utils.WriteError(w, statusForCode(code), code, message)
		return
	}
	log.Printf("INFO incident-service volunteer_task_assign completed incidentId=%s taskId=%s volunteerId=%s status=%s", task.IncidentID, task.ID, task.VolunteerID, task.Status)
	utils.WriteJSON(w, http.StatusCreated, task)
}

func (s *server) updateVolunteerTaskStatusHandler(w http.ResponseWriter, r *http.Request) {
	taskID := strings.TrimSpace(r.PathValue("id"))
	if _, ok := s.requireVolunteerActor(w, r, s.volunteerOwnerForTask(taskID)); !ok {
		return
	}

	var request models.VolunteerTaskStatusRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		// #nosec G706 -- logged values are sanitized with utils.SafeLogValue.
		log.Printf("WARN incident-service volunteer_task_status invalid_json taskId=%s error=%v", utils.SafeLogValue(taskID), err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	// #nosec G706 -- logged values are sanitized with utils.SafeLogValue.
	log.Printf("INFO incident-service volunteer_task_status received taskId=%s volunteerId=%s status=%s", utils.SafeLogValue(taskID), utils.SafeLogValue(request.VolunteerID), utils.SafeLogValue(request.Status))

	normalized, code, message := normalizeVolunteerTaskStatusRequest(request)
	if code != "" {
		// #nosec G706 -- logged values are sanitized with utils.SafeLogValue.
		log.Printf("WARN incident-service volunteer_task_status validation_failed taskId=%s volunteerId=%s code=%s", utils.SafeLogValue(taskID), utils.SafeLogValue(request.VolunteerID), utils.SafeLogValue(code))
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	task, code, message := s.store.UpdateVolunteerTaskStatus(taskID, normalized, s.now())
	if code != "" {
		// #nosec G706 -- logged values are sanitized with utils.SafeLogValue.
		log.Printf("WARN incident-service volunteer_task_status failed taskId=%s volunteerId=%s code=%s", utils.SafeLogValue(taskID), utils.SafeLogValue(normalized.VolunteerID), utils.SafeLogValue(code))
		utils.WriteError(w, statusForCode(code), code, message)
		return
	}
	// #nosec G706 -- logged values are sanitized with utils.SafeLogValue.
	log.Printf("INFO incident-service volunteer_task_status completed taskId=%s volunteerId=%s status=%s escalation=%t", utils.SafeLogValue(task.ID), utils.SafeLogValue(task.VolunteerID), utils.SafeLogValue(task.Status), task.EscalationRequired)
	utils.WriteJSON(w, http.StatusOK, task)
}

func (s *server) submitVolunteerObservationHandler(w http.ResponseWriter, r *http.Request) {
	taskID := strings.TrimSpace(r.PathValue("id"))
	if _, ok := s.requireVolunteerActor(w, r, s.volunteerOwnerForTask(taskID)); !ok {
		return
	}

	var request models.VolunteerObservationRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		// #nosec G706 -- logged values are sanitized with utils.SafeLogValue.
		log.Printf("WARN incident-service volunteer_observation invalid_json taskId=%s error=%v", utils.SafeLogValue(taskID), err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	// #nosec G706 -- logged values are sanitized with utils.SafeLogValue.
	log.Printf("INFO incident-service volunteer_observation received taskId=%s volunteerId=%s safetyStatus=%s escalationRequested=%t", utils.SafeLogValue(taskID), utils.SafeLogValue(request.VolunteerID), utils.SafeLogValue(request.SafetyStatus), request.EscalationRequested)

	normalized, code, message := normalizeVolunteerObservationRequest(request)
	if code != "" {
		// #nosec G706 -- logged values are sanitized with utils.SafeLogValue.
		log.Printf("WARN incident-service volunteer_observation validation_failed taskId=%s volunteerId=%s code=%s", utils.SafeLogValue(taskID), utils.SafeLogValue(request.VolunteerID), utils.SafeLogValue(code))
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	switch err := s.store.ValidateMediaReferences(normalized.Media); {
	case errors.Is(err, store.ErrUnknownMedia):
		utils.WriteError(w, http.StatusBadRequest, "unknown_media", "media references must be created through the upload initiation endpoint before reporting")
		return
	case errors.Is(err, store.ErrMediaAlreadyLinked):
		utils.WriteError(w, http.StatusBadRequest, "media_already_linked", "one or more media references are already linked to another incident")
		return
	case err != nil:
		utils.WriteError(w, http.StatusBadRequest, "invalid_media", err.Error())
		return
	}

	task, code, message := s.store.AddVolunteerObservation(taskID, normalized, s.now())
	if code != "" {
		// #nosec G706 -- logged values are sanitized with utils.SafeLogValue.
		log.Printf("WARN incident-service volunteer_observation failed taskId=%s volunteerId=%s code=%s", utils.SafeLogValue(taskID), utils.SafeLogValue(normalized.VolunteerID), utils.SafeLogValue(code))
		utils.WriteError(w, statusForCode(code), code, message)
		return
	}
	// #nosec G706 -- logged values are sanitized with utils.SafeLogValue.
	log.Printf("INFO incident-service volunteer_observation completed taskId=%s volunteerId=%s escalation=%t updateCount=%d", utils.SafeLogValue(task.ID), utils.SafeLogValue(task.VolunteerID), task.EscalationRequired, len(task.Updates))
	utils.WriteJSON(w, http.StatusOK, task)
}

// volunteerOwnerForTask resolves the registered citizen user id of the
// volunteer owning a task, used to authorize volunteer self-access.
func (s *server) volunteerOwnerForTask(taskID string) string {
	task, found := s.store.VolunteerTaskByID(taskID)
	if !found {
		return ""
	}
	volunteer, ok := s.store.VolunteerByID(task.VolunteerID)
	if !ok {
		return ""
	}
	return volunteer.CitizenUserID
}

func normalizeVolunteerRegistrationRequest(request models.RegisterVolunteerRequest) (models.RegisterVolunteerRequest, string, string) {
	request.CitizenUserID = strings.TrimSpace(request.CitizenUserID)
	request.Name = strings.TrimSpace(request.Name)
	request.Phone = strings.TrimSpace(request.Phone)
	request.Region = strings.TrimSpace(request.Region)
	request.District = strings.TrimSpace(request.District)
	request.Community = strings.TrimSpace(request.Community)
	request.AvailabilityStatus = strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(strings.ToLower(request.AvailabilityStatus)), "-", "_"), " ", "_")
	request.Skills = normalizeSafeList(request.Skills, 8, 64)
	request.Languages = normalizeSafeList(request.Languages, 6, 32)

	if request.CitizenUserID == "" || !utils.MediaRefPattern.MatchString(request.CitizenUserID) {
		return request, "invalid_citizen_user_id", "citizenUserId is required and must be a safe user reference"
	}
	if len(request.Name) < 2 || len(request.Name) > 120 || utils.UnsafeText(request.Name) {
		return request, "invalid_volunteer_name", "name must be 2 to 120 safe characters"
	}
	if request.Phone == "" || len(request.Phone) > 32 || utils.UnsafeText(request.Phone) {
		return request, "invalid_volunteer_phone", "phone is required and must be 32 safe characters or fewer"
	}
	if len(request.Region) < 2 || len(request.Region) > 80 || utils.UnsafeText(request.Region) {
		return request, "invalid_region", "region must be 2 to 80 safe characters"
	}
	if len(request.District) < 2 || len(request.District) > 100 || utils.UnsafeText(request.District) {
		return request, "invalid_district", "district must be 2 to 100 safe characters"
	}
	if len(request.Community) < 2 || len(request.Community) > 100 || utils.UnsafeText(request.Community) {
		return request, "invalid_community", "community must be 2 to 100 safe characters"
	}
	if request.AvailabilityStatus == "" {
		request.AvailabilityStatus = "available"
	}
	if !allowedVolunteerAvailability[request.AvailabilityStatus] {
		return request, "invalid_availability", "availabilityStatus must be available, busy, or off_duty"
	}
	if len(request.Skills) == 0 {
		return request, "missing_skills", "at least one volunteer skill is required"
	}
	if len(request.Languages) == 0 {
		request.Languages = []string{"en"}
	}
	return request, "", ""
}

func normalizeVolunteerVerifyRequest(request models.VerifyVolunteerRequest) (models.VerifyVolunteerRequest, string, string) {
	request.Decision = strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(strings.ToLower(request.Decision)), "-", "_"), " ", "_")
	request.Note = strings.TrimSpace(request.Note)
	if !allowedVolunteerVerificationDecisions[request.Decision] {
		return request, "invalid_decision", "decision must be verify, reject, or suspend"
	}
	if len(request.Note) < 5 || len(request.Note) > 1000 || utils.UnsafeText(request.Note) {
		return request, "invalid_note", "note must be 5 to 1000 safe characters"
	}
	return request, "", ""
}

func normalizeVolunteerTaskRequest(request models.VolunteerTaskRequest) (models.VolunteerTaskRequest, string, string) {
	request.VolunteerID = strings.TrimSpace(request.VolunteerID)
	request.Type = strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(strings.ToLower(request.Type)), "-", "_"), " ", "_")
	request.Priority = strings.TrimSpace(strings.ToLower(request.Priority))
	request.Instructions = strings.TrimSpace(request.Instructions)
	request.LocationLabel = strings.TrimSpace(request.LocationLabel)
	if request.VolunteerID == "" || !utils.MediaRefPattern.MatchString(request.VolunteerID) {
		return request, "invalid_volunteer_id", "volunteerId is required and must be a safe volunteer reference"
	}
	if !allowedVolunteerTaskTypes[request.Type] {
		return request, "invalid_task_type", "type must be welfare_check, shelter_support, supply_distribution, damage_observation, route_observation, or community_alerting"
	}
	if request.Priority == "" {
		request.Priority = "normal"
	}
	if !allowedAssignmentPriorities[request.Priority] {
		return request, "invalid_priority", "priority must be low, normal, high, or urgent"
	}
	if len(request.Instructions) < 10 || len(request.Instructions) > 1200 || utils.UnsafeText(request.Instructions) {
		return request, "invalid_instructions", "instructions must be 10 to 1200 safe characters"
	}
	if unsafeVolunteerInstructions(request.Instructions) {
		return request, "unsafe_volunteer_instructions", "volunteer tasks must not instruct civilians to enter floodwater, fight fires, conduct rescues, or approach violent/structural hazards"
	}
	if len(request.LocationLabel) < 2 || len(request.LocationLabel) > 180 || utils.UnsafeText(request.LocationLabel) {
		return request, "invalid_location_label", "locationLabel must be 2 to 180 safe characters"
	}
	return request, "", ""
}

func normalizeVolunteerTaskStatusRequest(request models.VolunteerTaskStatusRequest) (models.VolunteerTaskStatusRequest, string, string) {
	request.VolunteerID = strings.TrimSpace(request.VolunteerID)
	request.Status = strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(strings.ToLower(request.Status)), "-", "_"), " ", "_")
	request.Note = strings.TrimSpace(request.Note)
	request.SafetyStatus = strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(strings.ToLower(request.SafetyStatus)), "-", "_"), " ", "_")
	if request.VolunteerID == "" || !utils.MediaRefPattern.MatchString(request.VolunteerID) {
		return request, "invalid_volunteer_id", "volunteerId is required and must be a safe volunteer reference"
	}
	if !allowedVolunteerTaskStatuses[request.Status] {
		return request, "invalid_task_status", "status must be accepted, en_route, on_scene, completed, cancelled, or needs_escalation"
	}
	if len(request.Note) > 1000 || utils.UnsafeText(request.Note) {
		return request, "invalid_note", "note must be 1000 safe characters or fewer"
	}
	if request.SafetyStatus == "" {
		request.SafetyStatus = "safe"
	}
	if !allowedVolunteerSafetyStatuses[request.SafetyStatus] {
		return request, "invalid_safety_status", "safetyStatus must be safe, caution, unsafe, or needs_authority"
	}
	if request.Location != nil && !utils.ValidCoordinates(*request.Location) {
		return request, "invalid_location", "location must contain valid lat and lng values"
	}
	return request, "", ""
}

func normalizeVolunteerObservationRequest(request models.VolunteerObservationRequest) (models.VolunteerObservationRequest, string, string) {
	request.VolunteerID = strings.TrimSpace(request.VolunteerID)
	request.Observation = strings.TrimSpace(request.Observation)
	request.SafetyStatus = strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(strings.ToLower(request.SafetyStatus)), "-", "_"), " ", "_")
	request.Media = normalizeSafeList(request.Media, 8, 128)
	if request.VolunteerID == "" || !utils.MediaRefPattern.MatchString(request.VolunteerID) {
		return request, "invalid_volunteer_id", "volunteerId is required and must be a safe volunteer reference"
	}
	if len(request.Observation) < 5 || len(request.Observation) > 1500 || utils.UnsafeText(request.Observation) {
		return request, "invalid_observation", "observation must be 5 to 1500 safe characters"
	}
	if request.SafetyStatus == "" {
		request.SafetyStatus = "safe"
	}
	if !allowedVolunteerSafetyStatuses[request.SafetyStatus] {
		return request, "invalid_safety_status", "safetyStatus must be safe, caution, unsafe, or needs_authority"
	}
	if request.Location != nil && !utils.ValidCoordinates(*request.Location) {
		return request, "invalid_location", "location must contain valid lat and lng values"
	}
	return request, "", ""
}

func unsafeVolunteerInstructions(value string) bool {
	lower := strings.ToLower(value)
	unsafePhrases := []string{
		"enter floodwater",
		"walk through flood",
		"wade through",
		"fight fire",
		"put out fire",
		"rescue trapped",
		"enter collapsed",
		"go inside collapsed",
		"approach armed",
		"handle violent",
		"direct traffic on highway",
	}
	for _, phrase := range unsafePhrases {
		if strings.Contains(lower, phrase) {
			return true
		}
	}
	return false
}

func normalizeSafeList(values []string, limit int, maxLen int) []string {
	seen := map[string]bool{}
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || len(value) > maxLen || utils.UnsafeText(value) || seen[strings.ToLower(value)] {
			continue
		}
		seen[strings.ToLower(value)] = true
		normalized = append(normalized, value)
		if len(normalized) >= limit {
			break
		}
	}
	return normalized
}
