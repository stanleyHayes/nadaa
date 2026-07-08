package store

// VolunteerSafetyRules returns the default safety guidance for volunteer tasks.
func VolunteerSafetyRules() []string {
	return []string{
		"Stay in public, safe areas and never enter floodwater, fire zones, collapsed structures, or violent scenes.",
		"Call 112 and request authority escalation for injuries, trapped people, unsafe crowds, or blocked emergency access.",
		"Share observations, photos, and status updates only when doing so does not delay evacuation or personal safety.",
	}
}
