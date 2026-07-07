import Feather from "@expo/vector-icons/Feather";
import { Pressable, StyleSheet, Text, TextInput, View } from "react-native";
import { mobileTheme } from "../app/theme";

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
        placeholderTextColor="rgba(85, 91, 102, 0.72)"
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
  return (
    <View style={[styles.pill, styles[`pill_${tone}`]]}>
      <Text style={styles.pillText}>{label}</Text>
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
      onPress={onPress}
      style={[styles.listItem, selected ? styles.listItemSelected : null]}
    >
      {children}
    </Pressable>
  );
}

const styles = StyleSheet.create({
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
    borderColor: "rgba(229, 57, 53, 0.22)",
  },
  card_green: {
    backgroundColor: mobileTheme.colors.softGreen,
    borderColor: "rgba(17, 141, 78, 0.18)",
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
  listItem: {
    backgroundColor: mobileTheme.colors.white,
    borderColor: mobileTheme.colors.border,
    borderRadius: mobileTheme.radius.md,
    borderWidth: 1,
    gap: mobileTheme.spacing.sm,
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
    color: mobileTheme.colors.white,
    fontFamily: mobileTheme.font.semibold,
    fontSize: 12,
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
    borderRadius: 6,
    flex: 1,
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
  selectGrid: {
    flexDirection: "row",
    flexWrap: "wrap",
    gap: mobileTheme.spacing.sm,
  },
  selectOption: {
    backgroundColor: mobileTheme.colors.white,
    borderColor: mobileTheme.colors.border,
    borderRadius: mobileTheme.radius.md,
    borderWidth: 1,
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

export const uiStyles = styles;
