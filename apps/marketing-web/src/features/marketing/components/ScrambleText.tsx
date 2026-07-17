import { useEffect, useState } from "react";
import { useInView, usePrefersReducedMotion } from "../hooks";

const GLYPHS = "0123456789#*+";

type ScrambleTextProps = {
  text: string;
  className?: string;
  durationMs?: number;
};

/**
 * Decrypt-style scramble: glyphs churn and settle left-to-right into the
 * real text once the element scrolls into view. Static under reduced-motion.
 */
export function ScrambleText({
  text,
  className,
  durationMs = 900,
}: ScrambleTextProps) {
  const { ref, inView } = useInView<HTMLSpanElement>();
  const reducedMotion = usePrefersReducedMotion();
  const [output, setOutput] = useState(text);

  useEffect(() => {
    if (reducedMotion || !inView) {
      setOutput(text);
      return;
    }
    let frame = 0;
    let start = 0;
    const tick = (timestamp: number) => {
      if (!start) {
        start = timestamp;
      }
      const progress = Math.min(1, (timestamp - start) / durationMs);
      const settled = Math.floor(progress * text.length);
      setOutput(
        text
          .split("")
          .map((char, index) =>
            index < settled || char === " "
              ? char
              : GLYPHS[Math.floor(Math.random() * GLYPHS.length)],
          )
          .join(""),
      );
      if (progress < 1) {
        frame = requestAnimationFrame(tick);
      }
    };
    frame = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(frame);
  }, [inView, reducedMotion, text, durationMs]);

  return (
    <span className={className} ref={ref}>
      {output}
    </span>
  );
}
