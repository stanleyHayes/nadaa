package handlers

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/stanleyHayes/nadaa/services/school-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/school-service/internal/utils"
)

var allowedDrillTypes = map[string]bool{
	"fire":       true,
	"flood":      true,
	"storm":      true,
	"earthquake": true,
	"lockdown":   true,
	"evacuation": true,
	"medical":    true,
}

var allowedReadinessStatuses = map[string]bool{
	"ready":             true,
	"needs_improvement": true,
	"not_ready":         true,
	"not_assessed":      true,
}

var allowedRiskLevels = map[string]bool{
	"low":       true,
	"moderate":  true,
	"high":      true,
	"severe":    true,
	"emergency": true,
}

// maxRecordedDateFutureSkew is the small clock-skew tolerance applied when
// rejecting future drill/readiness dates, so a typo'd date cannot dominate
// the "latest" readiness and last-drill computations.
const maxRecordedDateFutureSkew = 24 * time.Hour

func (s *Server) listSchoolsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r, schoolReadRoles)
	if !ok {
		return
	}
	filter := models.SchoolFilter{
		District: strings.TrimSpace(strings.ToLower(r.URL.Query().Get("district"))),
		Query:    strings.TrimSpace(strings.ToLower(r.URL.Query().Get("q"))),
	}
	if len(filter.Query) > 120 || utils.UnsafeText(filter.Query) {
		utils.WriteError(w, http.StatusBadRequest, "invalid_query", "q must be 120 safe characters or fewer")
		return
	}
	if len(filter.District) > 100 || utils.UnsafeText(filter.District) {
		utils.WriteError(w, http.StatusBadRequest, "invalid_district", "district filter must be 100 safe characters or fewer")
		return
	}

	schools := s.store.ListSchools(filter, scopedDistrict(ctx), isSystemAdmin(ctx))
	// #nosec G706 -- filter and actor values are sanitized with utils.SafeLogValue.
	log.Printf("INFO school-service list_schools count=%d district=%s query=%s actor=%s role=%s", len(schools), utils.SafeLogValue(filter.District), utils.SafeLogValue(filter.Query), utils.SafeLogValue(ctx.ActorUserID), utils.SafeLogValue(ctx.ActorRole))
	utils.WriteJSON(w, http.StatusOK, models.SchoolListResponse{Schools: schools, GeneratedAt: s.now().UTC()})
}

func (s *Server) getSchoolHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r, schoolReadRoles)
	if !ok {
		return
	}
	school, found := s.store.GetSchool(r.PathValue("id"))
	if !found {
		utils.WriteError(w, http.StatusNotFound, "not_found", "school profile was not found")
		return
	}
	if !isSystemAdmin(ctx) && scopedDistrict(ctx) != "" && !strings.EqualFold(school.District, scopedDistrict(ctx)) {
		utils.WriteError(w, http.StatusForbidden, "district_scope_violation", "school is outside your district scope")
		return
	}
	// #nosec G706 -- actor values are sanitized with utils.SafeLogValue.
	log.Printf("INFO school-service get_school id=%s actor=%s role=%s", school.ID, utils.SafeLogValue(ctx.ActorUserID), utils.SafeLogValue(ctx.ActorRole))
	utils.WriteJSON(w, http.StatusOK, models.SchoolDetailResponse{School: school, GeneratedAt: s.now().UTC()})
}

func (s *Server) createSchoolHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r, schoolWriteRoles)
	if !ok {
		return
	}
	var request models.CreateSchoolRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		// #nosec G706 -- actor value is sanitized with utils.SafeLogValue.
		log.Printf("WARN school-service create_school invalid_json actor=%s error=%v", utils.SafeLogValue(ctx.ActorUserID), err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	request, code, message := normalizeCreateSchool(request)
	if code != "" {
		// #nosec G706 -- actor value is sanitized with utils.SafeLogValue.
		log.Printf("WARN school-service create_school validation_failed actor=%s code=%s", utils.SafeLogValue(ctx.ActorUserID), code)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}
	if !isSystemAdmin(ctx) && scopedDistrict(ctx) != "" && !strings.EqualFold(request.District, scopedDistrict(ctx)) {
		utils.WriteError(w, http.StatusForbidden, "district_scope_violation", "cannot create a school outside your district scope")
		return
	}
	school := s.store.CreateSchool(request, ctx, s.now().UTC())
	// #nosec G706 -- actor and district values are sanitized with utils.SafeLogValue.
	log.Printf("INFO school-service create_school id=%s actor=%s district=%s", school.ID, utils.SafeLogValue(ctx.ActorUserID), utils.SafeLogValue(school.District))
	utils.WriteJSON(w, http.StatusCreated, school)
}

