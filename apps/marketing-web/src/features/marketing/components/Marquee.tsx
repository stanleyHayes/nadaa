type MarqueeProps = {
  items: readonly string[];
  className?: string;
  ariaLabel?: string;
};

/**
 * Slow infinite keyword marquee with edge fade masks; the track is
 * duplicated for a seamless loop and pauses on hover. Screen readers get
 * the plain item list; the animated track is aria-hidden.
 */
export function Marquee({ items, className, ariaLabel }: MarqueeProps) {
  const loop = [...items, ...items];
  return (
    <div
      aria-label={ariaLabel ?? items.join(", ")}
      className={`marquee${className ? ` ${className}` : ""}`}
      role="note"
    >
      <ul aria-hidden="true" className="marquee__track">
        {loop.map((item, index) => (
          <li className="marquee__item" key={`${item}-${index}`}>
            {item}
          </li>
        ))}
      </ul>
    </div>
  );
}
