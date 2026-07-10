import { useCallback, useEffect, useState } from "react";
import { Button } from "@mui/material";
import { ArrowLeft, ArrowRight, Check, X } from "lucide-react";

type TourStep = {
  selector: string | string[];
  title: string;
  description: string;
};

const TOUR_EVENT = "nadaa:replay-tour";
const COMPLETE_KEY = "nadaa.authority.tour-complete";

/** Fire from anywhere (e.g. the account menu) to restart the coachmark tour. */
export function dispatchReplayTour() {
  window.dispatchEvent(new CustomEvent(TOUR_EVENT));
}

const STEPS: TourStep[] = [
  {
    selector: "[data-tour='sidebar']",
    title: "Command sections",
    description:
      "Move between every workspace here — Overview, Incidents, Alerts, Shelters, and the intelligence and recovery desks. Collapse the rail when you need more room.",
  },
  {
    selector: "[data-tour='page-header']",
    title: "Where you are",
    description:
      "The header names the current workspace and sums up what it is for, so you always know which desk you are working.",
  },
  {
    selector: "[data-tour='page-help']",
    title: "How to use this page",
    description:
      "Open this help button on any page for numbered steps. Press Listen inside it to have the steps read aloud.",
  },
  {
    selector: "[data-tour='topbar-theme']",
    title: "Light and dark",
    description:
      "Switch the operations room between light and dark mode. Your choice is saved on this browser.",
  },
  {
    selector: "[data-tour='topbar-notifications']",
    title: "Operational notices",
    description:
      "High-severity incidents, alerts awaiting approval, and shelters near capacity surface here as you work.",
  },
  {
    selector: "[data-tour='user-menu']",
    title: "Your account",
    description:
      "Reach your profile, settings, the full user guide, and replay this tour from the account menu.",
  },
];

function isVisible(element: Element): element is HTMLElement {
  if (!(element instanceof HTMLElement)) {
    return false;
  }
  const rect = element.getBoundingClientRect();
  const style = window.getComputedStyle(element);
  return rect.width > 0 && rect.height > 0 && style.visibility !== "hidden";
}

function findTarget(selector: string | string[]): HTMLElement | null {
  const selectors = Array.isArray(selector) ? selector : [selector];
  for (const item of selectors) {
    const candidate = Array.from(document.querySelectorAll(item)).find(isVisible);
    if (candidate) {
      return candidate;
    }
  }
  return null;
}

function clamp(value: number, min: number, max: number) {
  return Math.min(max, Math.max(min, value));
}

function prefersReducedMotion() {
  if (typeof window === "undefined") {
    return false;
  }
  if (
    document.documentElement.getAttribute("data-nadaa-reduced-motion") ===
    "reduce"
  ) {
    return true;
  }
  return window.matchMedia("(prefers-reduced-motion: reduce)").matches;
}

/**
 * First-login coachmark tour. Spotlights a sequence of shell elements tagged
 * with `data-tour` attributes, positions an explanatory panel beside each, and
 * supports keyboard navigation (Esc to close, arrows to move). Completion is
 * persisted to localStorage so it auto-starts only once; the account menu can
 * replay it via a `nadaa:replay-tour` event. Honours reduced-motion.
 */
