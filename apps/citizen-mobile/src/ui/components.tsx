import Feather from "@expo/vector-icons/Feather";
import { Pressable, StyleSheet, Text, TextInput, View } from "react-native";
import { hazardBadgeFor, severityBadgeFor } from "@nadaa/brand/native";
import { hexToRgba, mobileTheme } from "../app/theme";

type ActionButtonProps = {
  disabled?: boolean;
  icon: string;
  label: string;
  onPress?: () => void;
  tone?: "danger" | "gold" | "green" | "navy" | "plain";
};

export function ActionButton({
  disabled = false,
  icon,
  label,
  onPress,
  tone = "navy",
}: ActionButtonProps) {
  const color =
    tone === "plain"
      ? mobileTheme.colors.navy
      : tone === "navy"
        ? mobileTheme.colors.white
        : mobileTheme.colors.ink;
  return (
    <Pressable
      accessibilityLabel={label}
      accessibilityRole="button"
      disabled={disabled}
      onPress={onPress}
      style={[
        styles.actionButton,
        styles[`button_${tone}`],
        disabled ? styles.disabled : null,
      ]}
    >
      <Feather color={color} name={icon} size={17} />
      <Text style={[styles.actionButtonText, { color }]}>{label}</Text>
    </Pressable>
  );
}

export function Card({
  children,
  tone = "plain",
}: {
  children: unknown;
  tone?: "danger" | "green" | "navy" | "plain";
}) {
  return <View style={[styles.card, styles[`card_${tone}`]]}>{children}</View>;
}

export function Field({
  label,
  multiline,
  onChangeText,
  placeholder,
  value,
}: {
  label: string;
  multiline?: boolean;
  onChangeText: (value: string) => void;
  placeholder?: string;
  value: string;
}) {
  return (
    <View style={styles.field}>
      <Text style={styles.label}>{label}</Text>
      <TextInput
        multiline={multiline}
        onChangeText={onChangeText}
        placeholder={placeholder}
        placeholderTextColor={hexToRgba(mobileTheme.colors.muted, 0.72)}
        style={[styles.input, multiline ? styles.inputMultiline : null]}
        value={value}
      />
    </View>
  );
}

export function Metric({
  label,
  value,
}: {
  label: string;
  value: number | string;
}) {
  return (
    <View style={styles.metric}>
      <Text style={styles.metricValue}>{value}</Text>
      <Text style={styles.metricLabel}>{label}</Text>
    </View>
  );
}

export function ScreenHeading({
  kicker,
  title,
}: {
  kicker?: string;
  title: string;
}) {
  return (
    <View style={styles.heading}>
      {kicker ? <Text style={styles.kicker}>{kicker}</Text> : null}
      <Text style={styles.title}>{title}</Text>
    </View>
  );
}

export function SegmentedControl<Option extends string>({
  onChange,
  options,
  value,
}: {
  onChange: (value: Option) => void;
  options: Array<{ label: string; value: Option }>;
  value: Option;
}) {
  return (
    <View style={styles.segmented}>
      {options.map((option) => (
        <Pressable
          accessibilityLabel={option.label}
          accessibilityRole="button"
          key={option.value}
          onPress={() => onChange(option.value)}
          style={[
            styles.segment,
            value === option.value ? styles.segmentActive : null,
          ]}
        >
          <Text
            style={[
              styles.segmentText,
              value === option.value ? styles.segmentTextActive : null,
            ]}
          >
            {option.label}
          </Text>
        </Pressable>
      ))}
    </View>
  );
}

export function StatusPill({
  label,
  tone = "navy",
}: {
  label: string;
  tone?: "danger" | "gold" | "green" | "navy";
}) {
  const pillStyle =
    tone === "danger"
      ? styles.pill_danger
      : tone === "gold"
        ? styles.pill_gold
        : tone === "green"
          ? styles.pill_green
          : styles.pill_navy;
  return (
    <View style={[styles.pill, pillStyle]}>
      <Text style={[styles.pillText, styles[`pillText_${tone}`]]}>{label}</Text>
    </View>
  );
}

/** Accessible severity badge with icon + text + color. */
export function SeverityBadge({ severity }: { severity: string }) {
  const normalized = normalizeSeverity(severity);
  const badge = severityBadgeFor(normalized);
  const icon = severityIconMap[badge.icon] ?? "info";
  return (
    <View
      style={[
        styles.badge,
        {
          backgroundColor: badge.background,
          borderColor: badge.border,
        },
      ]}
    >
      <Feather color={badge.color} name={icon} size={14} />
      <Text style={[styles.badgeText, { color: badge.color }]}>{severity}</Text>
    </View>
  );
}

