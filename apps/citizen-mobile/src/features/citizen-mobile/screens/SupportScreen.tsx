import { Text, View } from "react-native";
import { hexToRgba, mobileTheme } from "../../../app/theme";
import {
  ActionButton,
  Card,
  EmptyState,
  Field,
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
        {state.shelters.shelters.length === 0 ? (
          <EmptyState
            description="Shelter locations load when you refresh with a connection. In an emergency call 112."
            icon="map-pin"
            title="No shelter data yet"
          />
        ) : (
          state.shelters.shelters.slice(0, 3).map((shelter) => (
            <View key={shelter.id} style={stylesListItem}>
              <View style={stylesGrow}>
                <Text style={stylesListTitle}>{shelter.name}</Text>
                <Text style={stylesMuted}>
                  {shelter.status} · {shelter.distanceMeters ?? 0} m
                </Text>
              </View>
              <StatusPill label={shelter.contact ?? "112"} tone="green" />
            </View>
          ))
        )}
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
        <Text style={stylesTitle}>Citizen sign-in</Text>
        {state.session.isGuest ? (
          <View style={stylesStack}>
            <Text style={stylesMuted}>
              Guest mode keeps emergency actions available. Sign in with your
              phone number to sync reports and volunteer assignments — we send
              a one-time code.
            </Text>
            <Field
              label="Full name (new accounts)"
              onChangeText={(name) =>
                actions.saveSignInDraft({ ...state.signIn, name })
              }
              placeholder="Ama Mensah"
              value={state.signIn.name}
            />
            <Field
              label="Phone"
              onChangeText={(phone) =>
                actions.saveSignInDraft({ ...state.signIn, phone })
              }
              placeholder="+233201234567"
              value={state.signIn.phone}
            />
            <ActionButton
              icon="message-circle"
              label="Send sign-in code"
              onPress={() => void actions.requestSignInCode()}
              tone="navy"
            />
            <Field
              label="One-time code"
              onChangeText={(otp) =>
                actions.saveSignInDraft({ ...state.signIn, otp })
              }
              placeholder="Code from SMS"
              value={state.signIn.otp}
            />
            <ActionButton
              icon="check-circle"
              label="Verify and sign in"
              onPress={() => void actions.verifySignIn()}
              tone="green"
            />
          </View>
        ) : (
          <Text style={stylesMuted}>
            {`Signed in as ${state.session.name} (${state.session.phone}).`}
          </Text>
        )}
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