func (s *Server) updateSchoolHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r, schoolWriteRoles)
	if !ok {
		return
	}
	var request models.UpdateSchoolRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		// #nosec G706 -- actor value is sanitized with utils.SafeLogValue.
		log.Printf("WARN school-service update_school invalid_json actor=%s error=%v", utils.SafeLogValue(ctx.ActorUserID), err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	request, code, message := normalizeUpdateSchool(request)
	if code != "" {
		// #nosec G706 -- actor value is sanitized with utils.SafeLogValue.
		log.Printf("WARN school-service update_school validation_failed actor=%s code=%s", utils.SafeLogValue(ctx.ActorUserID), code)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}
	existing, found := s.store.GetSchool(r.PathValue("id"))
	if !found {
		utils.WriteError(w, http.StatusNotFound, "not_found", "school profile was not found")
		return
	}
	if !isSystemAdmin(ctx) && scopedDistrict(ctx) != "" && !strings.EqualFold(existing.District, scopedDistrict(ctx)) {
		utils.WriteError(w, http.StatusForbidden, "district_scope_violation", "school is outside your district scope")
		return
	}
	if request.District != "" && !isSystemAdmin(ctx) && scopedDistrict(ctx) != "" && !strings.EqualFold(request.District, scopedDistrict(ctx)) {
		utils.WriteError(w, http.StatusForbidden, "district_scope_violation", "cannot move a school outside your district scope")
		return
	}
	school, code, message := s.store.UpdateSchool(r.PathValue("id"), request, ctx, s.now().UTC())
	if code != "" {
		// #nosec G706 -- path value and actor id are sanitized with utils.SafeLogValue.
		log.Printf("WARN school-service update_school failed id=%s actor=%s code=%s", utils.SafeLogValue(r.PathValue("id")), utils.SafeLogValue(ctx.ActorUserID), code)
		utils.WriteError(w, utils.StatusForCode(code), code, message)
		return
	}
	// #nosec G706 -- actor value is sanitized with utils.SafeLogValue.
	log.Printf("INFO school-service update_school id=%s actor=%s", school.ID, utils.SafeLogValue(ctx.ActorUserID))
	utils.WriteJSON(w, http.StatusOK, school)
}

func (s *Server) listDrillsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r, schoolReadRoles)
	if !ok {
		return
	}
	school, found := s.store.GetSchool(r.PathValue("id"))
	if !found {
		utils.WriteError(w, http.StatusNotFound, "not_found", "school profile was not found")
		return
	}
	if !isSystemAdmin(ctx) && scopedDistrict(ctx) != "" && !strings.EqualFold(school.District, scopedDistrict(ctx)) {
		utils.WriteError(w, http.StatusForbidden, "district_scope_violation", "school is outside your district scope")
		return
	}
	drills := s.store.ListDrills(r.PathValue("id"))
	// #nosec G706 -- actor value is sanitized with utils.SafeLogValue.
	log.Printf("INFO school-service list_drills schoolId=%s count=%d actor=%s", school.ID, len(drills), utils.SafeLogValue(ctx.ActorUserID))
	utils.WriteJSON(w, http.StatusOK, models.DrillListResponse{Drills: drills, GeneratedAt: s.now().UTC()})
}

func (s *Server) createDrillHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r, schoolWriteRoles)
	if !ok {
		return
	}
	school, found := s.store.GetSchool(r.PathValue("id"))
	if !found {
		utils.WriteError(w, http.StatusNotFound, "not_found", "school profile was not found")
		return
	}
	if !isSystemAdmin(ctx) && scopedDistrict(ctx) != "" && !strings.EqualFold(school.District, scopedDistrict(ctx)) {
		utils.WriteError(w, http.StatusForbidden, "district_scope_violation", "school is outside your district scope")
		return
	}
	var request models.CreateDrillRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		// #nosec G706 -- actor value is sanitized with utils.SafeLogValue.
		log.Printf("WARN school-service create_drill invalid_json actor=%s error=%v", utils.SafeLogValue(ctx.ActorUserID), err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	request, code, message := normalizeCreateDrill(request, s.now().UTC())
	if code != "" {
		// #nosec G706 -- actor value is sanitized with utils.SafeLogValue.
		log.Printf("WARN school-service create_drill validation_failed actor=%s code=%s", utils.SafeLogValue(ctx.ActorUserID), code)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}
	drill, code, message := s.store.CreateDrill(r.PathValue("id"), request, ctx, s.now().UTC())
	if code != "" {
		// #nosec G706 -- path value and actor id are sanitized with utils.SafeLogValue.
		log.Printf("WARN school-service create_drill failed schoolId=%s actor=%s code=%s", utils.SafeLogValue(r.PathValue("id")), utils.SafeLogValue(ctx.ActorUserID), code)
		utils.WriteError(w, utils.StatusForCode(code), code, message)
		return
	}
	// #nosec G706 -- actor value is sanitized with utils.SafeLogValue.
	log.Printf("INFO school-service create_drill id=%s schoolId=%s actor=%s", drill.ID, drill.SchoolID, utils.SafeLogValue(ctx.ActorUserID))
	utils.WriteJSON(w, http.StatusCreated, drill)
}

