import { Text, View } from "react-native";
import { hexToRgba, mobileTheme } from "../../../app/theme";
import { mobileAreaPresets } from "../data";
import {
  ActionButton,
  Card,
  Metric,
  ScreenHeading,
  SeverityBadge,
  uiStyles,
} from "../../../ui/components";
import type { CitizenScreenProps } from "./types";

export function HomeScreen({ actions, state }: CitizenScreenProps) {
  const floodRisk = state.risk.risks.find((risk) => risk.type === "flood");
  return (
    <View style={uiStyles.card_plain}>
      <ScreenHeading
        kicker="Citizen mobile"
        title={`Hello, ${state.session.name}`}
      />
      <Card tone="navy">
        <Text style={stylesHeroTitle}>NADAA is watching your area.</Text>
        <Text style={stylesHeroText}>
          {state.loadState.message ??
            "Check risk, get alerts, save reports, and keep guides offline."}
        </Text>
        <View style={stylesMetricRow}>
          <Metric label="Current alerts" value={state.currentAlertCount} />
          <Metric label="Offline guides" value={state.offlineGuideCount} />
        </View>
      </Card>

      <Card>
        <View style={stylesRow}>
          <View style={stylesGrow}>
            <Text style={stylesSectionTitle}>Area risk</Text>
            <Text style={stylesMuted}>{state.risk.location}</Text>
          </View>
          <SeverityBadge severity={state.risk.overallRisk} />
        </View>
        {floodRisk ? (
          <Text style={stylesBody}>{floodRisk.reason}</Text>
        ) : (
          <Text style={stylesBody}>No flood risk details are available.</Text>
        )}
        <View style={stylesButtonGrid}>
          {mobileAreaPresets.map((preset, index) => (
            <ActionButton
              icon="map-pin"
              key={preset.label}
              label={preset.label}
              onPress={() => actions.chooseArea(index)}
              tone={
                state.selectedArea.label === preset.label ? "green" : "plain"
              }
            />
          ))}
        </View>
        <ActionButton
          icon="refresh-cw"
          label="Refresh risk and alerts"
          onPress={actions.refreshAll}
          tone="navy"
        />
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

const stylesButtonGrid = {
  gap: mobileTheme.spacing.sm,
};

const stylesGrow = {
  flex: 1,
};

const stylesHeroText = {
  color: hexToRgba(mobileTheme.colors.white, 0.78),
  fontFamily: mobileTheme.font.regular,
  fontSize: 15,
  lineHeight: 22,
};

const stylesHeroTitle = {
  color: mobileTheme.colors.white,
  fontFamily: mobileTheme.font.bold,
  fontSize: 23,
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