/** Accessible hazard badge with icon + text + color. */
export function HazardBadge({ hazard }: { hazard: string }) {
  const badge = hazardBadgeFor(hazard);
  const icon = hazardIconMap[hazard.toLowerCase()] ?? "alert-circle";
  return (
    <View
      style={[
        styles.badge,
        {
          backgroundColor: badge.background,
          borderColor: badge.border,
        },
      ]}
    >
      <Feather color={badge.color} name={icon} size={14} />
      <Text style={[styles.badgeText, { color: badge.color }]}>{hazard}</Text>
    </View>
  );
}

function normalizeSeverity(severity: string): string {
  const map: Record<string, string> = {
    advisory: "info",
    emergency: "severe",
    high: "high",
    life_threatening: "severe",
    low: "low",
    moderate: "medium",
    normal: "low",
    severe: "severe",
    severe_warning: "severe",
    watch: "medium",
    warning: "high",
  };
  return map[severity.toLowerCase()] ?? "info";
}

const severityIconMap: Record<string, string> = {
  AlertOctagon: "alert-octagon",
  AlertTriangle: "alert-triangle",
  CheckCircle2: "check-circle",
  Info: "info",
};

const hazardIconMap: Record<string, string> = {
  blocked_drain: "droplet",
  building_collapse: "home",
  disease_outbreak: "thermometer",
  electrical_hazard: "zap",
  fire: "flame",
  flood: "droplet",
  landslide: "anchor",
  medical_emergency: "activity",
  other: "alert-circle",
  road_crash: "truck",
  security_incident: "shield",
};

