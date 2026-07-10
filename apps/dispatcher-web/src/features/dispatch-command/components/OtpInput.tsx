import { styled } from "@mui/material/styles";
import {
  type ClipboardEvent,
  type KeyboardEvent,
  useRef,
} from "react";

type OtpInputProps = {
  /** The full code so far (digits only). */
  value: string;
  onChange: (value: string) => void;
  length?: number;
  autoFocus?: boolean;
  /** Fired once every box is filled. */
  onComplete?: (value: string) => void;
  ariaDescribedBy?: string;
};

const DigitBox = styled("input")({
  flex: "1 1 0",
  minWidth: 0,
  width: "100%",
  height: 58,
  padding: 0,
  textAlign: "center",
  fontSize: "1.6rem",
  fontWeight: 700,
  lineHeight: 1,
  fontFamily: "inherit",
  color: "var(--nadaa-navy, #0d1b3d)",
  background: "var(--nadaa-white, #ffffff)",
  border:
    "1.5px solid color-mix(in srgb, var(--nadaa-navy, #0d1b3d) 24%, transparent)",
  borderRadius: 12,
  outline: "none",
  transition: "border-color 150ms ease, box-shadow 150ms ease",
  appearance: "textfield",
  "&::-webkit-outer-spin-button, &::-webkit-inner-spin-button": {
    WebkitAppearance: "none",
    margin: 0,
  },
  "&:hover": {
    borderColor: "color-mix(in srgb, var(--nadaa-navy, #0d1b3d) 40%, transparent)",
  },
  "&:focus": {
    borderColor: "var(--nadaa-gold, #f4c20d)",
    boxShadow:
      "0 0 0 3px color-mix(in srgb, var(--nadaa-gold, #f4c20d) 42%, transparent)",
  },
  "@media (prefers-reduced-motion: reduce)": {
    transition: "none",
  },
});

/**
 * Segmented one-time-code entry: one box per digit with auto-advance,
 * backspace-to-previous, arrow-key movement, and full-code paste. The joined
 * value is reported through onChange so callers keep treating it as a string.
 */
export function OtpInput({
  value,
  onChange,
  length = 6,
  autoFocus = false,
  onComplete,
  ariaDescribedBy,
}: OtpInputProps) {
  const boxes = useRef<Array<HTMLInputElement | null>>([]);
  const digits = Array.from({ length }, (_, index) => value[index] ?? "");

  const focusBox = (index: number) => {
    const target = boxes.current[Math.max(0, Math.min(length - 1, index))];
    target?.focus();
    target?.select();
  };

  const commit = (next: string[]) => {
    const joined = next.join("").slice(0, length);
    onChange(joined);
    if (joined.length === length) {
      onComplete?.(joined);
    }
  };

  const handleChange = (index: number, raw: string) => {
    const cleaned = raw.replace(/\D/g, "");
    const next = digits.slice();
    if (!cleaned) {
      next[index] = "";
      commit(next);
      return;
    }
    // Spread typed/overtyped characters across this box and the ones after it.
    let cursor = index;
    for (const character of cleaned) {
      if (cursor >= length) {
        break;
      }
      next[cursor] = character;
      cursor += 1;
    }
    commit(next);
    focusBox(cursor);
  };

  const handleKeyDown = (index: number, event: KeyboardEvent<HTMLInputElement>) => {
    if (event.key === "Backspace") {
      event.preventDefault();
      const next = digits.slice();
      if (next[index]) {
        next[index] = "";
        commit(next);
      } else if (index > 0) {
        next[index - 1] = "";
        commit(next);
        focusBox(index - 1);
      }
    } else if (event.key === "ArrowLeft") {
      event.preventDefault();
      focusBox(index - 1);
    } else if (event.key === "ArrowRight") {
      event.preventDefault();
      focusBox(index + 1);
    }
  };

  const handlePaste = (event: ClipboardEvent<HTMLInputElement>) => {
    event.preventDefault();
    const pasted = event.clipboardData
      .getData("text")
      .replace(/\D/g, "")
      .slice(0, length);
    if (!pasted) {
      return;
    }
    commit(pasted.split(""));
    focusBox(pasted.length);
  };

  return (
    <div
      aria-label={`${length}-digit authenticator code`}
      role="group"
      style={{ display: "flex", gap: 10 }}
    >
      {digits.map((digit, index) => (
        <DigitBox
          aria-describedby={ariaDescribedBy}
          aria-label={`Digit ${index + 1} of ${length}`}
          autoComplete={index === 0 ? "one-time-code" : "off"}
          autoFocus={autoFocus && index === 0}
          inputMode="numeric"
          // eslint-disable-next-line react/no-array-index-key
          key={index}
          maxLength={1}
          onChange={(event) => handleChange(index, event.target.value)}
          onFocus={(event) => event.currentTarget.select()}
          onKeyDown={(event) => handleKeyDown(index, event)}
          onPaste={handlePaste}
          ref={(element) => {
            boxes.current[index] = element;
          }}
          type="text"
          value={digit}
        />
      ))}
    </div>
  );
}