func (s *Server) getReadinessHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r, schoolReadRoles)
	if !ok {
		return
	}
	school, found := s.store.GetSchool(r.PathValue("id"))
	if !found {
		utils.WriteError(w, http.StatusNotFound, "not_found", "school profile was not found")
		return
	}
	if !isSystemAdmin(ctx) && scopedDistrict(ctx) != "" && !strings.EqualFold(school.District, scopedDistrict(ctx)) {
		utils.WriteError(w, http.StatusForbidden, "district_scope_violation", "school is outside your district scope")
		return
	}
	readiness, ok := s.store.GetLatestReadiness(r.PathValue("id"))
	// #nosec G706 -- actor value is sanitized with utils.SafeLogValue.
	log.Printf("INFO school-service get_readiness schoolId=%s found=%t actor=%s", school.ID, ok, utils.SafeLogValue(ctx.ActorUserID))
	utils.WriteJSON(w, http.StatusOK, models.ReadinessResponse{Readiness: readiness, GeneratedAt: s.now().UTC()})
}

func (s *Server) createReadinessHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r, schoolWriteRoles)
	if !ok {
		return
	}
	school, found := s.store.GetSchool(r.PathValue("id"))
	if !found {
		utils.WriteError(w, http.StatusNotFound, "not_found", "school profile was not found")
		return
	}
	if !isSystemAdmin(ctx) && scopedDistrict(ctx) != "" && !strings.EqualFold(school.District, scopedDistrict(ctx)) {
		utils.WriteError(w, http.StatusForbidden, "district_scope_violation", "school is outside your district scope")
		return
	}
	var request models.CreateReadinessRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		// #nosec G706 -- actor value is sanitized with utils.SafeLogValue.
		log.Printf("WARN school-service create_readiness invalid_json actor=%s error=%v", utils.SafeLogValue(ctx.ActorUserID), err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	request, code, message := normalizeCreateReadiness(request, s.now().UTC())
	if code != "" {
		// #nosec G706 -- actor value is sanitized with utils.SafeLogValue.
		log.Printf("WARN school-service create_readiness validation_failed actor=%s code=%s", utils.SafeLogValue(ctx.ActorUserID), code)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}
	check, code, message := s.store.CreateReadinessCheck(r.PathValue("id"), request, ctx, s.now().UTC())
	if code != "" {
		// #nosec G706 -- path value and actor id are sanitized with utils.SafeLogValue.
		log.Printf("WARN school-service create_readiness failed schoolId=%s actor=%s code=%s", utils.SafeLogValue(r.PathValue("id")), utils.SafeLogValue(ctx.ActorUserID), code)
		utils.WriteError(w, utils.StatusForCode(code), code, message)
		return
	}
	// #nosec G706 -- actor value is sanitized with utils.SafeLogValue.
	log.Printf("INFO school-service create_readiness id=%s schoolId=%s actor=%s status=%s", check.ID, check.SchoolID, utils.SafeLogValue(ctx.ActorUserID), check.OverallStatus)
	utils.WriteJSON(w, http.StatusCreated, check)
}

func normalizeCreateSchool(request models.CreateSchoolRequest) (models.CreateSchoolRequest, string, string) {
	request.Name = strings.TrimSpace(request.Name)
	request.Region = strings.TrimSpace(request.Region)
	request.District = strings.TrimSpace(request.District)
	request.Address = strings.TrimSpace(request.Address)
	request.Hazards = utils.NormalizeStrings(request.Hazards)
	request.EmergencyContacts = utils.NormalizeContacts(request.EmergencyContacts)
	request.EvacuationPoints = utils.NormalizeEvacuationPoints(request.EvacuationPoints)

	if request.Name == "" || len(request.Name) > 200 || utils.UnsafeText(request.Name) {
		return request, "invalid_name", "name is required and must be 200 safe characters or fewer"
	}
	if !utils.ValidCoordinates(request.Location) {
		return request, "invalid_location", "location must be a valid WGS84 latitude and longitude"
	}
	if request.Region == "" || len(request.Region) > 100 || utils.UnsafeText(request.Region) {
		return request, "invalid_region", "region is required and must be 100 safe characters or fewer"
	}
	if request.District == "" || len(request.District) > 100 || utils.UnsafeText(request.District) {
		return request, "invalid_district", "district is required and must be 100 safe characters or fewer"
	}
	if request.StudentPopulation < 0 || request.StudentPopulation > 10000 {
		return request, "invalid_student_population", "studentPopulation must be between 0 and 10000"
	}
	if code, message := validateEvacuationPoints(request.EvacuationPoints); code != "" {
		return request, code, message
	}
	return request, "", ""
}

