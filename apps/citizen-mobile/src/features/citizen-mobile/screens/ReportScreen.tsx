import Feather from "@expo/vector-icons/Feather";
import { Linking, Pressable, Switch, Text, View } from "react-native";
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
      <View style={stylesDistressCard}>
        <View style={stylesDistressHeading}>
          <Feather
            color={mobileTheme.colors.white}
            name="alert-octagon"
            size={24}
          />
          <View style={{ flex: 1 }}>
            <Text style={stylesDistressTitle}>In danger? Request rescue</Text>
            <Text style={stylesDistressCopy}>
              Share GPS with dispatch. This does not replace a 112 call.
            </Text>
          </View>
        </View>
        <View style={stylesActions}>
          <Pressable
            accessibilityLabel="Call 112"
            accessibilityRole="button"
            onPress={() => void Linking.openURL("tel:112")}
            style={stylesDistressSecondary}
          >
            <Feather
              color={mobileTheme.colors.white}
              name="phone-call"
              size={17}
            />
            <Text style={stylesDistressButtonText}>Call 112</Text>
          </Pressable>
          <Pressable
            accessibilityLabel="Refresh rescue GPS location"
            accessibilityRole="button"
            onPress={() => void actions.togglePermission("location")}
            style={stylesDistressSecondary}
          >
            <Feather
              color={mobileTheme.colors.white}
              name="map-pin"
              size={17}
            />
            <Text style={stylesDistressButtonText}>Refresh GPS</Text>
          </Pressable>
        </View>
        <Pressable
          accessibilityHint="Hold for just over two seconds to prevent an accidental SOS"
          accessibilityLabel="Hold to send SOS rescue request"
          accessibilityRole="button"
          delayLongPress={2200}
          onLongPress={() => void actions.submitDistress()}
          style={stylesDistressPrimary}
        >
          <Feather color={mobileTheme.colors.danger} name="radio" size={17} />
          <Text style={stylesDistressPrimaryText}>Hold to send SOS</Text>
        </Pressable>
        <Text style={stylesDistressHint}>
          Hold the SOS button for 2.2 seconds to send
        </Text>
      </View>
      {state.loadState.message ? (
        <View style={stylesFeedback} accessibilityLiveRegion="polite">
          <Feather
            color={mobileTheme.colors.info}
            name={state.loadState.status === "error" ? "alert-circle" : "info"}
            size={17}
          />
          <Text style={stylesFeedbackText}>{state.loadState.message}</Text>
        </View>
      ) : null}
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

const stylesDistressCard = {
  backgroundColor: mobileTheme.colors.danger,
  borderRadius: mobileTheme.radius.lg,
  gap: mobileTheme.spacing.md,
  padding: mobileTheme.spacing.lg,
};

const stylesDistressHeading = {
  alignItems: "flex-start",
  flexDirection: "row",
  gap: mobileTheme.spacing.md,
};

const stylesDistressTitle = {
  color: mobileTheme.colors.white,
  fontFamily: mobileTheme.font.bold,
  fontSize: 18,
};

const stylesDistressCopy = {
  color: hexToRgba(mobileTheme.colors.white, 0.88),
  fontFamily: mobileTheme.font.regular,
  fontSize: 13,
  lineHeight: 19,
  marginTop: 3,
};

const stylesDistressSecondary = {
  alignItems: "center",
  borderColor: hexToRgba(mobileTheme.colors.white, 0.48),
  borderRadius: mobileTheme.radius.md,
  borderWidth: 1,
  flex: 1,
  flexDirection: "row",
  gap: 7,
  justifyContent: "center",
  minHeight: 48,
};

const stylesDistressPrimary = {
  alignItems: "center",
  backgroundColor: mobileTheme.colors.white,
  borderRadius: mobileTheme.radius.md,
  flexDirection: "row",
  gap: 7,
  justifyContent: "center",
  minHeight: 48,
};

const stylesDistressButtonText = {
  color: mobileTheme.colors.white,
  fontFamily: mobileTheme.font.bold,
  fontSize: 13,
};

const stylesDistressPrimaryText = {
  color: mobileTheme.colors.danger,
  fontFamily: mobileTheme.font.bold,
  fontSize: 13,
};

const stylesDistressHint = {
  color: hexToRgba(mobileTheme.colors.white, 0.78),
  fontFamily: mobileTheme.font.semibold,
  fontSize: 11,
  textAlign: "center",
};

const stylesFeedback = {
  alignItems: "flex-start",
  backgroundColor: mobileTheme.colors.softBlue,
  borderRadius: mobileTheme.radius.md,
  flexDirection: "row",
  gap: mobileTheme.spacing.sm,
  padding: mobileTheme.spacing.md,
};

const stylesFeedbackText = {
  color: mobileTheme.colors.ink,
  flex: 1,
  fontFamily: mobileTheme.font.semibold,
  fontSize: 13,
  lineHeight: 19,
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
