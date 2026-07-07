import { Text, View } from "react-native";
import { Card, ScreenHeading, StatusPill } from "../../../ui/components";
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
      {state.guides.map((guide) => (
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
      ))}
    </View>
  );
}

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