func normalizeUpdateSchool(request models.UpdateSchoolRequest) (models.UpdateSchoolRequest, string, string) {
	request.Name = strings.TrimSpace(request.Name)
	request.Region = strings.TrimSpace(request.Region)
	request.District = strings.TrimSpace(request.District)
	request.Address = strings.TrimSpace(request.Address)
	request.Hazards = utils.NormalizeStrings(request.Hazards)
	request.EmergencyContacts = utils.NormalizeContacts(request.EmergencyContacts)
	request.EvacuationPoints = utils.NormalizeEvacuationPoints(request.EvacuationPoints)

	if request.Name != "" && (len(request.Name) > 200 || utils.UnsafeText(request.Name)) {
		return request, "invalid_name", "name must be 200 safe characters or fewer"
	}
	if request.Location != nil && !utils.ValidCoordinates(*request.Location) {
		return request, "invalid_location", "location must be a valid WGS84 latitude and longitude"
	}
	if request.Region != "" && (len(request.Region) > 100 || utils.UnsafeText(request.Region)) {
		return request, "invalid_region", "region must be 100 safe characters or fewer"
	}
	if request.District != "" && (len(request.District) > 100 || utils.UnsafeText(request.District)) {
		return request, "invalid_district", "district must be 100 safe characters or fewer"
	}
	if request.StudentPopulation != nil && (*request.StudentPopulation < 0 || *request.StudentPopulation > 10000) {
		return request, "invalid_student_population", "studentPopulation must be between 0 and 10000"
	}
	if code, message := validateEvacuationPoints(request.EvacuationPoints); code != "" {
		return request, code, message
	}
	return request, "", ""
}

// validateEvacuationPoints rejects evacuation points with out-of-range
// coordinates or a negative capacity instead of persisting and serving them.
func validateEvacuationPoints(points []models.EvacuationPoint) (string, string) {
	for _, point := range points {
		if !utils.ValidCoordinates(point.Location) || point.Capacity < 0 {
			return "invalid_evacuation_point", "evacuationPoints must have valid WGS84 coordinates and non-negative capacity"
		}
	}
	return "", ""
}

func normalizeCreateDrill(request models.CreateDrillRequest, now time.Time) (models.CreateDrillRequest, string, string) {
	request.Type = utils.NormalizeString(request.Type)
	request.Notes = strings.TrimSpace(request.Notes)
	if request.Type == "" || !allowedDrillTypes[request.Type] {
		return request, "invalid_type", "type must be fire, flood, storm, earthquake, lockdown, evacuation, or medical"
	}
	if request.Date.IsZero() {
		return request, "invalid_date", "date is required"
	}
	if request.Date.After(now.Add(maxRecordedDateFutureSkew)) {
		return request, "invalid_date", "date must not be more than 24 hours in the future"
	}
	if request.Participants < 0 || request.Participants > 50000 {
		return request, "invalid_participants", "participants must be between 0 and 50000"
	}
	if len(request.Notes) > 1000 || utils.UnsafeText(request.Notes) {
		return request, "invalid_notes", "notes must be 1000 safe characters or fewer"
	}
	return request, "", ""
}

func normalizeCreateReadiness(request models.CreateReadinessRequest, now time.Time) (models.CreateReadinessRequest, string, string) {
	request.RiskLevel = utils.NormalizeString(request.RiskLevel)
	request.OverallStatus = utils.NormalizeString(request.OverallStatus)
	request.AreaRiskRef = strings.TrimSpace(request.AreaRiskRef)
	request.Notes = strings.TrimSpace(request.Notes)
	request.ChecklistItems = utils.NormalizeChecklistItems(request.ChecklistItems)

	if request.OverallStatus == "" || !allowedReadinessStatuses[request.OverallStatus] {
		return request, "invalid_overall_status", "overallStatus must be ready, needs_improvement, not_ready, or not_assessed"
	}
	if !allowedRiskLevels[request.RiskLevel] {
		return request, "invalid_risk_level", "riskLevel must be low, moderate, high, severe, or emergency"
	}
	if request.CheckDate.IsZero() {
		return request, "invalid_check_date", "checkDate is required"
	}
	if request.CheckDate.After(now.Add(maxRecordedDateFutureSkew)) {
		return request, "invalid_check_date", "checkDate must not be more than 24 hours in the future"
	}
	if len(request.Notes) > 1000 || utils.UnsafeText(request.Notes) {
		return request, "invalid_notes", "notes must be 1000 safe characters or fewer"
	}
	return request, "", ""
}
