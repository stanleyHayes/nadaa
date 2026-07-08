import { Text, View } from "react-native";
import { hexToRgba, mobileTheme } from "../../../app/theme";
import {
  ActionButton,
  Card,
  ScreenHeading,
  StatusPill,
} from "../../../ui/components";
import { permissionCopy, permissionMessage } from "../permissions";
import type { CitizenScreenProps } from "./types";

export function SupportScreen({ actions, state }: CitizenScreenProps) {
  return (
    <View style={stylesStack}>
      <ScreenHeading kicker="Shelters and setup" title="Help around you" />
      <Card tone="green">
        <Text style={stylesTitle}>Nearest shelters</Text>
        {state.shelters.shelters.slice(0, 3).map((shelter) => (
          <View key={shelter.id} style={stylesListItem}>
            <View style={stylesGrow}>
              <Text style={stylesListTitle}>{shelter.name}</Text>
              <Text style={stylesMuted}>
                {shelter.status} · {shelter.distanceMeters ?? 0} m
              </Text>
            </View>
            <StatusPill label={shelter.contact ?? "112"} tone="green" />
          </View>
        ))}
      </Card>

      <Card>
        <Text style={stylesTitle}>Permissions</Text>
        {(
          Object.keys(state.permissions) as Array<
            keyof typeof state.permissions
          >
        ).map((key) => (
          <View key={key} style={stylesPermissionRow}>
            <View style={stylesGrow}>
              <Text style={stylesListTitle}>{permissionCopy[key].title}</Text>
              <Text style={stylesMuted}>
                {permissionMessage(key, state.permissions[key])}
              </Text>
            </View>
            <ActionButton
              icon="check-circle"
              label={state.permissions[key]}
              onPress={() => void actions.togglePermission(key)}
              tone={state.permissions[key] === "granted" ? "green" : "plain"}
            />
          </View>
        ))}
      </Card>

      <Card>
        <Text style={stylesTitle}>Session</Text>
        <Text style={stylesMuted}>
          {state.session.isGuest
            ? "Guest mode keeps emergency actions available while login is completed."
            : `Signed in as ${state.session.phone}`}
        </Text>
        <ActionButton
          icon="user-check"
          label="Use demo citizen session"
          onPress={() => void actions.updateSessionPhone("+233200000000")}
          tone="navy"
        />
        <Text style={stylesMuted}>
          Push status:{" "}
          {state.pushState.status === "registered"
            ? `${state.pushState.provider} registered`
            : state.pushState.message}
        </Text>
      </Card>
    </View>
  );
}

const stylesGrow = {
  flex: 1,
};

const stylesListItem = {
  alignItems: "center",
  flexDirection: "row",
  gap: mobileTheme.spacing.md,
};

const stylesListTitle = {
  color: mobileTheme.colors.ink,
  fontFamily: mobileTheme.font.semibold,
  fontSize: 15,
};

const stylesMuted = {
  color: mobileTheme.colors.muted,
  fontFamily: mobileTheme.font.regular,
  fontSize: 13,
  lineHeight: 19,
};

const stylesPermissionRow = {
  alignItems: "center",
  borderTopColor: hexToRgba(mobileTheme.colors.navy, 0.08),
  borderTopWidth: 1,
  flexDirection: "row",
  gap: 10,
  paddingTop: mobileTheme.spacing.md,
  minHeight: 44,
};

const stylesStack = {
  gap: mobileTheme.spacing.md + 2,
};

const stylesTitle = {
  color: mobileTheme.colors.navy,
  fontFamily: mobileTheme.font.bold,
  fontSize: 18,
};
