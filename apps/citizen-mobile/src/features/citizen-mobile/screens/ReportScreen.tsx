import { Switch, Text, View } from "react-native";
import { hexToRgba, mobileTheme } from "../../../app/theme";
import {
  ActionButton,
  Card,
  Field,
  ScreenHeading,
} from "../../../ui/components";
import { hazardOptions } from "../data";
import type { CitizenScreenProps } from "./types";

export function ReportScreen({ actions, state }: CitizenScreenProps) {
  const draft = state.reportDraft;
  const save = (patch: Partial<typeof draft>) =>
    void actions.saveDraft({ ...draft, ...patch });

  return (
    <View style={stylesStack}>
      <ScreenHeading kicker="Emergency report" title="Save and submit safely" />
      <Card>
        <Text style={stylesMuted}>
          Drafts stay on the device until submitted. If the network drops, keep
          the draft and retry.
        </Text>
        <View style={stylesButtonGrid}>
          {hazardOptions.map((hazard) => (
            <ActionButton
              icon="alert-triangle"
              key={hazard.value}
              label={hazard.label}
              onPress={() => save({ hazard: hazard.value })}
              tone={draft.hazard === hazard.value ? "green" : "plain"}
            />
          ))}
        </View>
        <Field
          label="Latitude"
          onChangeText={(lat) => save({ lat })}
          value={draft.lat}
        />
        <Field
          label="Longitude"
          onChangeText={(lng) => save({ lng })}
          value={draft.lng}
        />
        <Field
          label="What happened?"
          multiline
          onChangeText={(description) => save({ description })}
          placeholder="Describe the emergency, landmarks, and who needs help."
          value={draft.description}
        />
        <Field
          label="People affected"
          onChangeText={(peopleAffected) => save({ peopleAffected })}
          value={draft.peopleAffected}
        />
        <ToggleRow
          label="Injuries reported"
          onChange={(injuriesReported) => save({ injuriesReported })}
          value={draft.injuriesReported}
        />
        <ToggleRow
          label="Send anonymously"
          onChange={(anonymous) =>
            save({
              anonymous,
              contactPermission: anonymous ? false : draft.contactPermission,
            })
          }
          value={draft.anonymous}
        />
        <ToggleRow
          label="Allow responder follow-up"
          onChange={(contactPermission) => save({ contactPermission })}
          value={draft.contactPermission}
        />
        <View style={stylesActions}>
          <ActionButton
            icon="save"
            label="Save draft"
            onPress={() => void actions.saveDraft(draft)}
            tone="plain"
          />
          <ActionButton
            icon="send"
            label="Submit report"
            onPress={() => void actions.submitDraft()}
            tone="danger"
          />
        </View>
      </Card>
    </View>
  );
}

function ToggleRow({
  label,
  onChange,
  value,
}: {
  label: string;
  onChange: (value: boolean) => void;
  value: boolean;
}) {
  return (
    <View style={stylesToggleRow}>
      <Text style={stylesLabel}>{label}</Text>
      <Switch
        accessibilityLabel={label}
        onValueChange={onChange}
        thumbColor={value ? mobileTheme.colors.green : mobileTheme.colors.white}
        trackColor={{
          false: hexToRgba(mobileTheme.colors.muted, 0.24),
          true: hexToRgba(mobileTheme.colors.green, 0.28),
        }}
        value={value}
      />
    </View>
  );
}

const stylesActions = {
  flexDirection: "row",
  gap: 10,
};

const stylesButtonGrid = {
  gap: mobileTheme.spacing.sm,
};

const stylesLabel = {
  color: mobileTheme.colors.ink,
  fontFamily: mobileTheme.font.semibold,
  fontSize: 14,
};

const stylesMuted = {
  color: mobileTheme.colors.muted,
  fontFamily: mobileTheme.font.regular,
  fontSize: 14,
  lineHeight: 21,
};

const stylesStack = {
  gap: mobileTheme.spacing.md + 2,
};

const stylesToggleRow = {
  alignItems: "center",
  flexDirection: "row",
  justifyContent: "space-between",
  minHeight: 44,
};
