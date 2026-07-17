import { useEffect, useState } from "react";
import { usePrefersReducedMotion } from "../hooks";

type RotatingWordsProps = {
  words: readonly string[];
  /** Time each word stays visible. */
  intervalMs?: number;
  className?: string;
};

/**
 * Cycles words with a masked vertical rise. The full word list is exposed
 * to screen readers statically; the animated viewport is aria-hidden.
 */
export function RotatingWords({
  words,
  intervalMs = 2600,
  className,
}: RotatingWordsProps) {
  const reducedMotion = usePrefersReducedMotion();
  const [index, setIndex] = useState(0);

  useEffect(() => {
    if (reducedMotion || words.length < 2) {
      return;
    }
    const timer = window.setInterval(
      () => setIndex((current) => (current + 1) % words.length),
      intervalMs,
    );
    return () => window.clearInterval(timer);
  }, [words.length, intervalMs, reducedMotion]);

  return (
    <span className={`rotating-words${className ? ` ${className}` : ""}`}>
      <span className="sr-only">{words.join(", ")}</span>
      <span aria-hidden="true" className="rotating-words__viewport">
        <span className="rotating-words__word" key={index}>
          {words[index]}
        </span>
      </span>
    </span>
  );
}
