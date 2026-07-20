import Feather from "@expo/vector-icons/Feather";
import { Pressable, StyleSheet, Text, TextInput, View } from "react-native";
import { hazardBadgeFor, severityBadgeFor } from "@nadaa/brand/native";
import { mobileTheme, withAlpha } from "../app/theme";

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
    tone === "plain" ? mobileTheme.colors.navy : mobileTheme.colors.white;
  return (
    <Pressable
      accessibilityLabel={label}
      disabled={disabled}
      onPress={onPress}
      style={[
        styles.actionButton,
        styles[`button_${tone}`],
        disabled ? styles.disabled : null,
      ]}
    >
      {icon ? <Feather color={color} name={icon} size={17} /> : null}
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
  secure = false,
  value,
}: {
  label: string;
  multiline?: boolean;
  onChangeText: (value: string) => void;
  placeholder?: string;
  /** Masks entry and opts out of autofill — use for passwords and MFA codes. */
  secure?: boolean;
  value: string;
}) {
  return (
    <View style={styles.field}>
      <Text style={styles.label}>{label}</Text>
      <TextInput
        autoComplete={secure ? "off" : undefined}
        multiline={multiline}
        onChangeText={onChangeText}
        placeholder={placeholder}
        placeholderTextColor={mobileTheme.colors.muted}
        secureTextEntry={secure}
        style={[styles.input, multiline ? styles.inputMultiline : null]}
        textContentType={secure ? "password" : undefined}
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

/** Accessible status pill with WCAG 2.1 AA contrast. */
export function StatusPill({
  label,
  tone = "navy",
}: {
  label: string;
  tone?: "danger" | "gold" | "green" | "navy";
}) {
  return (
    <View style={[styles.pill, styles[`pill_${tone}`]]}>
      <Text style={[styles.pillText, styles[`pillText_${tone}`]]}>{label}</Text>
    </View>
  );
}

const severityIconMap: Record<string, string> = {
  AlertOctagon: "alert-octagon",
  AlertTriangle: "alert-triangle",
  CheckCircle2: "check-circle",
  Info: "info",
};

function severityIcon(name: string): string {
  return severityIconMap[name] ?? "alert-circle";
}

/** Accessible severity badge: icon + text + color. */
export function SeverityBadge({ severity }: { severity: string }) {
  const badge = severityBadgeFor(severity);
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
      <Feather color={badge.color} name={severityIcon(badge.icon)} size={14} />
      <Text style={[styles.badgeText, { color: badge.color }]}>{severity}</Text>
    </View>
  );
}

const hazardIconMap: Record<string, string> = {
  blocked_drain: "x-octagon",
  building_collapse: "home",
  disease_outbreak: "activity",
  electrical_hazard: "zap",
  fire: "sun",
  flood: "cloud-rain",
  landslide: "alert-triangle",
  marine_accident: "anchor",
  medical_emergency: "plus",
  other: "help-circle",
  road_crash: "truck",
  security_incident: "shield",
  storm: "cloud-lightning",
  tidal_wave: "wind",
};

function hazardIcon(hazard: string): string {
  return hazardIconMap[hazard.toLowerCase()] ?? "alert-circle";
}

function formatHazardLabel(hazard: string): string {
  return hazard
    .replace(/_/g, " ")
    .replace(/\b\w/g, (char) => char.toUpperCase());
}

/** Accessible hazard badge: icon + text + color. */
export function HazardBadge({ hazard }: { hazard: string }) {
  const badge = hazardBadgeFor(hazard);
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
      <Feather color={badge.color} name={hazardIcon(hazard)} size={14} />
      <Text style={[styles.badgeText, { color: badge.color }]}>
        {formatHazardLabel(hazard)}
      </Text>
    </View>
  );
}

const urgencySeverityMap: Record<string, string> = {
  high: "high",
  life_threatening: "severe",
  low: "low",
  moderate: "medium",
};

function formatUrgencyLabel(urgency: string): string {
  return urgency
    .replace(/_/g, " ")
    .replace(/\b\w/g, (char) => char.toUpperCase());
}

/** Accessible urgency badge mapped to severity tokens. */
export function UrgencyBadge({ urgency }: { urgency: string }) {
  const badge = severityBadgeFor(urgencySeverityMap[urgency] ?? urgency);
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
      <Feather color={badge.color} name={severityIcon(badge.icon)} size={14} />
      <Text style={[styles.badgeText, { color: badge.color }]}>
        {formatUrgencyLabel(urgency)}
      </Text>
    </View>
  );
}

const capacitySeverityMap: Record<string, string> = {
  available: "low",
  full: "severe",
  limited: "medium",
};

