package store

import (
	"time"

	"github.com/stanleyHayes/nadaa/services/guide-service/internal/models"
)

// seedGuides returns the fixture guide catalog for the given timestamp.
func seedGuides(now time.Time) []models.EmergencyGuide {
	return []models.EmergencyGuide{
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

// newGuide creates an EmergencyGuide fixture with the provided fields.
func newGuide(id, hazard, stage, title, body, language string, offline bool, sortOrder int, now time.Time) models.EmergencyGuide {
	return models.EmergencyGuide{
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
