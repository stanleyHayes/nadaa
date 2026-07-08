import { Text, View } from "react-native";
import { nadaaBrand } from "@nadaa/brand";
import {
  ActionButton,
  Card,
  ListItem,
  Metric,
  ScreenHeading,
  SeverityBadge,
  StatusPill,
  uiStyles,
} from "../../../ui/components";
import { mobileTheme } from "../../../app/theme";
import { formatRelativeTime, hazardOptions, statusLabel } from "../data";
import type { DispatcherScreenProps } from "./types";

export function QueueScreen({ actions, state }: DispatcherScreenProps) {
  return (
    <View style={uiStyles.card_plain}>
      <ScreenHeading kicker="Incident queue" title="Dispatch triage" />

      <Card tone="navy">
        <Text style={stylesHeroTitle}>Shift summary</Text>
        <Text style={stylesHeroText}>
          {state.loadState.message ?? "Review, verify, and assign incidents."}
        </Text>
        <View style={stylesMetricRow}>
          <Metric label="Urgent" value={state.queueMetrics.urgent} />
          <Metric label="Open" value={state.queueMetrics.open} />
          <Metric label="Total" value={state.queueMetrics.total} />
        </View>
      </Card>

      <Card>
        <View style={stylesRow}>
          <Text style={stylesSectionTitle}>Filters</Text>
          <ActionButton
            icon="refresh-cw"
            label="Refresh"
            onPress={actions.refreshQueue}
            tone="plain"
          />
        </View>
        <Text style={stylesMuted}>Hazard</Text>
        <View style={stylesChipRow}>
          {[
            { label: "All", value: "all" },
            ...hazardOptions.map((option) => ({
              label: option.label,
              value: option.value,
            })),
          ].map((option) => (
            <ActionButton
              icon=""
              key={option.value}
              label={option.label}
              onPress={() =>
                actions.updateFilter(
                  "hazard",
                  option.value as typeof state.filters.hazard,
                )
              }
              tone={state.filters.hazard === option.value ? "navy" : "plain"}
            />
          ))}
        </View>
        <Text style={stylesMuted}>Status</Text>
        <View style={stylesChipRow}>
          {[
            { label: "All", value: "all" },
            { label: "Reported", value: "reported" },
            { label: "Verified", value: "verified" },
            { label: "Assigned", value: "assigned" },
            { label: "En route", value: "response_en_route" },
            { label: "On scene", value: "on_scene" },
          ].map((option) => (
            <ActionButton
              icon=""
              key={option.value}
              label={option.label}
              onPress={() =>
                actions.updateFilter(
                  "status",
                  option.value as typeof state.filters.status,
                )
              }
              tone={state.filters.status === option.value ? "navy" : "plain"}
            />
          ))}
        </View>
      </Card>

      <Card>
        <Text style={stylesSectionTitle}>
          {state.filteredIncidents.length} incidents
        </Text>
        {state.filteredIncidents.length === 0 ? (
          <Text style={stylesBody}>
            No incidents match the current filters.
          </Text>
        ) : (
          state.filteredIncidents.map((incident) => (
            <ListItem
              key={incident.id}
              onPress={() => actions.selectIncident(incident.id)}
              selected={state.selectedIncidentId === incident.id}
            >
              <View style={stylesRow}>
                <View style={stylesGrow}>
                  <Text style={stylesIncidentReference}>
                    {incident.reference}
                  </Text>
                  <Text style={stylesBody}>{incident.description}</Text>
                </View>
                <SeverityBadge severity={incident.severity} />
              </View>
              <View style={stylesRow}>
                <StatusPill label={statusLabel(incident.status)} tone="navy" />
                <Text style={stylesMuted}>
                  {formatRelativeTime(incident.createdAt)}
                </Text>
                {incident.priorityReview ? (
                  <StatusPill label="Priority" tone="danger" />
                ) : null}
              </View>
            </ListItem>
          ))
        )}
      </Card>

      <View style={stylesCallRow}>
        <ActionButton
          icon="phone-call"
          label={`Call ${nadaaBrand.supportLine}`}
          onPress={() => {}}
          tone="danger"
        />
      </View>
    </View>
  );
}

const stylesBody = {
  color: mobileTheme.colors.ink,
  fontFamily: mobileTheme.font.regular,
  fontSize: 15,
  lineHeight: 22,
};

const stylesCallRow = {
  paddingBottom: mobileTheme.spacing["3xl"],
};

const stylesChipRow = {
  flexDirection: "row",
  flexWrap: "wrap",
  gap: mobileTheme.spacing.sm,
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

const stylesIncidentReference = {
  color: mobileTheme.colors.navy,
  fontFamily: mobileTheme.font.bold,
  fontSize: 16,
};

const stylesMetricRow = {
  flexDirection: "row",
  gap: mobileTheme.spacing.md,
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
