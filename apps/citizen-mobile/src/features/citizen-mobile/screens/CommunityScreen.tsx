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
  Field,
  Metric,
  ScreenHeading,
  SegmentedControl,
  SeverityBadge,
  StatusPill,
  uiStyles,
} from "../../../ui/components";
import type { CitizenScreenProps } from "./types";

const statusActions: Array<{
  icon: string;
  label: string;
  status: Exclude<VolunteerTaskStatus, "assigned">;
  tone: "danger" | "green" | "navy" | "plain";
}> = [
  { icon: "check-circle", label: "Accept", status: "accepted", tone: "green" },
  { icon: "navigation", label: "On scene", status: "on_scene", tone: "navy" },
  { icon: "flag", label: "Complete", status: "completed", tone: "plain" },
  {
    icon: "alert-triangle",
    label: "Escalate",
    status: "needs_escalation",
    tone: "danger",
  },
];

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

      <Card tone="green">
        <View style={stylesRow}>
          <View style={stylesGrow}>
            <Text style={stylesSectionTitle}>
              {state.volunteerProfile.name}
            </Text>
            <Text style={stylesMuted}>
              {state.volunteerProfile.community},{" "}
              {state.volunteerProfile.district}
            </Text>
          </View>
          <StatusPill
            label={state.volunteerProfile.verificationStatus}
            tone={
              state.volunteerProfile.verificationStatus === "verified"
                ? "green"
                : "gold"
            }
          />
        </View>
        <View style={stylesMetricRow}>
          <Metric label="Active tasks" value={state.activeVolunteerTaskCount} />
          <Metric label="Skills" value={state.volunteerProfile.skills.length} />
        </View>
        <Text style={stylesBody}>
          {state.volunteerProfile.skills.join(", ")} volunteer supporting{" "}
          {state.volunteerProfile.groupId}.
        </Text>
        <View style={stylesButtonGrid}>
          <ActionButton
            icon="user-check"
            label="Refresh volunteer profile"
            onPress={actions.registerVolunteer}
            tone="navy"
          />
          <ActionButton
            icon="refresh-cw"
            label="Refresh assignments"
            onPress={actions.refreshVolunteerTasks}
            tone="plain"
          />
        </View>
      </Card>

      <Card>
        <View style={stylesRow}>
          <View style={stylesGrow}>
            <Text style={stylesSectionTitle}>Safety rules</Text>
            <Text style={stylesMuted}>For community volunteers</Text>
          </View>
          <StatusPill label="112 escalation" tone="danger" />
        </View>
        {state.volunteerProfile.safetyNotes.map((note) => (
          <Text key={note} style={stylesBody}>
            {note}
          </Text>
        ))}
      </Card>

      {state.volunteerTasks.map((task) => (
        <VolunteerTaskCard
          key={task.id}
          onUpdate={(status) => actions.updateVolunteerStatus(task.id, status)}
          task={task}
        />
      ))}

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
  onUpdate: (status: Exclude<VolunteerTaskStatus, "assigned">) => void;
  task: VolunteerTaskRecord;
}) {
  const isClosed = ["completed", "cancelled"].includes(task.status);
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
        {statusActions.map((item) => (
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
