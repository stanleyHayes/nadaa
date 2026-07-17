import { Text, View } from "react-native";
import {
  ActionButton,
  Card,
  Field,
  ScreenHeading,
  StatusPill,
  uiStyles,
} from "../../../ui/components";
import { mobileTheme } from "../../../app/theme";
import { permissionMessage } from "../permissions";
import type { DispatcherScreenProps } from "./types";

export function ProfileScreen({ actions, state }: DispatcherScreenProps) {
  const session = state.session;
  const permissions = state.permissions;

  return (
    <View style={uiStyles.card_plain}>
      <ScreenHeading
        kicker="Dispatcher profile"
        title="Session & permissions"
      />

      <Card tone="navy">
        <Text style={stylesHeroTitle}>{session.userName}</Text>
        <Text style={stylesHeroText}>{session.agencyName}</Text>
        <View style={stylesRow}>
          <StatusPill label={session.role} tone="gold" />
          <StatusPill
            label={session.mfaCompleted ? "MFA verified" : "MFA needed"}
            tone={session.mfaCompleted ? "green" : "danger"}
          />
        </View>
      </Card>

      <Card>
        <Text style={stylesSectionTitle}>Agency sign in</Text>
        <Field
          label="Email"
          onChangeText={(value) => actions.updateAuthForm({ email: value })}
          placeholder="dispatcher@nadaa.gov.gh"
          value={state.authForm.email}
        />
        <Field
          label="Password"
          onChangeText={(value) => actions.updateAuthForm({ password: value })}
          placeholder="••••••••"
          value={state.authForm.password}
        />
        <Field
          label="MFA code"
          onChangeText={(value) => actions.updateAuthForm({ mfaCode: value })}
          placeholder="000000"
          value={state.authForm.mfaCode}
        />
        <ActionButton
          disabled={!state.authForm.email || !state.authForm.password}
          icon="log-in"
          label="Sign in"
          onPress={actions.login}
          tone="green"
        />
        <Text style={stylesMuted}>
          Dev fixture session is active when auth-service is unavailable.
        </Text>
      </Card>

      <Card>
        <Text style={stylesSectionTitle}>Permissions</Text>
        {(Object.keys(permissions) as Array<keyof typeof permissions>).map(
          (key) => (
            <View key={key} style={stylesPermissionRow}>
              <View style={stylesGrow}>
                <Text style={stylesBody}>
                  {key.charAt(0).toUpperCase() + key.slice(1)}
                </Text>
                <Text style={stylesMuted}>
                  {permissionMessage(key, permissions[key])}
                </Text>
              </View>
              <ActionButton
                icon={permissions[key] === "granted" ? "check" : "x"}
                label={permissions[key]}
                onPress={() => actions.togglePermission(key)}
                tone={permissions[key] === "granted" ? "green" : "plain"}
              />
            </View>
          ),
        )}
      </Card>

      <Card>
        <Text style={stylesSectionTitle}>Push state</Text>
        <Text style={stylesBody}>{state.pushState.status}</Text>
        {state.pushState.status === "registered" ? (
          <Text style={stylesMuted}>Token: {state.pushState.token}</Text>
        ) : (
          <Text style={stylesMuted}>{state.pushState.message}</Text>
        )}
      </Card>

      <Card>
        <Text style={stylesSectionTitle}>System</Text>
        <Text style={stylesBody}>{state.loadState.status}</Text>
        <Text style={stylesMuted}>{state.loadState.message}</Text>
      </Card>
    </View>
  );
}

const stylesBody = {
  color: mobileTheme.colors.ink,
  fontFamily: mobileTheme.font.regular,
  fontSize: 15,
  lineHeight: 22,
};

const stylesGrow = {
  flex: 1,
};

const stylesHeroText = {
  color: mobileTheme.colors.textInverse,
  fontFamily: mobileTheme.font.regular,
  fontSize: 15,
  lineHeight: 22,
  opacity: 0.78,
};

const stylesHeroTitle = {
  color: mobileTheme.colors.textInverse,
  fontFamily: mobileTheme.font.bold,
  fontSize: 23,
};

const stylesMuted = {
  color: mobileTheme.colors.muted,
  fontFamily: mobileTheme.font.regular,
  fontSize: 13,
};

const stylesPermissionRow = {
  alignItems: "center",
  flexDirection: "row",
  gap: mobileTheme.spacing.md,
  paddingVertical: mobileTheme.spacing.sm,
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
