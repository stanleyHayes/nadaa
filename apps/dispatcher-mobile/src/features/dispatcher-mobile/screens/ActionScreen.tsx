import { Text, View } from "react-native";
import {
  ActionButton,
  Card,
  Field,
  ScreenHeading,
  SelectField,
  uiStyles,
} from "../../../ui/components";
import { mobileTheme } from "../../../app/theme";
import {
  assignmentAgencyOptions,
  incidentTransitionOptions,
  statusLabel,
} from "../data";
import type { DispatcherScreenProps } from "./types";

export function ActionScreen({ actions, state }: DispatcherScreenProps) {
  const incident = state.selectedIncident;
  const allowedTransitions = incident
    ? incidentTransitionOptions[incident.status]
    : [];

  return (
    <View style={uiStyles.card_plain}>
      <ScreenHeading kicker="Take action" title="Update incident" />

      {!incident ? (
        <Card>
          <Text style={stylesBody}>
            Select an incident from the Queue tab to verify, change status,
            assign, or add timeline notes.
          </Text>
        </Card>
      ) : (
        <>
          <Card>
            <Text style={stylesSectionTitle}>{incident.reference}</Text>
            <Text style={stylesBody}>{incident.description}</Text>
          </Card>

          <Card>
            <Text style={stylesSectionTitle}>Status update</Text>
            <SelectField
              label="New status"
              onChange={(value) => actions.setStatusForTransition(value)}
              options={allowedTransitions.map((status) => ({
                label: statusLabel(status),
                value: status,
              }))}
              value={state.statusForm.status}
            />
            <Field
              label="Note"
              multiline
              onChangeText={(value) =>
                actions.updateStatusForm({ note: value })
              }
              placeholder="Reason for status change"
              value={state.statusForm.note}
            />
            {state.statusForm.status === "closed" ||
            state.statusForm.status === "false_report" ? (
              <Field
                label="Resolution notes (required)"
                multiline
                onChangeText={(value) =>
                  actions.updateStatusForm({ resolutionNotes: value })
                }
                placeholder="Explain how the incident was resolved"
                value={state.statusForm.resolutionNotes}
              />
            ) : null}
            <ActionButton
              disabled={
                state.statusForm.status === incident.status ||
                ((state.statusForm.status === "closed" ||
                  state.statusForm.status === "false_report") &&
                  !state.statusForm.resolutionNotes.trim())
              }
              icon="check-circle"
              label="Update status"
              onPress={actions.submitStatusUpdate}
              tone="navy"
            />
          </Card>

          <Card>
            <Text style={stylesSectionTitle}>Assignment handoff</Text>
            <SelectField
              label="Agency"
              onChange={(value) => actions.chooseAssignmentAgency(value)}
              options={assignmentAgencyOptions.map((agency) => ({
                label: agency.name,
                value: agency.id,
              }))}
              value={state.assignmentForm.agencyId}
            />
            <Field
              label="Instructions"
              multiline
              onChangeText={(value) =>
                actions.updateAssignmentForm({ instructions: value })
              }
              placeholder="Specific instructions for the responding agency"
              value={state.assignmentForm.instructions}
            />
            <SelectField
              label="Priority"
              onChange={(value) =>
                actions.updateAssignmentForm({ priority: value })
              }
              options={[
                { label: "Low", value: "low" },
                { label: "Normal", value: "normal" },
                { label: "High", value: "high" },
                { label: "Urgent", value: "urgent" },
              ]}
              value={state.assignmentForm.priority}
            />
            <ActionButton
              disabled={!state.assignmentForm.agencyId}
              icon="user-check"
              label="Assign incident"
              onPress={actions.submitAssignment}
              tone="green"
            />
          </Card>

          <Card>
            <Text style={stylesSectionTitle}>Timeline note</Text>
            <Field
              label="Note"
              multiline
              onChangeText={(value) =>
                actions.updateTimelineNoteForm({ note: value })
              }
              placeholder="Add an operational note to the timeline"
              value={state.timelineNoteForm.note}
            />
            <ActionButton
              icon="message-square"
              label="Add note"
              onPress={actions.submitTimelineNote}
              tone="plain"
            />
          </Card>
        </>
      )}
    </View>
  );
}

const stylesBody = {
  color: mobileTheme.colors.ink,
  fontFamily: mobileTheme.font.regular,
  fontSize: 15,
  lineHeight: 22,
};

const stylesSectionTitle = {
  color: mobileTheme.colors.navy,
  fontFamily: mobileTheme.font.bold,
  fontSize: 18,
};
