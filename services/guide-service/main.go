package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type server struct {
	store *memoryStore
}

type memoryStore struct {
	mu     sync.RWMutex
	guides []emergencyGuide
}

type emergencyGuide struct {
	ID               string    `json:"id"`
	HazardType       string    `json:"hazardType"`
	Stage            string    `json:"stage"`
	Title            string    `json:"title"`
	Body             string    `json:"body"`
	Language         string    `json:"language"`
	OfflineAvailable bool      `json:"offlineAvailable"`
	SortOrder        int       `json:"sortOrder"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type guideListResponse struct {
	Guides []emergencyGuide `json:"guides"`
}

type guideFilters struct {
	HazardType string
	Stage      string
	Language   string
	Offline    *bool
}

type apiError struct {
	Error apiErrorBody `json:"error"`
}

type apiErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

var allowedHazards = map[string]bool{
	"flood":             true,
	"fire":              true,
	"road_crash":        true,
	"building_collapse": true,
	"medical_emergency": true,
	"security_incident": true,
	"disease_outbreak":  true,
	"electrical_hazard": true,
	"blocked_drain":     true,
	"landslide":         true,
	"marine_accident":   true,
	"storm":             true,
	"tidal_wave":        true,
	"other":             true,
}

var allowedStages = map[string]bool{
	"before":   true,
	"during":   true,
	"after":    true,
	"recovery": true,
}

func main() {
	srv := &server{store: newMemoryStore()}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", srv.healthHandler)
	mux.HandleFunc("GET /api/v1/guides", srv.listGuidesHandler)

	addr := envOrDefault("NADAA_GUIDE_ADDR", ":8086")
	log.Printf("guide-service listening on %s", addr)
	if err := http.ListenAndServe(addr, withCORS(mux)); err != nil {
		log.Fatal(err)
	}
}

func newMemoryStore() *memoryStore {
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	return &memoryStore{guides: seedGuides(now)}
}

func (s *server) healthHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "guide-service"})
}

func (s *server) listGuidesHandler(w http.ResponseWriter, r *http.Request) {
	filters, code, message := parseGuideFilters(r)
	if code != "" {
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	writeJSON(w, http.StatusOK, guideListResponse{Guides: s.store.listGuides(filters)})
}

func parseGuideFilters(r *http.Request) (guideFilters, string, string) {
	query := r.URL.Query()
	filters := guideFilters{
		HazardType: normalizeQueryValue(query.Get("hazard")),
		Stage:      normalizeQueryValue(query.Get("stage")),
		Language:   normalizeLanguage(query.Get("language")),
	}

	if filters.HazardType != "" && !allowedHazards[filters.HazardType] {
		return guideFilters{}, "invalid_hazard", "hazard must be a supported NADAA hazard type"
	}
	if filters.Stage != "" && !allowedStages[filters.Stage] {
		return guideFilters{}, "invalid_stage", "stage must be before, during, after, or recovery"
	}
	if offlineRaw := normalizeQueryValue(query.Get("offline")); offlineRaw != "" {
		offline, err := strconv.ParseBool(offlineRaw)
		if err != nil {
			return guideFilters{}, "invalid_offline", "offline must be true or false"
		}
		filters.Offline = &offline
	}

	return filters, "", ""
}

func (m *memoryStore) listGuides(filters guideFilters) []emergencyGuide {
	m.mu.RLock()
	defer m.mu.RUnlock()

	guides := make([]emergencyGuide, 0, len(m.guides))
	for _, guide := range m.guides {
		if filters.HazardType != "" && guide.HazardType != filters.HazardType {
			continue
		}
		if filters.Stage != "" && guide.Stage != filters.Stage {
			continue
		}
		if filters.Language != "" && guide.Language != filters.Language {
			continue
		}
		if filters.Offline != nil && guide.OfflineAvailable != *filters.Offline {
			continue
		}
		guides = append(guides, guide)
	}

	if filters.Language != "" && filters.Language != "en" && len(guides) == 0 {
		fallbackFilters := filters
		fallbackFilters.Language = "en"
		return m.listGuidesWithoutLock(fallbackFilters)
	}

	sortGuides(guides)
	return guides
}

func (m *memoryStore) listGuidesWithoutLock(filters guideFilters) []emergencyGuide {
	guides := make([]emergencyGuide, 0, len(m.guides))
	for _, guide := range m.guides {
		if filters.HazardType != "" && guide.HazardType != filters.HazardType {
			continue
		}
		if filters.Stage != "" && guide.Stage != filters.Stage {
			continue
		}
		if filters.Language != "" && guide.Language != filters.Language {
			continue
		}
		if filters.Offline != nil && guide.OfflineAvailable != *filters.Offline {
			continue
		}
		guides = append(guides, guide)
	}
	sortGuides(guides)
	return guides
}

func sortGuides(guides []emergencyGuide) {
	sort.Slice(guides, func(i, j int) bool {
		if guides[i].SortOrder == guides[j].SortOrder {
			return guides[i].Title < guides[j].Title
		}
		return guides[i].SortOrder < guides[j].SortOrder
	})
}

func seedGuides(now time.Time) []emergencyGuide {
	return []emergencyGuide{
		newGuide("guide_flood_before_en", "flood", "before", "Prepare before flooding", "Know your nearest shelter, keep documents dry, clear drains safely, prepare drinking water, and agree on a family meeting point.", "en", true, 10, now),
		newGuide("guide_flood_during_en", "flood", "during", "Stay safe during flooding", "Move to higher ground, avoid walking or driving through floodwater, turn off electricity only if safe, and call 112 for life-threatening danger.", "en", true, 20, now),
		newGuide("guide_flood_after_en", "flood", "after", "Return safely after flooding", "Wait for official guidance, avoid contaminated water, photograph damage before cleanup, and report blocked drains or damaged utilities.", "en", true, 30, now),
		newGuide("guide_fire_during_en", "fire", "during", "Fire safety response", "Leave immediately, warn people nearby, stay low under smoke, never use lifts, and call 112 for Ghana National Fire Service support.", "en", true, 40, now),
		newGuide("guide_road_crash_during_en", "road_crash", "during", "Road crash first response", "Move to a safe place, switch on hazard lights if possible, do not move injured people unless there is immediate danger, and call 112.", "en", true, 50, now),
		newGuide("guide_electrical_during_en", "electrical_hazard", "during", "Electrical hazard safety", "Stay away from fallen wires, flooded electrical equipment, and sparking poles. Keep others clear and call 112 or the utility emergency line.", "en", true, 60, now),
		newGuide("guide_disease_before_en", "disease_outbreak", "before", "Disease prevention basics", "Wash hands often, isolate when symptomatic, follow Ghana Health Service guidance, keep medicine supplies ready, and protect vulnerable family members.", "en", true, 70, now),
		newGuide("guide_evacuation_during_en", "other", "during", "Safe evacuation", "Take only essentials, follow official routes, help children and elderly people first, avoid floodwater or smoke, and tell relatives where you are going.", "en", true, 80, now),
		newGuide("guide_bag_before_en", "other", "before", "Emergency bag checklist", "Pack water, food, torch, radio, power bank, first aid, medicine, copies of documents, cash, hygiene items, and child or disability-specific supplies.", "en", true, 90, now),
		newGuide("guide_family_before_en", "other", "before", "Family emergency plan", "Choose meeting points, store emergency contacts, teach children how to call 112, plan transport, and decide who checks on vulnerable relatives.", "en", true, 100, now),
		newGuide("guide_112_during_en", "other", "during", "Calling 112", "Call 112 for life-threatening emergencies. Share the hazard, exact location, people affected, injuries, and a safe callback number if available.", "en", true, 110, now),
		newGuide("guide_flood_before_tw", "flood", "before", "Siesie wo ho ansa na nsuyiri aba", "Hu baabi a wobɛkɔ akɔpɛ ahobammɔ, sie wo nkrataa wɔ baabi a nsuo renka, na siesie nneɛma a ehia wo abusua.", "tw", true, 120, now),
	}
}

func newGuide(id string, hazard string, stage string, title string, body string, language string, offline bool, sortOrder int, now time.Time) emergencyGuide {
	return emergencyGuide{
		ID:               id,
		HazardType:       hazard,
		Stage:            stage,
		Title:            title,
		Body:             body,
		Language:         language,
		OfflineAvailable: offline,
		SortOrder:        sortOrder,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

func decodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("write json response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, apiError{Error: apiErrorBody{Code: code, Message: message}})
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func normalizeQueryValue(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}

func normalizeLanguage(language string) string {
	language = normalizeQueryValue(language)
	if language == "" {
		return "en"
	}
	return language
}

func envOrDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