export function AppTour() {
  const total = STEPS.length;
  const [active, setActive] = useState(false);
  const [index, setIndex] = useState(0);
  const [rect, setRect] = useState<DOMRect | null>(null);
  const [reduced, setReduced] = useState(false);

  const current = STEPS[index];

  const start = useCallback(() => {
    setReduced(prefersReducedMotion());
    setIndex(0);
    setActive(true);
  }, []);

  const finish = useCallback((markComplete = true) => {
    setActive(false);
    setRect(null);
    if (markComplete) {
      try {
        window.localStorage.setItem(COMPLETE_KEY, "true");
      } catch {
        /* storage unavailable */
      }
    }
  }, []);

  // Replay when the account menu dispatches the event.
  useEffect(() => {
    const replay = () => start();
    window.addEventListener(TOUR_EVENT, replay);
    return () => window.removeEventListener(TOUR_EVENT, replay);
  }, [start]);

  // Auto-start once, on the first load for this browser.
  useEffect(() => {
    let complete = false;
    try {
      complete = window.localStorage.getItem(COMPLETE_KEY) === "true";
    } catch {
      complete = true;
    }
    if (complete) {
      return;
    }
    const timer = window.setTimeout(start, 600);
    return () => window.clearTimeout(timer);
  }, [start]);

  // Measure the active target and follow it on scroll/resize.
  useEffect(() => {
    if (!active || !current) {
      return;
    }
    const step = current;
    let frame = 0;

    const measure = () => {
      const target = findTarget(step.selector);
      if (!target) {
        // Skip a target that is not present (e.g. hidden on mobile).
        setIndex((value) => (value >= total - 1 ? value : value + 1));
        return;
      }
      target.scrollIntoView({
        block: "nearest",
        inline: "nearest",
        behavior: prefersReducedMotion() ? "auto" : "smooth",
      });
      frame = window.requestAnimationFrame(() => {
        setRect(target.getBoundingClientRect());
      });
    };

    measure();
    window.addEventListener("resize", measure);
    window.addEventListener("scroll", measure, true);
    return () => {
      window.cancelAnimationFrame(frame);
      window.removeEventListener("resize", measure);
      window.removeEventListener("scroll", measure, true);
    };
  }, [active, current, total]);

  // Keyboard navigation.
  useEffect(() => {
    if (!active) {
      return;
    }
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        finish();
      }
      if (event.key === "ArrowRight") {
        setIndex((value) => (value >= total - 1 ? value : value + 1));
      }
      if (event.key === "ArrowLeft") {
        setIndex((value) => Math.max(0, value - 1));
      }
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [active, finish, total]);

  if (!active || !current || !rect) {
    return null;
  }

  const margin = 10;
  const highlight = {
    top: clamp(rect.top - margin, 8, window.innerHeight),
    left: clamp(rect.left - margin, 8, window.innerWidth),
    width: clamp(rect.width + margin * 2, 40, window.innerWidth - 16),
    height: clamp(rect.height + margin * 2, 40, window.innerHeight - 16),
  };
  const panelWidth = Math.min(360, window.innerWidth - 32);
  const below = highlight.top + highlight.height + 16;
  const placeBelow = below + 240 < window.innerHeight;
  const panelTop = placeBelow ? below : Math.max(16, highlight.top - 250);
  const panelLeft = clamp(
    highlight.left + highlight.width / 2 - panelWidth / 2,
    16,
    Math.max(16, window.innerWidth - panelWidth - 16),
  );
  const isLast = index === total - 1;

  return (
    <div
      className="cc-tour"
      role="dialog"
      aria-modal="true"
      aria-label="Authority dashboard tour"
    >
      <div className="cc-tour__scrim" />
      <div
        className={`cc-tour__spot${reduced ? " is-static" : ""}`}
        style={highlight}
        aria-hidden
      />
      <div
        className="cc-tour__panel"
        style={{ width: panelWidth, top: panelTop, left: panelLeft }}
      >
        <div className="cc-tour__head">
          <span className="cc-tour__step">
            Step {index + 1} of {total}
          </span>
          <button
            type="button"
            className="cc-tour__close"
            aria-label="Skip tour"
            onClick={() => finish()}
          >
            <X size={16} aria-hidden />
          </button>
        </div>
        <h2 className="cc-tour__title">{current.title}</h2>
        <p className="cc-tour__body">{current.description}</p>
        <div className="cc-tour__dots" aria-hidden>
          {STEPS.map((step, dot) => (
            <span
              key={step.title}
              className={`cc-tour__dot${dot === index ? " is-active" : ""}`}
            />
          ))}
        </div>
        <div className="cc-tour__actions">
          <Button
            size="small"
            variant="text"
            className="cc-tour__skip"
            onClick={() => finish()}
          >
            Skip
          </Button>
          <div className="cc-tour__nav">
            <Button
              size="small"
              variant="outlined"
              startIcon={<ArrowLeft size={15} />}
              disabled={index === 0}
              onClick={() => setIndex((value) => Math.max(0, value - 1))}
            >
              Back
            </Button>
            {isLast ? (
              <Button
                size="small"
                variant="contained"
                startIcon={<Check size={15} />}
                onClick={() => finish()}
              >
                Done
              </Button>
            ) : (
              <Button
                size="small"
                variant="contained"
                endIcon={<ArrowRight size={15} />}
                onClick={() =>
                  setIndex((value) => Math.min(total - 1, value + 1))
                }
              >
                Next
              </Button>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
