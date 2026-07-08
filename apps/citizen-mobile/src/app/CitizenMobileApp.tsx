import { useMemo, useState } from "react";
import Feather from "@expo/vector-icons/Feather";
import {
  Image,
  Pressable,
  SafeAreaView,
  ScrollView,
  StatusBar,
  StyleSheet,
  Text,
  View,
} from "react-native";
import { nadaaBrand } from "@nadaa/brand";
import { citizenTabs, type CitizenTab } from "./navigation";
import { hexToRgba, mobileTheme } from "./theme";
import { useCitizenMobileState } from "../features/citizen-mobile/useCitizenMobileState";
import { AlertsScreen } from "../features/citizen-mobile/screens/AlertsScreen";
import { CommunityScreen } from "../features/citizen-mobile/screens/CommunityScreen";
import { GuidesScreen } from "../features/citizen-mobile/screens/GuidesScreen";
import { HomeScreen } from "../features/citizen-mobile/screens/HomeScreen";
import { ReportScreen } from "../features/citizen-mobile/screens/ReportScreen";
import { SupportScreen } from "../features/citizen-mobile/screens/SupportScreen";

export default function CitizenMobileApp() {
  const [activeTab, setActiveTab] = useState<CitizenTab>("home");
  const controller = useCitizenMobileState();
  const screen = useMemo(() => {
    const props = { actions: controller.actions, state: controller.state };
    switch (activeTab) {
      case "alerts":
        return <AlertsScreen {...props} />;
      case "report":
        return <ReportScreen {...props} />;
      case "community":
        return <CommunityScreen {...props} />;
      case "guides":
        return <GuidesScreen {...props} />;
      case "support":
        return <SupportScreen {...props} />;
      case "home":
      default:
        return <HomeScreen {...props} />;
    }
  }, [activeTab, controller.actions, controller.state]);

  return (
    <SafeAreaView style={styles.safeArea}>
      <StatusBar
        backgroundColor={mobileTheme.colors.navy}
        barStyle="light-content"
      />
      <View style={styles.header}>
        <View style={styles.brandRow}>
          <Image
            accessibilityLabel="NADAA shield"
            source={require("../../assets/nadaa-logo.png")}
            style={styles.logo}
          />
          <View style={styles.brandText}>
            <Text style={styles.brandName}>{nadaaBrand.name}</Text>
            <Text style={styles.slogan}>{nadaaBrand.slogan}</Text>
          </View>
        </View>
        <Pressable
          accessibilityLabel="Call 112"
          accessibilityRole="button"
          style={styles.callButton}
        >
          <Feather
            color={mobileTheme.colors.white}
            name="phone-call"
            size={17}
          />
          <Text style={styles.callText}>{nadaaBrand.supportLine}</Text>
        </Pressable>
      </View>

      <ScrollView
        contentContainerStyle={styles.content}
        style={styles.scrollView}
      >
        {screen}
      </ScrollView>

      <View style={styles.tabs}>
        {citizenTabs.map((tab) => {
          const isActive = tab.id === activeTab;
          return (
            <Pressable
              accessibilityLabel={tab.label}
              accessibilityRole="button"
              key={tab.id}
              onPress={() => setActiveTab(tab.id)}
              style={[styles.tab, isActive ? styles.tabActive : null]}
            >
              <Feather
                color={
                  isActive ? mobileTheme.colors.white : mobileTheme.colors.muted
                }
                name={tab.icon}
                size={18}
              />
              <Text
                style={[
                  styles.tabLabel,
                  isActive ? styles.tabLabelActive : null,
                ]}
              >
                {tab.label}
              </Text>
            </Pressable>
          );
        })}
      </View>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  brandName: {
    color: mobileTheme.colors.white,
    fontFamily: mobileTheme.font.bold,
    fontSize: 22,
  },
  brandRow: {
    alignItems: "center",
    flex: 1,
    flexDirection: "row",
    gap: mobileTheme.spacing.md,
  },
  brandText: {
    flex: 1,
    gap: 2,
  },
  callButton: {
    alignItems: "center",
    borderColor: hexToRgba(mobileTheme.colors.white, 0.36),
    borderRadius: mobileTheme.radius.md,
    borderWidth: 1,
    flexDirection: "row",
    gap: mobileTheme.spacing.sm,
    minHeight: 44,
    minWidth: 44,
    paddingHorizontal: mobileTheme.spacing.md,
  },
  callText: {
    color: mobileTheme.colors.white,
    fontFamily: mobileTheme.font.bold,
    fontSize: 15,
  },
  content: {
    gap: mobileTheme.spacing.lg,
    padding: mobileTheme.spacing.lg,
    paddingBottom: 116,
  },
  header: {
    alignItems: "center",
    backgroundColor: mobileTheme.colors.navy,
    borderBottomColor: mobileTheme.colors.gold,
    borderBottomWidth: 4,
    flexDirection: "row",
    gap: mobileTheme.spacing.md,
    minHeight: 82,
    paddingHorizontal: mobileTheme.spacing.lg,
    paddingVertical: mobileTheme.spacing.md,
  },
  logo: {
    backgroundColor: mobileTheme.colors.white,
    borderRadius: mobileTheme.radius.md,
    height: 52,
    width: 52,
  },
  safeArea: {
    backgroundColor: mobileTheme.colors.background,
    flex: 1,
  },
  scrollView: {
    backgroundColor: mobileTheme.colors.background,
    flex: 1,
  },
  slogan: {
    color: hexToRgba(mobileTheme.colors.white, 0.74),
    fontFamily: mobileTheme.font.regular,
    fontSize: 12,
  },
  tab: {
    alignItems: "center",
    borderRadius: mobileTheme.radius.md,
    flex: 1,
    gap: 4,
    justifyContent: "center",
    minHeight: 58,
    minWidth: 44,
  },
  tabActive: {
    backgroundColor: mobileTheme.colors.navy,
  },
  tabLabel: {
    color: mobileTheme.colors.muted,
    fontFamily: mobileTheme.font.semibold,
    fontSize: 11,
  },
  tabLabelActive: {
    color: mobileTheme.colors.white,
  },
  tabs: {
    backgroundColor: mobileTheme.colors.white,
    borderTopColor: mobileTheme.colors.border,
    borderTopWidth: 1,
    bottom: 0,
    flexDirection: "row",
    gap: 4,
    left: 0,
    padding: 8,
    position: "absolute",
    right: 0,
  },
});