const styles = StyleSheet.create({
  emptyState: {
    alignItems: "center",
    gap: mobileTheme.spacing.sm,
    paddingVertical: mobileTheme.spacing.xl,
    paddingHorizontal: mobileTheme.spacing.md,
  },
  emptyIcon: {
    alignItems: "center",
    justifyContent: "center",
    width: 60,
    height: 60,
    borderRadius: 30,
    backgroundColor: hexToRgba(mobileTheme.colors.navy, 0.08),
  },
  emptyTitle: {
    color: mobileTheme.colors.navy,
    fontFamily: mobileTheme.font.bold,
    fontSize: 16,
    textAlign: "center",
  },
  emptyDescription: {
    color: mobileTheme.colors.muted,
    fontFamily: mobileTheme.font.regular,
    fontSize: 14,
    lineHeight: 20,
    maxWidth: 300,
    textAlign: "center",
  },
  actionButton: {
    alignItems: "center",
    borderRadius: mobileTheme.radius.md,
    flexDirection: "row",
    gap: mobileTheme.spacing.sm,
    justifyContent: "center",
    minHeight: 44,
    minWidth: 44,
    paddingHorizontal: mobileTheme.spacing.md,
    paddingVertical: mobileTheme.spacing.sm,
  },
  actionButtonText: {
    fontFamily: mobileTheme.font.semibold,
    fontSize: 14,
  },
  badge: {
    alignItems: "center",
    alignSelf: "flex-start",
    borderRadius: 999,
    borderWidth: 1,
    flexDirection: "row",
    gap: 6,
    paddingHorizontal: 10,
    paddingVertical: 6,
  },
  badgeText: {
    fontFamily: mobileTheme.font.semibold,
    fontSize: 12,
  },
  button_danger: {
    backgroundColor: mobileTheme.colors.danger,
  },
  button_gold: {
    backgroundColor: mobileTheme.colors.gold,
  },
  button_green: {
    backgroundColor: mobileTheme.colors.green,
  },
  button_navy: {
    backgroundColor: mobileTheme.colors.navy,
  },
  button_plain: {
    backgroundColor: mobileTheme.colors.white,
    borderColor: mobileTheme.colors.border,
    borderWidth: 1,
  },
  card: {
    backgroundColor: mobileTheme.colors.card,
    borderColor: mobileTheme.colors.border,
    borderRadius: mobileTheme.radius.md,
    borderWidth: 1,
    gap: mobileTheme.spacing.md,
    padding: mobileTheme.spacing.lg,
  },
  card_danger: {
    backgroundColor: mobileTheme.colors.softRed,
    borderColor: hexToRgba(mobileTheme.colors.danger, 0.22),
  },
  card_green: {
    backgroundColor: mobileTheme.colors.softGreen,
    borderColor: hexToRgba(mobileTheme.colors.green, 0.18),
  },
  card_navy: {
    backgroundColor: mobileTheme.colors.navy,
  },
  card_plain: {},
  disabled: {
    opacity: 0.58,
  },
  field: {
    gap: 6,
  },
  heading: {
    gap: 4,
  },
  input: {
    backgroundColor: mobileTheme.colors.white,
    borderColor: mobileTheme.colors.border,
    borderRadius: mobileTheme.radius.md,
    borderWidth: 1,
    color: mobileTheme.colors.ink,
    fontFamily: mobileTheme.font.regular,
    fontSize: 15,
    minHeight: 44,
    paddingHorizontal: mobileTheme.spacing.md,
  },
  inputMultiline: {
    minHeight: 96,
    paddingTop: mobileTheme.spacing.md,
  },
  kicker: {
    color: mobileTheme.colors.green,
    fontFamily: mobileTheme.font.semibold,
    fontSize: 12,
    textTransform: "uppercase",
  },
  label: {
    color: mobileTheme.colors.muted,
    fontFamily: mobileTheme.font.semibold,
    fontSize: 12,
  },
  metric: {
    backgroundColor: mobileTheme.colors.softBlue,
    borderRadius: mobileTheme.radius.md,
    flex: 1,
    gap: 2,
    padding: mobileTheme.spacing.md,
  },
  metricLabel: {
    color: mobileTheme.colors.muted,
    fontFamily: mobileTheme.font.regular,
    fontSize: 12,
  },
  metricValue: {
    color: mobileTheme.colors.navy,
    fontFamily: mobileTheme.font.bold,
    fontSize: 22,
  },
  pill: {
    alignSelf: "flex-start",
    borderRadius: 999,
    paddingHorizontal: 10,
    paddingVertical: 5,
  },
  pill_danger: {
    backgroundColor: mobileTheme.colors.softRed,
    borderColor: hexToRgba(mobileTheme.colors.danger, 0.25),
    borderWidth: 1,
  },
  pill_gold: {
    backgroundColor: mobileTheme.colors.softGold,
    borderColor: hexToRgba(mobileTheme.colors.gold, 0.35),
    borderWidth: 1,
  },
  pill_green: {
    backgroundColor: mobileTheme.colors.softGreen,
    borderColor: hexToRgba(mobileTheme.colors.green, 0.25),
    borderWidth: 1,
  },
  pill_navy: {
    backgroundColor: mobileTheme.colors.navy,
  },
  pillText: {
    fontFamily: mobileTheme.font.semibold,
    fontSize: 12,
  },
  pillText_danger: {
    color: mobileTheme.colors.danger,
  },
  pillText_gold: {
    color: mobileTheme.colors.ink,
  },
  pillText_green: {
    color: mobileTheme.colors.green,
  },
  pillText_navy: {
    color: mobileTheme.colors.white,
  },
  segmented: {
    backgroundColor: mobileTheme.colors.white,
    borderColor: mobileTheme.colors.border,
    borderRadius: mobileTheme.radius.md,
    borderWidth: 1,
    flexDirection: "row",
    padding: 3,
  },
  segment: {
    alignItems: "center",
    borderRadius: 6,
    flex: 1,
    justifyContent: "center",
    minHeight: 44,
    paddingVertical: 9,
  },
  segmentActive: {
    backgroundColor: mobileTheme.colors.navy,
  },
  segmentText: {
    color: mobileTheme.colors.muted,
    fontFamily: mobileTheme.font.semibold,
    fontSize: 12,
    textAlign: "center",
  },
  segmentTextActive: {
    color: mobileTheme.colors.white,
  },
  title: {
    color: mobileTheme.colors.navy,
    fontFamily: mobileTheme.font.bold,
    fontSize: 24,
  },
});

export function EmptyState({
  icon = "inbox",
  title,
  description,
}: {
  icon?: string;
  title: string;
  description?: string;
}) {
  return (
    <View style={styles.emptyState}>
      <View style={styles.emptyIcon}>
        <Feather color={mobileTheme.colors.navy} name={icon} size={26} />
      </View>
      <Text style={styles.emptyTitle}>{title}</Text>
      {description ? (
        <Text style={styles.emptyDescription}>{description}</Text>
      ) : null}
    </View>
  );
}

export const uiStyles = styles;
