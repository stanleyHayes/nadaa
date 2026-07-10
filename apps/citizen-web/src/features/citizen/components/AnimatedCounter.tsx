import { useCountUp, useInView } from "../hooks";

type AnimatedCounterProps = {
  value: number;
  prefix?: string;
  suffix?: string;
};

/** Counts up to `value` the first time it scrolls into view (tabular-nums). */
export function AnimatedCounter({
  value,
  prefix = "",
  suffix = "",
}: AnimatedCounterProps) {
  const { ref, inView } = useInView<HTMLSpanElement>();
  const current = useCountUp(value, inView);
  return (
    <span className="count-up" ref={ref}>
      {prefix}
      {current}
      {suffix}
    </span>
  );
}

export default AnimatedCounter;
