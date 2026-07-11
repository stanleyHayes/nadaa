import { Text, View } from "react-native";
import { mobileTheme } from "../../../app/theme";
import {
  Card,
  EmptyState,
  ScreenHeading,
  StatusPill,
} from "../../../ui/components";
import type { CitizenScreenProps } from "./types";

export function GuidesScreen({ state }: CitizenScreenProps) {
  return (
    <View style={stylesStack}>
      <ScreenHeading kicker="Offline guidance" title="Emergency guides" />
      <Card tone="green">
        <Text style={stylesTitle}>
          {state.offlineGuideCount} guides ready offline
        </Text>
        <Text style={stylesBody}>
          Saved guidance is available even when data service is interrupted.
        </Text>
      </Card>
      {state.guides.length === 0 ? (
        <EmptyState
          description="No guides are saved yet. Reconnect to download offline guidance."
          icon="book-open"
          title="No guides yet"
        />
      ) : (
        state.guides.map((guide) => (
        <Card key={guide.id}>
          <View style={stylesRow}>
            <View style={stylesGrow}>
              <Text style={stylesTitle}>{guide.title}</Text>
              <Text style={stylesMuted}>
                {guide.hazardType} · {guide.stage} ·{" "}
                {guide.language.toUpperCase()}
              </Text>
            </View>
            {guide.offlineAvailable ? (
              <StatusPill label="Offline" tone="green" />
            ) : null}
          </View>
          <Text style={stylesBody}>{guide.body}</Text>
        </Card>
        ))
      )}
    </View>
  );
}

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
