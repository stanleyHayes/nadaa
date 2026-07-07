import { Text, View } from "react-native";
import {
  Card,
  ScreenHeading,
  SegmentedControl,
  StatusPill,
} from "../../../ui/components";
import type { AlertView } from "../types";
import type { CitizenScreenProps } from "./types";

export function AlertsScreen({ actions, state }: CitizenScreenProps) {
  return (
    <View style={stylesStack}>
      <ScreenHeading kicker="Live warnings" title="Approved citizen alerts" />
      <SegmentedControl<AlertView>
        onChange={actions.setAlertView}
        options={[
          { label: "Current", value: "current" },
          { label: "Expired", value: "expired" },
          { label: "All", value: "all" },
        ]}
        value={state.alertView}
      />
      {state.visibleAlerts.map((alert) => (
        <Card
          key={alert.id}
          tone={alert.status === "current" ? "danger" : "plain"}
        >
          <View style={stylesRow}>
            <View style={stylesGrow}>
              <Text style={stylesTitle}>{alert.title}</Text>
              <Text style={stylesMuted}>{alert.targetLabel}</Text>
            </View>
            <StatusPill
              label={alert.status}
              tone={alert.status === "current" ? "danger" : "navy"}
            />
          </View>
          <Text style={stylesBody}>{alert.message}</Text>
          <Text style={stylesAction}>{alert.recommendedAction}</Text>
        </Card>
      ))}
    </View>
  );
}

const stylesAction = {
  color: "#0D1B3D",
  fontFamily: "Outfit_600SemiBold",
  fontSize: 14,
  lineHeight: 20,
};

const stylesBody = {
  color: "#101828",
  fontFamily: "Outfit_400Regular",
  fontSize: 14,
  lineHeight: 21,
};

const stylesGrow = {
  flex: 1,
};

const stylesMuted = {
  color: "#555B66",
  fontFamily: "Outfit_400Regular",
  fontSize: 13,
};

const stylesRow = {
  alignItems: "flex-start",
  flexDirection: "row",
  gap: 12,
};

const stylesStack = {
  gap: 14,
};

const stylesTitle = {
  color: "#0D1B3D",
  fontFamily: "Outfit_800ExtraBold",
  fontSize: 18,
};
