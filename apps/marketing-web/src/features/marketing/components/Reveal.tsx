import type { CSSProperties, ReactNode } from "react";
import { useInView } from "../hooks";

type RevealVariant = "up" | "3d" | "scale";

type RevealProps = {
  children: ReactNode;
  variant?: RevealVariant;
  delay?: number;
  className?: string;
  style?: CSSProperties;
};

/** Wraps content and reveals it (fade + move, 3D tilt, or scale) on scroll. */
export function Reveal({
  children,
  variant = "up",
  delay = 0,
  className,
  style: styleProp,
}: RevealProps) {
  const { ref, inView } = useInView<HTMLDivElement>();
  const style: CSSProperties | undefined =
    delay > 0 || styleProp
      ? { ...styleProp, ...(delay > 0 ? { transitionDelay: `${delay}ms` } : {}) }
      : undefined;
  const classes = [
    "reveal",
    `reveal--${variant}`,
    inView ? "is-visible" : "",
    className ?? "",
  ]
    .filter(Boolean)
    .join(" ");

  return (
    <div className={classes} ref={ref} style={style}>
      {children}
    </div>
  );
}
