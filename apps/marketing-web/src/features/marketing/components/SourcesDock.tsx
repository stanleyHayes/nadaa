import { BookOpenText, ChevronDown, ExternalLink } from "lucide-react";
import { useState } from "react";
import { researchNotes } from "../data";

/**
 * Evidence & sources, docked at the foot of the Trust page. Collapsed by
 * default behind a single action button so the references don't sprawl as a
 * card grid; expanding reveals a compact, scannable source list.
 */
export function SourcesDock() {
  const [open, setOpen] = useState(false);

  return (
    <section aria-label="Evidence and sources" className="sources-dock">
      <div className="sources-dock__bar">
        <div className="sources-dock__intro">
          <BookOpenText aria-hidden="true" size={22} />
          <div>
            <p className="eyebrow">Evidence &amp; sources</p>
            <p className="sources-dock__lead">
              Every claim on this page is grounded in public record.
            </p>
          </div>
        </div>
        <button
          aria-controls="sources-dock-panel"
          aria-expanded={open}
          className="sources-dock__toggle"
          onClick={() => setOpen((value) => !value)}
          type="button"
        >
          {open ? "Hide sources" : `Review the ${researchNotes.length} sources`}
          <ChevronDown
            aria-hidden="true"
            className={open ? "is-open" : undefined}
            size={16}
          />
        </button>
      </div>
      <div
        className={`sources-dock__panel${open ? " is-open" : ""}`}
        id="sources-dock-panel"
      >
        <div className="sources-dock__panel-inner">
          <ul className="sources-dock__list">
            {researchNotes.map((note) => (
              <li key={note.title}>
                <div>
                  <h3>{note.title}</h3>
                  <p>{note.body}</p>
                </div>
                <a href={note.href} rel="noreferrer" target="_blank">
                  <ExternalLink aria-hidden="true" size={14} />
                  <span>{note.source}</span>
                </a>
              </li>
            ))}
          </ul>
        </div>
      </div>
    </section>
  );
}
