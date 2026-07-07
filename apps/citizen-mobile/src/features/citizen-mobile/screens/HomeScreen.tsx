import { Text, View } from "react-native";
import { mobileAreaPresets } from "../data";
import {
  ActionButton,
  Card,
  Metric,
  ScreenHeading,
  StatusPill,
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
          <StatusPill label={state.risk.overallRisk} tone="gold" />
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
  color: "#101828",
  fontFamily: "Outfit_400Regular",
  fontSize: 15,
  lineHeight: 22,
};

const stylesButtonGrid = {
  gap: 8,
};

const stylesGrow = {
  flex: 1,
};

const stylesHeroText = {
  color: "rgba(255, 255, 255, 0.78)",
  fontFamily: "Outfit_400Regular",
  fontSize: 15,
  lineHeight: 22,
};

const stylesHeroTitle = {
  color: "#FFFFFF",
  fontFamily: "Outfit_800ExtraBold",
  fontSize: 23,
};

const stylesMetricRow = {
  flexDirection: "row",
  gap: 10,
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
