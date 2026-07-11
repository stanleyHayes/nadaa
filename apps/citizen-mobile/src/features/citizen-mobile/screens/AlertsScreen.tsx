import { Text, View } from "react-native";
import { mobileTheme } from "../../../app/theme";
import {
  Card,
  EmptyState,
  HazardBadge,
  ScreenHeading,
  SegmentedControl,
  SeverityBadge,
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
      {state.visibleAlerts.length === 0 ? (
        <EmptyState
          description={
            state.alertView === "current"
              ? "There are no current warnings for your area right now. You're all clear."
              : "No alerts match this view yet."
          }
          icon="bell"
          title="No alerts to show"
        />
      ) : (
        state.visibleAlerts.map((alert) => (
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
          <View style={stylesBadgeRow}>
            <HazardBadge hazard={alert.hazardType} />
            <SeverityBadge severity={alert.severity} />
          </View>
          <Text style={stylesBody}>{alert.message}</Text>
          <Text style={stylesAction}>{alert.recommendedAction}</Text>
        </Card>
        ))
      )}
    </View>
  );
}

const stylesAction = {
  color: mobileTheme.colors.navy,
  fontFamily: mobileTheme.font.semibold,
  fontSize: 14,
  lineHeight: 20,
};

const stylesBadgeRow = {
  flexDirection: "row",
  flexWrap: "wrap",
  gap: mobileTheme.spacing.sm,
};

const stylesBody = {
  color: mobileTheme.colors.ink,
  fontFamily: mobileTheme.font.regular,
  fontSize: 14,
  lineHeight: 21,
};

const stylesGrow = {
  flex: 1,
};

const stylesMuted = {
  color: mobileTheme.colors.muted,
  fontFamily: mobileTheme.font.regular,
  fontSize: 13,
};

const stylesRow = {
  alignItems: "flex-start",
  flexDirection: "row",
  gap: mobileTheme.spacing.md,
};

const stylesStack = {
  gap: mobileTheme.spacing.md + 2,
};

const stylesTitle = {
  color: mobileTheme.colors.navy,
  fontFamily: mobileTheme.font.bold,
  fontSize: 18,
};
