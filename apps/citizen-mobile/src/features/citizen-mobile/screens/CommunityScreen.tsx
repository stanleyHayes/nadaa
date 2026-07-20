import { Text, View } from "react-native";
import type {
  VolunteerSafetyStatus,
  VolunteerTaskRecord,
  VolunteerTaskStatus,
} from "@nadaa/shared-types";
import { mobileTheme } from "../../../app/theme";
import {
  ActionButton,
  Card,
  EmptyState,
  Field,
  Metric,
  ScreenHeading,
  SegmentedControl,
  SeverityBadge,
  StatusPill,
  uiStyles,
} from "../../../ui/components";
import type { CitizenScreenProps } from "./types";

type VolunteerStatusAction = Exclude<VolunteerTaskStatus, "assigned">;

const statusActions: Array<{
  icon: string;
  label: string;
  status: VolunteerStatusAction;
  tone: "danger" | "green" | "navy" | "plain";
}> = [
  { icon: "check-circle", label: "Accept", status: "accepted", tone: "green" },
  { icon: "navigation", label: "En route", status: "en_route", tone: "navy" },
  { icon: "map-pin", label: "On scene", status: "on_scene", tone: "navy" },
  { icon: "flag", label: "Complete", status: "completed", tone: "plain" },
  {
    icon: "alert-triangle",
    label: "Escalate",
    status: "needs_escalation",
    tone: "danger",
  },
];

/**
 * The incident-service enforces the chain
 * assigned→accepted→(en_route→)on_scene→completed (needs_escalation allowed
 * from any active status) and 400s invalid_transition otherwise — only offer
 * the actions that are valid from the task's current status.
 */
const allowedNextStatuses: Record<string, VolunteerStatusAction[]> = {
  accepted: ["en_route", "on_scene", "needs_escalation"],
  assigned: ["accepted", "needs_escalation"],
  en_route: ["on_scene", "needs_escalation"],
  needs_escalation: ["accepted", "en_route", "on_scene", "completed"],
  on_scene: ["completed", "needs_escalation"],
};

const safetyOptions: Array<{
  label: string;
  value: VolunteerSafetyStatus;
}> = [
  { label: "Safe", value: "safe" },
  { label: "Caution", value: "caution" },
  { label: "Unsafe", value: "unsafe" },
  { label: "Authority", value: "needs_authority" },
];

export function CommunityScreen({ actions, state }: CitizenScreenProps) {
  const profile = state.volunteerProfile;
  const activeTask =
    state.volunteerTasks.find(
      (task) => !["completed", "cancelled"].includes(task.status),
    ) ?? state.volunteerTasks[0];

  return (
    <View style={uiStyles.card_plain}>
      <ScreenHeading
        kicker="Community response"
        title="Volunteer assignments"
      />

      {profile ? (
        <Card tone="green">
          <View style={stylesRow}>
            <View style={stylesGrow}>
              <Text style={stylesSectionTitle}>{profile.name}</Text>
              <Text style={stylesMuted}>
                {profile.community}, {profile.district}
              </Text>
            </View>
            <StatusPill
              label={profile.verificationStatus}
              tone={profile.verificationStatus === "verified" ? "green" : "gold"}
            />
          </View>
          <View style={stylesMetricRow}>
            <Metric label="Active tasks" value={state.activeVolunteerTaskCount} />
            <Metric label="Skills" value={profile.skills.length} />
          </View>
          <Text style={stylesBody}>
            {profile.skills.join(", ")} volunteer supporting {profile.groupId}.
          </Text>
          <View style={stylesButtonGrid}>
            <ActionButton
              icon="refresh-cw"
              label="Refresh assignments"
              onPress={actions.refreshVolunteerTasks}
              tone="navy"
            />
          </View>
        </Card>
      ) : (
        <Card tone="green">
          <Text style={stylesSectionTitle}>No volunteer profile yet</Text>
          <Text style={stylesBody}>
            Sign in on the Support tab, then register as a community volunteer
            to receive and update assignments.
          </Text>
          <Field
            label="Community"
            onChangeText={(community) =>
              actions.saveVolunteerRegistration({
                ...state.volunteerRegistration,
                community,
              })
            }
            placeholder="e.g. Jamestown"
            value={state.volunteerRegistration.community}
          />
          <Field
            label="District"
            onChangeText={(district) =>
              actions.saveVolunteerRegistration({
                ...state.volunteerRegistration,
                district,
              })
            }
            placeholder="e.g. Accra Metropolitan"
            value={state.volunteerRegistration.district}
          />
          <Field
            label="Region"
            onChangeText={(region) =>
              actions.saveVolunteerRegistration({
                ...state.volunteerRegistration,
                region,
              })
            }
            placeholder="e.g. Greater Accra"
            value={state.volunteerRegistration.region}
          />
          <Field
            label="Skills (comma separated)"
            onChangeText={(skills) =>
              actions.saveVolunteerRegistration({
                ...state.volunteerRegistration,
                skills,
              })
            }
            placeholder="e.g. first aid, community alerts"
            value={state.volunteerRegistration.skills}
          />
          <ActionButton
            icon="user-check"
            label="Register volunteer profile"
            onPress={actions.registerVolunteer}
            tone="navy"
          />
        </Card>
      )}

      {profile ? (
        <Card>
          <View style={stylesRow}>
            <View style={stylesGrow}>
              <Text style={stylesSectionTitle}>Safety rules</Text>
              <Text style={stylesMuted}>For community volunteers</Text>
            </View>
            <StatusPill label="112 escalation" tone="danger" />
          </View>
          {profile.safetyNotes.map((note) => (
            <Text key={note} style={stylesBody}>
              {note}
            </Text>
          ))}
        </Card>
      ) : null}

      {state.volunteerTasks.length === 0 ? (
        <Card>
          <EmptyState
            description="Assignments appear here after you register and refresh with a connection."
            icon="clipboard"
            title="No assignments yet"
          />
        </Card>
      ) : (
        state.volunteerTasks.map((task) => (
          <VolunteerTaskCard
            key={task.id}
            onUpdate={(status) => actions.updateVolunteerStatus(task.id, status)}
            task={task}
          />
        ))
      )}

      {activeTask ? (
        <Card>
          <View style={stylesRow}>
            <View style={stylesGrow}>
              <Text style={stylesSectionTitle}>Field observation</Text>
              <Text style={stylesMuted}>{activeTask.incidentReference}</Text>
            </View>
            <StatusPill
              label={
                state.volunteerObservation.escalationRequested
                  ? "escalate"
                  : "observe"
              }
              tone={
                state.volunteerObservation.escalationRequested
                  ? "danger"
                  : "navy"
              }
            />
          </View>
          <SegmentedControl
            onChange={(value) =>
              actions.saveVolunteerObservation({
                ...state.volunteerObservation,
                safetyStatus: value,
              })
            }
            options={safetyOptions}
            value={state.volunteerObservation.safetyStatus}
          />
          <Field
            label="Observation"
            multiline
            onChangeText={(note) =>
              actions.saveVolunteerObservation({
                ...state.volunteerObservation,
                note,
              })
            }
            placeholder="What did you safely observe?"
            value={state.volunteerObservation.note}
          />
          <View style={stylesButtonGrid}>
            <ActionButton
              icon="alert-triangle"
              label={
                state.volunteerObservation.escalationRequested
                  ? "Escalation requested"
                  : "Request authority escalation"
              }
              onPress={() =>
                actions.saveVolunteerObservation({
                  ...state.volunteerObservation,
                  escalationRequested:
                    !state.volunteerObservation.escalationRequested,
                })
              }
              tone={
                state.volunteerObservation.escalationRequested
                  ? "danger"
                  : "plain"
              }
            />
            <ActionButton
              icon="send"
              label="Submit observation"
              onPress={() => actions.submitVolunteerObservation(activeTask.id)}
              tone="green"
            />
          </View>
        </Card>
      ) : null}
    </View>
  );
}

