import { useEffect, useState } from "react";
import { usePrefersReducedMotion } from "../hooks";

type SplitRevealProps = {
  text: string;
  mode?: "words" | "letters";
  className?: string;
  /** Delay between consecutive parts. */
  stepMs?: number;
  /** Delay before the first part. */
  delayMs?: number;
};

/**
 * One-shot masked rise reveal that splits text into words or letters.
 * Screen readers get the plain string via aria-label; split spans are
 * aria-hidden. Renders static text under reduced-motion.
 */
export function SplitReveal({
  text,
  mode = "words",
  className,
  stepMs = 55,
  delayMs = 0,
}: SplitRevealProps) {
  const reducedMotion = usePrefersReducedMotion();
  const [started, setStarted] = useState(reducedMotion);

  useEffect(() => {
    if (started) {
      return;
    }
    const frame = requestAnimationFrame(() => setStarted(true));
    return () => cancelAnimationFrame(frame);
  }, [started]);

  if (reducedMotion) {
    return <span className={className}>{text}</span>;
  }

  const parts = mode === "letters" ? Array.from(text) : text.split(" ");
  return (
    <span aria-label={text} className={className} role="text">
      {parts.map((part, index) => (
        <span
          aria-hidden="true"
          className="split-reveal__mask"
          key={`${part}-${index}`}
        >
          <span
            className={`split-reveal__part${started ? " is-in" : ""}`}
            style={{ transitionDelay: `${delayMs + index * stepMs}ms` }}
          >
            {part === " " ? " " : part}
            {mode === "words" && index < parts.length - 1 ? " " : ""}
          </span>
        </span>
      ))}
    </span>
  );
}
