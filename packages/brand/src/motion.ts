/**
 * Motion tokens, per the Brand, Look & Feel and Motion guide §7.
 *
 * The guide's rule is that motion explains state, direction, acknowledgement or
 * progress — never spectacle — and that durations are centralised rather than
 * hard-coded per component. Today they are scattered: the marketing stylesheet
 * alone carries 260ms, 360ms, 200ms, 150ms, 0.65s and more, chosen per rule.
 *
 * Emergency surfaces deliberately use the shorter end. During an incident,
 * motion should resolve faster and flourish less.
 *
 * Every value here has a reduced-motion path at the call site; the guide makes
 * that non-optional, and `prefersReducedMotion` below is the shared check.
 */

export const duration = {
  /** 80-120ms — press feedback, tiny state changes. */
  instant: "120ms",
  /** 160-220ms — menus, chips, toggles, alerts entering. */
  fast: "200ms",
  /** 260-360ms — cards, sheets, route-state transitions. */
  standard: "300ms",
  /** 500-700ms — hero and section reveals, major context changes. */
  slow: "650ms",
  /** 900-1400ms — logo reveal and cinematic brand moments ONLY. */
  brand: "1200ms",
} as const;

export const easing = {
  /** Entrances. Decelerating, no overshoot. */
  entrance: "cubic-bezier(0.22, 1, 0.36, 1)",
  /** Exits. Accelerating away. */
  exit: "cubic-bezier(0.4, 0, 1, 1)",
  /**
   * Critical emergency confirmations. Linear-out, deliberately without spring
   * overshoot: a confirmation that bounces reads as undecided.
   */
  confirm: "cubic-bezier(0.4, 0, 0.2, 1)",
} as const;

export type MotionDuration = keyof typeof duration;
export type MotionEasing = keyof typeof easing;

/**
 * True when the user asked the OS to reduce motion. Returns false outside a
 * browser so server rendering picks the still path rather than throwing.
 */
export function prefersReducedMotion(): boolean {
  if (typeof window === "undefined" || !window.matchMedia) {
    return false;
  }
  return window.matchMedia("(prefers-reduced-motion: reduce)").matches;
}
