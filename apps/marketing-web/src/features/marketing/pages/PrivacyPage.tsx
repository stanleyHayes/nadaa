import { nadaaBrand } from "@nadaa/brand";
import { PageBanner } from "../components/PageBanner";

/**
 * Public privacy notice.
 *
 * Apple and Google both require a reachable privacy policy URL before an app
 * that collects location, camera or account data can be submitted, so this page
 * is a release dependency for the mobile apps, not marketing copy.
 *
 * The data inventory below was derived from what the clients actually transmit,
 * not from intent: the citizen app sends precise location, incident media, a
 * phone number for OTP sign-in, a display name and a session id; the dispatcher
 * app sends only a name and session id. Keep it that way — if a client starts
 * sending something new, this page and the iOS privacy manifests in each app's
 * app.json have to change with it.
 *
 * LEGAL REVIEW REQUIRED before this goes live. Everything marked TO BE
 * CONFIRMED needs the registered entity's details and must be signed off by
 * counsel; Ghana's Data Protection Act, 2012 (Act 843) also requires the
 * controller to be registered with the Data Protection Commission, which is an
 * administrative filing that happens outside this repository.
 */

const PENDING = "[TO BE CONFIRMED — see PrivacyPage.tsx]";

type Section = {
  heading: string;
  body: string[];
  list?: { term: string; detail: string }[];
};

const sections: Section[] = [
  {
    heading: "Who is responsible for your data",
    body: [
      `${nadaaBrand.name} is operated by ${PENDING}, the data controller for the purposes of Ghana's Data Protection Act, 2012 (Act 843). Our Data Protection Commission registration number is ${PENDING}.`,
      `You can reach our Data Protection Officer at ${PENDING}.`,
    ],
  },
  {
    heading: "What we collect, and why",
    body: [
      "We collect the minimum needed to get help to you. Each item below is collected for the stated purpose and nothing else.",
    ],
    list: [
      {
        term: "Precise location",
        detail:
          "Only while you are checking your area's risk, reporting an incident, or have an active safety session. It is what lets responders find you, and what makes a report actionable rather than approximate.",
      },
      {
        term: "Photos and video you attach",
        detail:
          "Only the media you deliberately attach to a report. We do not scan your photo library.",
      },
      {
        term: "Phone number",
        detail:
          "Used to sign you in by one-time code, and to reach you about an incident you reported.",
      },
      {
        term: "Your name",
        detail:
          "Shown to the responders and coordinators handling your incident so they know who they are helping.",
      },
      {
        term: "Device and session identifiers",
        detail:
          "To keep you signed in, deliver alerts to your device, and detect abuse of the reporting system.",
      },
    ],
  },
  {
    heading: "Who your information reaches",
    body: [
      "An emergency report is only useful if it reaches the people who can act on it. When you report an incident or trigger an alert, the relevant details are shared with the coordinating authority and with the responders assigned to it.",
      "Responders receive only what the active incident requires — not your history, and not your location outside that incident. We do not sell personal data, and we do not share it for advertising.",
    ],
  },
  {
    heading: "How long we keep it",
    body: [
      `Incident records are retained for ${PENDING} to support coordination, review and accountability. Location collected during a safety session is retained only as long as that session's incident record. Account details are kept while your account is open.`,
      "Where a record must be preserved for a legal or public-safety reason, we keep it for that reason alone.",
    ],
  },
  {
    heading: "Your rights under Act 843",
    body: [
      "You can ask us what personal data we hold about you, ask us to correct it, ask us to delete it, object to how we are using it, or withdraw a consent you previously gave. Withdrawing consent does not undo processing that already happened.",
      `Make a request at ${PENDING}. You also have the right to complain to Ghana's Data Protection Commission.`,
    ],
  },
  {
    heading: "Children and dependants",
    body: [
      `Where an account is created or supervised on behalf of a child or dependant, that supervision is granted explicitly and is visible to everyone involved. The arrangement ends when the person reaches the age at which they hold their own rights, or sooner if it is withdrawn. ${PENDING} — the specific ages and consent rules require confirmation from counsel.`,
      "Safety features must not become surveillance. We build them so the person being kept safe can see what is being shared about them.",
    ],
  },
  {
    heading: "Security",
    body: [
      "Data is encrypted in transit. Access to incident data is restricted to the authority and responders handling it, and sensitive actions are logged so they can be reviewed afterwards.",
      "No system is perfect. If a breach affects your data, we will notify you and the Data Protection Commission as the law requires.",
    ],
  },
  {
    heading: "Changes to this notice",
    body: [
      "If we change what we collect or why, we will update this page and tell you in the app before the change takes effect.",
    ],
  },
];

export function PrivacyPage() {
  return (
    <>
      <PageBanner
        eyebrow="Privacy"
        subtitle={`How ${nadaaBrand.name} collects, uses and protects your personal data under Ghana's Data Protection Act, 2012 (Act 843).`}
        title="Your data, and what we do with it."
      />

      <section aria-labelledby="privacy-body" className="content-section">
        <h2 className="sr-only" id="privacy-body">
          Privacy notice
        </h2>

        <div className="legal-doc">
          <p className="legal-lede">
            {nadaaBrand.name} exists to get help moving toward someone who needs
            it. That requires knowing where they are and who to tell — so we
            collect as little as will do the job, keep it only as long as it is
            useful, and are specific about who sees it.
          </p>

          {sections.map((section) => (
            <section key={section.heading}>
              <h3>{section.heading}</h3>
              {section.body.map((paragraph) => (
                <p key={paragraph.slice(0, 48)}>{paragraph}</p>
              ))}
              {section.list ? (
                <dl className="legal-list">
                  {section.list.map((item) => (
                    <div key={item.term}>
                      <dt>{item.term}</dt>
                      <dd>{item.detail}</dd>
                    </div>
                  ))}
                </dl>
              ) : null}
            </section>
          ))}

          <p className="legal-foot">
            In a life-threatening emergency, call {nadaaBrand.supportLine}.
          </p>
        </div>
      </section>
    </>
  );
}