function VolunteerTaskCard({
  onUpdate,
  task,
}: {
  onUpdate: (status: VolunteerStatusAction) => void;
  task: VolunteerTaskRecord;
}) {
  const isClosed = ["completed", "cancelled"].includes(task.status);
  const nextActions = statusActions.filter((item) =>
    (allowedNextStatuses[task.status] ?? []).includes(item.status),
  );
  return (
    <Card tone={task.escalationRequired ? "danger" : "plain"}>
      <View style={stylesRow}>
        <View style={stylesGrow}>
          <Text style={stylesSectionTitle}>{task.locationLabel}</Text>
          <Text style={stylesMuted}>{task.incidentReference}</Text>
        </View>
        <StatusPill
          label={task.status}
          tone={task.escalationRequired ? "danger" : "gold"}
        />
      </View>
      <View style={stylesBadgeRow}>
        <SeverityBadge severity={task.priority} />
      </View>
      <Text style={stylesBody}>{task.instructions}</Text>
      <View style={stylesTaskMeta}>
        <Metric label="Priority" value={task.priority} />
        <Metric label="Updates" value={task.updates.length} />
      </View>
      <View style={stylesButtonGrid}>
        {nextActions.map((item) => (
          <ActionButton
            disabled={isClosed}
            icon={item.icon}
            key={item.status}
            label={item.label}
            onPress={() => onUpdate(item.status)}
            tone={item.tone}
          />
        ))}
      </View>
    </Card>
  );
}

const stylesBadgeRow = {
  flexDirection: "row",
  flexWrap: "wrap",
  gap: mobileTheme.spacing.sm,
};

const stylesBody = {
  color: mobileTheme.colors.ink,
  fontFamily: mobileTheme.font.regular,
  fontSize: 15,
  lineHeight: 22,
};

const stylesButtonGrid = {
  gap: mobileTheme.spacing.sm,
};

const stylesGrow = {
  flex: 1,
};

const stylesMetricRow = {
  flexDirection: "row",
  gap: 10,
};

const stylesMuted = {
  color: mobileTheme.colors.muted,
  fontFamily: mobileTheme.font.regular,
  fontSize: 13,
};

const stylesRow = {
  alignItems: "center",
  flexDirection: "row",
  gap: mobileTheme.spacing.md,
};

const stylesSectionTitle = {
  color: mobileTheme.colors.navy,
  fontFamily: mobileTheme.font.bold,
  fontSize: 18,
};

const stylesTaskMeta = {
  flexDirection: "row",
  gap: 10,
};
