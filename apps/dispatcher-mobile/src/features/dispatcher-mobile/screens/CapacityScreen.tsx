import { Text, View } from "react-native";
import {
  ActionButton,
  CapacityBadge,
  Card,
  EmptyState,
  ListItem,
  Metric,
  ScreenHeading,
  SelectField,
  uiStyles,
} from "../../../ui/components";
import { mobileTheme } from "../../../app/theme";
import { formatDateTime } from "../data";
import type { DispatcherScreenProps } from "./types";

export function CapacityScreen({ actions, state }: DispatcherScreenProps) {
  const incident = state.selectedIncident;

  return (
    <View style={uiStyles.card_plain}>
      <ScreenHeading kicker="Hospital capacity" title="Nearby facilities" />

      {!incident ? (
        <Card>
          <Text style={stylesBody}>
            Select an incident from the Queue tab to see hospital capacity near
            the incident location.
          </Text>
        </Card>
      ) : (
        <>
          <Card>
            <Text style={stylesSectionTitle}>{incident.reference}</Text>
            <Text style={stylesBody}>{incident.description}</Text>
            <Text style={stylesMuted}>
              Location: {incident.location.lat.toFixed(4)},{" "}
              {incident.location.lng.toFixed(4)}
            </Text>
          </Card>

          <Card>
            <Text style={stylesSectionTitle}>Filters</Text>
            <SelectField
              label="Emergency capacity"
              onChange={(value) =>
                actions.updateCapacityFilter("emergencyCapacity", value)
              }
              options={[
                { label: "All", value: "all" },
                { label: "Available", value: "available" },
                { label: "Limited", value: "limited" },
                { label: "Full", value: "full" },
              ]}
              value={state.capacityFilters.emergencyCapacity}
            />
            <SelectField
              label="Service"
              onChange={(value) =>
                actions.updateCapacityFilter("service", value)
              }
              options={[
                { label: "All", value: "all" },
                { label: "Emergency", value: "emergency" },
                { label: "Trauma", value: "trauma" },
                { label: "ICU", value: "icu" },
                { label: "Maternity", value: "maternity" },
                { label: "Pediatric", value: "pediatric" },
                { label: "Ambulance", value: "ambulance" },
                { label: "Oxygen", value: "oxygen" },
              ]}
              value={state.capacityFilters.service}
            />
            <ActionButton
              icon="refresh-cw"
              label="Refresh capacity"
              onPress={() => actions.refreshCapacityForIncident(incident)}
              tone="plain"
            />
          </Card>

          <Card>
            <Text style={stylesSectionTitle}>
              {state.filteredCapacity.length} facilities
            </Text>
            {state.filteredCapacity.length === 0 ? (
              <EmptyState
                description="No facilities match the current filters."
                icon="activity"
                title="No facilities"
              />
            ) : (
              state.filteredCapacity.map((facility) => (
                <ListItem key={facility.id}>
                  <View style={stylesRow}>
                    <View style={stylesGrow}>
                      <Text style={stylesFacilityName}>{facility.name}</Text>
                      <Text style={stylesMuted}>{facility.address}</Text>
                    </View>
                    <CapacityBadge capacity={facility.emergencyCapacity} />
                  </View>
                  <View style={stylesMetricRow}>
                    <Metric label="Beds" value={facility.availableBeds} />
                    <Metric label="ICU" value={facility.icuBedsAvailable} />
                    <Metric
                      label="Ambulances"
                      value={facility.ambulancesAvailable}
                    />
                  </View>
                  {facility.stale ? (
                    <Text style={stylesStale}>
                      Stale data · confirm before transfer
                    </Text>
                  ) : (
                    <Text style={stylesMuted}>
                      Updated {formatDateTime(facility.updatedAt)}
                    </Text>
                  )}
                </ListItem>
              ))
            )}
          </Card>
        </>
      )}
    </View>
  );
}

const stylesBody = {
  color: mobileTheme.colors.ink,
  fontFamily: mobileTheme.font.regular,
  fontSize: 15,
  lineHeight: 22,
};

const stylesFacilityName = {
  color: mobileTheme.colors.navy,
  fontFamily: mobileTheme.font.bold,
  fontSize: 16,
};

const stylesGrow = {
  flex: 1,
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

const stylesStale = {
  color: mobileTheme.colors.danger,
  fontFamily: mobileTheme.font.semibold,
  fontSize: 13,
};
