import { useEffect, useRef } from "react";
import { usePrefersReducedMotion } from "../hooks";

/**
 * Soft brand-tint glow that trails the cursor inside its parent section
 * (hero). Renders nothing on touch devices or under reduced-motion; the
 * parent must be positioned (hero-section already is).
 */
export function CursorSpotlight() {
  const ref = useRef<HTMLDivElement>(null);
  const reducedMotion = usePrefersReducedMotion();

  useEffect(() => {
    const element = ref.current;
    const parent = element?.parentElement;
    if (
      !element ||
      !parent ||
      reducedMotion ||
      typeof window.matchMedia !== "function" ||
      !window.matchMedia("(hover: hover) and (pointer: fine)").matches
    ) {
      return;
    }

    let frame = 0;
    let active = false;
    let x = 0;
    let y = 0;
    let targetX = 0;
    let targetY = 0;

    const render = () => {
      x += (targetX - x) * 0.16;
      y += (targetY - y) * 0.16;
      element.style.transform = `translate3d(${x.toFixed(1)}px, ${y.toFixed(1)}px, 0) translate(-50%, -50%)`;
      frame = requestAnimationFrame(render);
    };
    const onPointerMove = (event: PointerEvent) => {
      const rect = parent.getBoundingClientRect();
      targetX = event.clientX - rect.left;
      targetY = event.clientY - rect.top;
      if (!active) {
        active = true;
        x = targetX;
        y = targetY;
        element.classList.add("is-active");
        frame = requestAnimationFrame(render);
      }
    };
    const onPointerLeave = () => {
      active = false;
      element.classList.remove("is-active");
      cancelAnimationFrame(frame);
    };

    parent.addEventListener("pointermove", onPointerMove);
    parent.addEventListener("pointerleave", onPointerLeave);
    return () => {
      parent.removeEventListener("pointermove", onPointerMove);
      parent.removeEventListener("pointerleave", onPointerLeave);
      cancelAnimationFrame(frame);
    };
  }, [reducedMotion]);

  if (reducedMotion) {
    return null;
  }
  return <div aria-hidden="true" className="cursor-spotlight" ref={ref} />;
}
