import { Text, View } from "react-native";
import {
  ActionButton,
  Card,
  ListItem,
  ScreenHeading,
  StatusPill,
  uiStyles,
} from "../../../ui/components";
import { formatDateTime, severityTone, statusLabel } from "../data";
import type { DispatcherScreenProps } from "./types";

export function DetailScreen({ actions, state }: DispatcherScreenProps) {
  const incident = state.selectedIncident;

  if (!incident) {
    return (
      <View style={uiStyles.card_plain}>
        <ScreenHeading kicker="Incident detail" title="No incident selected" />
        <Card>
          <Text style={stylesBody}>
            Select an incident from the Queue tab to review details, timeline,
            and duplicate candidates.
          </Text>
          <ActionButton
            icon="list"
            label="Go to queue"
            onPress={() => {}}
            tone="navy"
          />
        </Card>
      </View>
    );
  }

  return (
    <View style={uiStyles.card_plain}>
      <ScreenHeading kicker={incident.reference} title="Incident detail" />

      <Card>
        <View style={stylesRow}>
          <View style={stylesGrow}>
            <Text style={stylesIncidentReference}>{incident.reference}</Text>
            <Text style={stylesMuted}>
              {incident.location.lat.toFixed(4)},{" "}
              {incident.location.lng.toFixed(4)}
            </Text>
          </View>
          <StatusPill
            label={incident.severity}
            tone={severityTone(incident.severity)}
          />
        </View>
        <Text style={stylesBody}>{incident.description}</Text>
        <View style={stylesRow}>
          <StatusPill label={statusLabel(incident.status)} tone="navy" />
          {incident.priorityReview ? (
            <StatusPill label="Priority review" tone="danger" />
          ) : null}
          {incident.anonymous ? (
            <StatusPill label="Anonymous" tone="gold" />
          ) : null}
        </View>
      </Card>

      <Card>
        <Text style={stylesSectionTitle}>Facts</Text>
        <View style={stylesFactGrid}>
          <Text style={stylesBody}>
            People affected: {incident.peopleAffected}
          </Text>
          <Text style={stylesBody}>
            Injuries: {incident.injuriesReported ? "Yes" : "No"}
          </Text>
          <Text style={stylesBody}>Urgency: {incident.urgency}</Text>
          <Text style={stylesBody}>
            Contact permission: {incident.contactPermission ? "Yes" : "No"}
          </Text>
          {incident.accessibilityNeeds ? (
            <Text style={stylesBody}>
              Accessibility: {incident.accessibilityNeeds}
            </Text>
          ) : null}
        </View>
      </Card>

      {incident.assignments.length > 0 ? (
        <Card>
          <Text style={stylesSectionTitle}>Assignments</Text>
          {incident.assignments.map((assignment) => (
            <ListItem key={assignment.id}>
              <Text style={stylesBody}>{assignment.agencyName}</Text>
              <Text style={stylesMuted}>
                Priority: {assignment.priority} · Lead:{" "}
                {assignment.responderLead}
              </Text>
              <Text style={stylesMuted}>{assignment.instructions}</Text>
            </ListItem>
          ))}
        </Card>
      ) : null}

      {incident.duplicateCandidates.length > 0 ? (
        <Card>
          <Text style={stylesSectionTitle}>Duplicate candidates</Text>
          {incident.duplicateCandidates.map((candidate) => (
            <ListItem key={candidate.incidentId}>
              <Text style={stylesBody}>{candidate.reference}</Text>
              <Text style={stylesMuted}>
                Score {Math.round(candidate.score * 100)}% ·{" "}
                {candidate.distanceMeters}m · {candidate.minutesApart}m apart
              </Text>
            </ListItem>
          ))}
        </Card>
      ) : null}

      <Card>
        <Text style={stylesSectionTitle}>Timeline</Text>
        {incident.timeline.length === 0 ? (
          <Text style={stylesBody}>No timeline events yet.</Text>
        ) : (
          incident.timeline.map((event) => (
            <ListItem key={event.id}>
              <Text style={stylesBody}>{event.message}</Text>
              <Text style={stylesMuted}>{formatDateTime(event.createdAt)}</Text>
            </ListItem>
          ))
        )}
      </Card>
    </View>
  );
}

const stylesBody = {
  color: "#101828",
  fontFamily: "Outfit_400Regular",
  fontSize: 15,
  lineHeight: 22,
};

const stylesFactGrid = {
  gap: 6,
};

const stylesGrow = {
  flex: 1,
};

const stylesIncidentReference = {
  color: "#0D1B3D",
  fontFamily: "Outfit_800ExtraBold",
  fontSize: 18,
};

const stylesMuted = {
  color: "#555B66",
  fontFamily: "Outfit_400Regular",
  fontSize: 13,
};

const stylesRow = {
  alignItems: "center",
  flexDirection: "row",
  gap: 12,
};

const stylesSectionTitle = {
  color: "#0D1B3D",
  fontFamily: "Outfit_800ExtraBold",
  fontSize: 18,
};