/** Accessible hospital capacity badge mapped to severity tokens. */
export function CapacityBadge({ capacity }: { capacity: string }) {
  const badge = severityBadgeFor(capacitySeverityMap[capacity] ?? capacity);
  const label = capacity.charAt(0).toUpperCase() + capacity.slice(1);
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
      <Feather color={badge.color} name={severityIcon(badge.icon)} size={14} />
      <Text style={[styles.badgeText, { color: badge.color }]}>{label}</Text>
    </View>
  );
}

export function SelectField<Option extends string>({
  label,
  onChange,
  options,
  value,
}: {
  label: string;
  onChange: (value: Option) => void;
  options: Array<{ label: string; value: Option }>;
  value: Option;
}) {
  return (
    <View style={styles.field}>
      <Text style={styles.label}>{label}</Text>
      <View style={styles.selectGrid}>
        {options.map((option) => {
          const selected = option.value === value;
          return (
            <Pressable
              accessibilityLabel={option.label}
              accessibilityRole="button"
              key={option.value}
              onPress={() => onChange(option.value)}
              style={[
                styles.selectOption,
                selected ? styles.selectOptionActive : null,
              ]}
            >
              <Text
                style={[
                  styles.selectOptionText,
                  selected ? styles.selectOptionTextActive : null,
                ]}
              >
                {option.label}
              </Text>
            </Pressable>
          );
        })}
      </View>
    </View>
  );
}

export function ListItem({
  children,
  onPress,
  selected,
}: {
  children: unknown;
  onPress?: () => void;
  selected?: boolean;
}) {
  return (
    <Pressable
      accessibilityRole="button"
      onPress={onPress}
      style={[styles.listItem, selected ? styles.listItemSelected : null]}
    >
      {children}
    </Pressable>
  );
}

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
    backgroundColor: withAlpha(mobileTheme.colors.navy, 0.08),
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
    gap: mobileTheme.spacing.sm,
    paddingHorizontal: mobileTheme.spacing.md,
    paddingVertical: mobileTheme.spacing.sm,
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
    borderColor: withAlpha(mobileTheme.colors.danger, 0.22),
  },
  card_green: {
    backgroundColor: mobileTheme.colors.softGreen,
    borderColor: withAlpha(mobileTheme.colors.green, 0.18),
  },
  card_navy: {
    backgroundColor: mobileTheme.colors.navy,
  },
  card_plain: {},
  disabled: {
    opacity: 0.58,
  },
  field: {
    gap: mobileTheme.spacing.sm,
  },
  heading: {
    gap: mobileTheme.spacing.sm,
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
  listItem: {
    backgroundColor: mobileTheme.colors.white,
    borderColor: mobileTheme.colors.border,
    borderRadius: mobileTheme.radius.md,
    borderWidth: 1,
    gap: mobileTheme.spacing.sm,
    minHeight: 44,
    padding: mobileTheme.spacing.md,
  },
  listItemSelected: {
    backgroundColor: mobileTheme.colors.softBlue,
    borderColor: mobileTheme.colors.navy,
  },
  metric: {
    backgroundColor: mobileTheme.colors.softBlue,
    borderRadius: mobileTheme.radius.md,
    flex: 1,
    gap: mobileTheme.spacing.sm,
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
    backgroundColor: mobileTheme.colors.danger,
  },
  pill_gold: {
    backgroundColor: mobileTheme.colors.gold,
  },
  pill_green: {
    backgroundColor: mobileTheme.colors.green,
  },
  pill_navy: {
    backgroundColor: mobileTheme.colors.navy,
  },
  pillText: {
    fontFamily: mobileTheme.font.semibold,
    fontSize: 12,
  },
  pillText_danger: {
    color: mobileTheme.colors.white,
  },
  pillText_gold: {
    color: mobileTheme.colors.ink,
  },
  pillText_green: {
    color: mobileTheme.colors.white,
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
    padding: mobileTheme.spacing.sm,
  },
  segment: {
    alignItems: "center",
    borderRadius: mobileTheme.radius.sm,
    flex: 1,
    justifyContent: "center",
    minHeight: 44,
    paddingVertical: mobileTheme.spacing.sm,
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
  selectGrid: {
    flexDirection: "row",
    flexWrap: "wrap",
    gap: mobileTheme.spacing.sm,
  },
  selectOption: {
    alignItems: "center",
    backgroundColor: mobileTheme.colors.white,
    borderColor: mobileTheme.colors.border,
    borderRadius: mobileTheme.radius.md,
    borderWidth: 1,
    justifyContent: "center",
    minHeight: 44,
    paddingHorizontal: mobileTheme.spacing.md,
    paddingVertical: mobileTheme.spacing.sm,
  },
  selectOptionActive: {
    backgroundColor: mobileTheme.colors.navy,
    borderColor: mobileTheme.colors.navy,
  },
  selectOptionText: {
    color: mobileTheme.colors.ink,
    fontFamily: mobileTheme.font.semibold,
    fontSize: 13,
  },
  selectOptionTextActive: {
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
