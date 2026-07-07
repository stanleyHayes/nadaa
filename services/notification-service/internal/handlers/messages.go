package handlers

import (
	"fmt"

	"github.com/stanleyHayes/nadaa/services/notification-service/internal/models"
)

func languageMenu() string {
	return "Select language / Paw kasa:\n1 English\n2 Twi\n3 Ga\n4 Ewe\n5 Dagbani\n6 Hausa"
}

func mainMenu(language string) string {
	switch language {
	case "tw":
		return "NADAA menu:\n1 Kɔkɔbɔ\n2 Bɔ amanneɛ\n3 Dwanekɔbea\n4 Frɛ 112"
	case "ga":
		return "NADAA menu:\n1 Alerts\n2 Report emergency\n3 Shelter\n4 Call 112"
	default:
		return "NADAA menu:\n1 Current alerts\n2 Report emergency\n3 Find shelter\n4 112 guidance"
	}
}

func hazardMenu(language string) string {
	switch language {
	case "tw":
		return "Paw asiane:\n1 Nsuyiri\n2 Ogya\n3 Ayaresa\n4 Kar akwanhyia\n5 Foforo"
	default:
		return "Select emergency type:\n1 Flood\n2 Fire\n3 Medical\n4 Road crash\n5 Other"
	}
}

func urgencyMenu(language string) string {
	switch language {
	case "tw":
		return "Ɛyɛ den sɛn?\n1 Kakra\n2 Mfinimfini\n3 Den\n4 Ɛhaw nkwa"
	default:
		return "Select urgency:\n1 Low\n2 Moderate\n3 High\n4 Life-threatening"
	}
}

func localizedMessage(language string, key string) string {
	switch key {
	case "provider_error":
		if language == "tw" {
			return "Nkitahodie no nni hɔ seesei. Sɛ ɛyɛ asianeɛ a, frɛ 112."
		}
		return "The channel is temporarily unavailable. If this is life-threatening, call 112."
	default:
		return smsHelpMessage()
	}
}

func alertSummaryMessage(language string, alerts []models.CitizenAlert) string {
	if len(alerts) == 0 {
		if language == "tw" {
			return "Kɔkɔbɔ foforo biara nni hɔ seesei. Sɛ ɛyɛ asianeɛ a, frɛ 112."
		}
		return "No current NADAA alerts. If this is life-threatening, call 112."
	}
	alert := alerts[0]
	return fmt.Sprintf("%s: %s. %s", alert.Title, alert.TargetLabel, alert.RecommendedAction)
}

func smsAlertMessage(alerts []models.CitizenAlert) string {
	return alertSummaryMessage("en", alerts)
}

func shelterMessage(language string) string {
	if language == "tw" {
		return "Bɛn dwanekɔbea: Accra Sports Hall, Osu Community Centre. Fa wo ho kɔ baabi a ɛyɛ banbɔ. Frɛ 112 sɛ ɛyɛ asianeɛ."
	}
	return "Nearby shelters: Accra Sports Hall, Osu Community Centre. Move to safe high ground when directed. Call 112 for immediate danger."
}

func guidance112Message(language string) string {
	if language == "tw" {
		return "Sɛ nkwa wɔ asiane mu a, frɛ 112 ntɛm. Ka baabi a wowɔ, asiane no, ne nnipa dodow."
	}
	return "If life is in immediate danger, call 112 now. Share your location, emergency type, and people affected."
}

func reportConfirmationMessage(language string, report models.InclusiveAccessReport) string {
	reference := report.ID
	if report.IncidentReference != "" {
		reference = report.IncidentReference
	}
	if language == "tw" {
		return fmt.Sprintf("Yɛagye wo amanneɛ no: %s. Frɛ 112 sɛ nkwa wɔ asiane mu.", reference)
	}
	return fmt.Sprintf("NADAA report received: %s. Call 112 if life is in immediate danger.", reference)
}

func smsHelpMessage() string {
	return "NADAA SMS commands: ALERTS, SHELTER, HELP, or REPORT FLOOD HIGH your location/details. Call 112 for immediate danger."
}

func smsReportUsage() string {
	return "Use: REPORT FLOOD HIGH your location/details. Hazards: FLOOD FIRE MEDICAL ROAD OTHER. Urgency: LOW MODERATE HIGH LIFE."
}

func whatsappHelpMessage() string {
	return "NADAA WhatsApp commands: ALERTS, RISK, REPORT, SHELTER, GUIDE FLOOD, HELP, or 112. To report: REPORT FLOOD HIGH your location/details, or send REPORT and answer the prompts."
}

func whatsappReportUsage() string {
	return "Use: REPORT FLOOD HIGH your location/details. Hazards: FLOOD FIRE MEDICAL ROAD OTHER. Urgency: LOW MODERATE HIGH LIFE."
}

func whatsappHazardPrompt() string {
	return "What type of emergency are you reporting? Reply FLOOD, FIRE, MEDICAL, ROAD, SECURITY, STORM, or OTHER."
}

func whatsappUrgencyPrompt() string {
	return "How urgent is it? Reply LOW, MODERATE, HIGH, or LIFE. Call 112 now if life is in immediate danger."
}

func whatsappLocationPrompt() string {
	return "Please send your location pin, nearest landmark, or a short description. You can attach a photo or voice note if safe."
}

func riskCheckMessage(language string, hasLocation bool, alerts []models.CitizenAlert) string {
	prefix := "Share your location pin for a more specific risk check. "
	if hasLocation {
		prefix = "Location received for this WhatsApp risk check. "
	}
	if len(alerts) == 0 {
		return prefix + "No current NADAA alerts are active in the notification feed. Stay alert and call 112 for immediate danger."
	}
	alert := alerts[0]
	return fmt.Sprintf("%sCurrent NADAA signal: %s for %s. %s", prefix, alert.Title, alert.TargetLabel, alert.RecommendedAction)
}

func emergencyGuideMessage(language string, hazard string) string {
	switch hazard {
	case "fire":
		return "Fire guide: leave the area, keep exits clear, avoid smoke, do not use lifts, and call 112 if flames or heavy smoke are visible."
	case "medical_emergency":
		return "Medical guide: call 112 for serious injury, keep the person still, share location clearly, and do not move them unless the area is unsafe."
	case "road_crash":
		return "Road crash guide: move away from traffic if safe, call 112, warn approaching vehicles, and do not move injured people unless there is immediate danger."
	case "storm":
		return "Storm guide: stay indoors, avoid trees and power lines, secure loose items if safe, and follow NADAA alerts."
	default:
		if language == "tw" {
			return "Nsuyiri akwankyerɛ: kɔ baabi a ɛkorɔn, kwati nsuo a ɛsen, sie nkrataa, na frɛ 112 sɛ nkwa wɔ asiane mu."
		}
		return "Flood guide: move to higher ground, avoid drains and floodwater, keep documents dry, follow official alerts, and call 112 for immediate danger."
	}
}
