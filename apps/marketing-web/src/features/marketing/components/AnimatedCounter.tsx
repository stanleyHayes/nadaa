import { useCountUp, useInView } from "../hooks";

type AnimatedCounterProps = {
  value: number;
  prefix?: string;
  suffix?: string;
};

/** Counts up to `value` the first time it scrolls into view. */
export function AnimatedCounter({
  value,
  prefix = "",
  suffix = "",
}: AnimatedCounterProps) {
  const { ref, inView } = useInView<HTMLSpanElement>();
  const current = useCountUp(value, inView);
  return (
    <span ref={ref}>
      {prefix}
      {current}
      {suffix}
    </span>
  );
}
