import { useRef, type PointerEvent, type ReactNode } from "react";
import { usePrefersReducedMotion } from "../hooks";

type MagneticProps = {
  children: ReactNode;
  /** Maximum pull in pixels. */
  strength?: number;
  className?: string;
};

const finePointer = () =>
  typeof window !== "undefined" &&
  typeof window.matchMedia === "function" &&
  window.matchMedia("(hover: hover) and (pointer: fine)").matches;

/**
 * Pulls wrapped content a few pixels toward the cursor for a magnetic
 * hover feel. No-op on touch devices and under reduced-motion.
 */
export function Magnetic({ children, strength = 6, className }: MagneticProps) {
  const ref = useRef<HTMLSpanElement>(null);
  const reducedMotion = usePrefersReducedMotion();

  if (reducedMotion || !finePointer()) {
    return <span className={className}>{children}</span>;
  }

  const onPointerMove = (event: PointerEvent<HTMLSpanElement>) => {
    const element = ref.current;
    if (!element) {
      return;
    }
    const rect = element.getBoundingClientRect();
    const relX =
      (event.clientX - (rect.left + rect.width / 2)) / (rect.width / 2);
    const relY =
      (event.clientY - (rect.top + rect.height / 2)) / (rect.height / 2);
    const x = Math.max(-1, Math.min(1, relX)) * strength;
    const y = Math.max(-1, Math.min(1, relY)) * strength;
    element.style.transform = `translate3d(${x.toFixed(1)}px, ${y.toFixed(1)}px, 0)`;
  };

  const onPointerLeave = () => {
    const element = ref.current;
    if (element) {
      element.style.transform = "";
    }
  };

  return (
    <span
      className={`magnetic${className ? ` ${className}` : ""}`}
      onPointerLeave={onPointerLeave}
      onPointerMove={onPointerMove}
      ref={ref}
    >
      {children}
    </span>
  );
}
