import { useEffect, useId, useMemo, useState, type MouseEvent } from "react";
import { Button, IconButton, Popover } from "@mui/material";
import { HelpCircle, Square, Volume2 } from "lucide-react";

const SUPPORTS_TTS =
  typeof window !== "undefined" && "speechSynthesis" in window;

/**
 * A "?" button that sits beside a page title and opens a popover with the
 * current page's numbered steps. A Listen control reads the guide aloud with
 * the Web Speech API (en-GB), toggling to Stop while speaking. A visually
 * hidden transcript keeps the steps available to screen readers.
 */
export function PageHelp({
  title,
  description,
  steps,
}: {
  title: string;
  description?: string;
  steps: string[];
}) {
  const [anchor, setAnchor] = useState<null | HTMLElement>(null);
  const [speaking, setSpeaking] = useState(false);
  const open = Boolean(anchor);
  const transcriptId = useId();

  const guideText = useMemo(
    () => [title, description, ...steps].filter(Boolean).join(". "),
    [title, description, steps],
  );

  // Cancel any in-flight speech when this help control unmounts.
  useEffect(() => {
    return () => {
      if (SUPPORTS_TTS) {
        window.speechSynthesis.cancel();
      }
    };
  }, []);

  const stopSpeech = () => {
    if (SUPPORTS_TTS) {
      window.speechSynthesis.cancel();
    }
    setSpeaking(false);
  };

  const playGuide = () => {
    if (!SUPPORTS_TTS) {
      return;
    }
    window.speechSynthesis.cancel();
    const utterance = new SpeechSynthesisUtterance(guideText);
    utterance.lang = "en-GB";
    utterance.onend = () => setSpeaking(false);
    utterance.onerror = () => setSpeaking(false);
    setSpeaking(true);
    window.speechSynthesis.speak(utterance);
  };

  const openHelp = (event: MouseEvent<HTMLElement>) =>
    setAnchor(event.currentTarget);

  const closeHelp = () => {
    stopSpeech();
    setAnchor(null);
  };

  return (
    <>
      <span id={transcriptId} className="cc-sr-only">
        {`How to use ${title}. `}
        {description ? `${description}. ` : ""}
        {steps.map((step, index) => `Step ${index + 1}: ${step}.`).join(" ")}
      </span>
      <IconButton
        className="cc-help__trigger"
        aria-label={`How to use ${title}`}
        aria-describedby={transcriptId}
        aria-haspopup="dialog"
        aria-expanded={open}
        onClick={openHelp}
        data-tour="page-help"
      >
        <HelpCircle size={18} aria-hidden />
      </IconButton>
      <Popover
        open={open}
        anchorEl={anchor}
        onClose={closeHelp}
        anchorOrigin={{ vertical: "bottom", horizontal: "left" }}
        transformOrigin={{ vertical: "top", horizontal: "left" }}
        slotProps={{ paper: { className: "cc-help__panel" } }}
      >
        <p className="cc-help__eyebrow">How to use this page</p>
        <p className="cc-help__title">{title}</p>
        {description ? <p className="cc-help__desc">{description}</p> : null}
        <ol className="cc-help__steps">
          {steps.map((step) => (
            <li key={step}>{step}</li>
          ))}
        </ol>
        {SUPPORTS_TTS ? (
          <div className="cc-help__foot">
            <Button
              size="small"
              variant="outlined"
              className="cc-help__listen"
              startIcon={speaking ? <Square size={15} /> : <Volume2 size={15} />}
              onClick={speaking ? stopSpeech : playGuide}
            >
              {speaking ? "Stop" : "Listen"}
            </Button>
            <span className="cc-help__hint">Reads the steps aloud</span>
          </div>
        ) : null}
      </Popover>
    </>
  );
}
