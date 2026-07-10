import { CheckCircle2 } from "lucide-react";
import { Reveal } from "../components/Reveal";
import { differentiators, platformPositioning, responseLoop } from "../data";

export function HowItWorksPage() {
  return (
    <>
      <section className="page-head">
        <p className="eyebrow">How it works</p>
        <h1>One response loop, end to end.</h1>
        <p>{platformPositioning}</p>
      </section>

      <section className="content-section" aria-label="Response loop">
        <ol className="loop-grid loop-grid--full">
          {responseLoop.map((stage, index) => (
            <Reveal
              className="loop-card"
              delay={(index % 3) * 100}
              key={stage.step}
              variant="up"
            >
              <span aria-hidden="true" className="loop-step">
                {stage.step}
              </span>
              <h3>{stage.title}</h3>
              <p>{stage.description}</p>
            </Reveal>
          ))}
        </ol>
      </section>

      <section
        aria-labelledby="hiw-why"
        className="content-section why-section"
      >
        <div className="section-heading">
          <p className="eyebrow">Why NADAA</p>
          <h2 id="hiw-why">Designed for real emergencies, not demos.</h2>
        </div>
        <div className="why-grid">
          {differentiators.map((item, index) => (
            <Reveal
              className="why-card"
              delay={(index % 3) * 80}
              key={item.title}
              variant="up"
            >
              <CheckCircle2 aria-hidden="true" size={22} />
              <h3>{item.title}</h3>
              <p>{item.description}</p>
            </Reveal>
          ))}
        </div>
      </section>
    </>
  );
}
