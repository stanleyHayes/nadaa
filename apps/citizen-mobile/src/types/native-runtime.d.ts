declare namespace JSX {
  interface IntrinsicAttributes {
    key?: string;
  }

  interface IntrinsicElements {
    [elementName: string]: unknown;
  }
}

declare const process: {
  env: Record<string, string | undefined>;
};

declare function require(path: string): unknown;

declare module "react" {
  export type ReactNode = unknown;
  export type ComponentType<Props = Record<string, unknown>> = (
    props: Props & { key?: string },
  ) => ReactNode;
  export function useEffect(
    effect: () => void | (() => void) | Promise<void>,
    deps?: unknown[],
  ): void;
  export function useMemo<Value>(factory: () => Value, deps?: unknown[]): Value;
  export function useState<Value>(
    initial: Value | (() => Value),
  ): [Value, (next: Value | ((current: Value) => Value)) => void];
  export function useRef<Value>(initial: Value): { current: Value };
}

declare module "react/jsx-runtime" {
  export const Fragment: unknown;
  export function jsx(type: unknown, props: unknown, key?: unknown): unknown;
  export function jsxs(type: unknown, props: unknown, key?: unknown): unknown;
}

declare module "react-native" {
  import type { ComponentType, ReactNode } from "react";

  export type ViewStyle = Record<string, unknown>;
  export type TextStyle = Record<string, unknown>;
  export type ImageStyle = Record<string, unknown>;
  export type StyleProp<T> = T | T[] | null | undefined | unknown;
  export const Platform: {
    OS: "android" | "ios" | "web" | string;
    select: <T>(spec: Record<string, T>) => T | undefined;
  };

  export const SafeAreaView: ComponentType<{
    children?: ReactNode;
    style?: StyleProp<ViewStyle>;
    [key: string]: unknown;
  }>;
  export const ScrollView: ComponentType<{
    children?: ReactNode;
    contentContainerStyle?: StyleProp<ViewStyle>;
    style?: StyleProp<ViewStyle>;
    [key: string]: unknown;
  }>;
  export const StatusBar: ComponentType<{
    barStyle?: "default" | "light-content" | "dark-content";
    backgroundColor?: string;
    [key: string]: unknown;
  }>;
  export const Switch: ComponentType<{
    value: boolean;
    onValueChange?: (value: boolean) => void;
    trackColor?: Record<string, string>;
    thumbColor?: string;
    [key: string]: unknown;
  }>;
  export const Text: ComponentType<{
    children?: ReactNode;
    numberOfLines?: number;
    style?: StyleProp<TextStyle>;
    [key: string]: unknown;
  }>;
  export const TextInput: ComponentType<{
    multiline?: boolean;
    keyboardType?: string;
    onChangeText?: (value: string) => void;
    placeholder?: string;
    placeholderTextColor?: string;
    style?: StyleProp<TextStyle>;
    value?: string;
    [key: string]: unknown;
  }>;
  export const Pressable: ComponentType<{
    accessibilityLabel?: string;
    children?: ReactNode;
    disabled?: boolean;
    onPress?: () => void;
    style?: StyleProp<ViewStyle>;
    [key: string]: unknown;
  }>;
  export const View: ComponentType<{
    children?: ReactNode;
    style?: StyleProp<ViewStyle>;
    [key: string]: unknown;
  }>;
  export const Image: ComponentType<{
    accessibilityLabel?: string;
    source?: unknown;
    style?: StyleProp<ImageStyle>;
    [key: string]: unknown;
  }>;
  export const StyleSheet: {
    create<Styles extends Record<string, ViewStyle | TextStyle | ImageStyle>>(
      styles: Styles,
    ): Styles;
  };
}

declare module "@expo/vector-icons/Feather" {
  import type { ComponentType } from "react";

  const Feather: ComponentType<{
    color?: string;
    name: string;
    size?: number;
  }>;
  export default Feather;
}

declare module "expo-notifications" {
  export enum AndroidImportance {
    MIN = 1,
    LOW = 2,
    DEFAULT = 3,
    HIGH = 4,
    MAX = 5,
  }
  export enum AndroidNotificationVisibility {
    SECRET = -1,
    PRIVATE = 0,
    PUBLIC = 1,
  }
  export type NotificationPermissionsStatus = {
    granted: boolean;
    status: "granted" | "denied" | "undetermined";
  };
  export function setNotificationHandler(handler: {
    handleNotification: () => Promise<{
      shouldShowAlert?: boolean;
      shouldShowBanner?: boolean;
      shouldShowList?: boolean;
      shouldPlaySound?: boolean;
      shouldSetBadge?: boolean;
    }>;
  }): void;
  export function getPermissionsAsync(): Promise<NotificationPermissionsStatus>;
  export function requestPermissionsAsync(request?: {
    ios?: {
      allowAlert?: boolean;
      allowSound?: boolean;
      allowBadge?: boolean;
      allowCriticalAlerts?: boolean;
    };
  }): Promise<NotificationPermissionsStatus>;
  export function setNotificationChannelAsync(
    channelId: string,
    channel: {
      name: string;
      importance: AndroidImportance;
      sound?: string | null;
      bypassDnd?: boolean;
      lockscreenVisibility?: AndroidNotificationVisibility;
      vibrationPattern?: number[];
      lightColor?: string;
      enableVibrate?: boolean;
    },
  ): Promise<unknown>;
  export function scheduleNotificationAsync(request: {
    content: {
      title: string;
      body?: string;
      sound?: boolean | string;
      badge?: number;
      interruptionLevel?: "active" | "critical" | "passive" | "timeSensitive";
      data?: Record<string, unknown>;
    };
    trigger: null | { channelId?: string };
  }): Promise<string>;
}
