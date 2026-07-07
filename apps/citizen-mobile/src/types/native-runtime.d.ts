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
