import { featurePillars, nadaaBrand } from "@nadaa/brand";

export const navItems = [
  { href: "#about", label: "About" },
  { href: "#how", label: "How it works" },
  { href: "#platforms", label: "Platforms" },
  { href: "#why", label: "Why NADAA" },
  { href: "#benefits", label: "Benefits" },
  { href: "#contact", label: "Contact" },
] as const;

export const heroMetrics = [
  { label: "Ghanaian languages", value: "6" },
  { label: "Ways we reach you", value: "7" },
  { label: "Coordinated roles", value: "5" },
] as const;

/** Hazards cycled in the home hero rotating-words line. */
export const heroHazards = [
  "floods",
  "fires",
  "road crashes",
  "storms",
  "disease outbreaks",
  "tidal waves",
] as const;

/** Capability keywords for the home marquee strip. */
export const heroMarqueeItems = [
  "Flood & fire risk scoring",
  "SMS + USSD alerts",
  "WhatsApp reporting",
  "Voice call warnings",
  "Cell broadcast",
  "Six Ghanaian languages",
  "112 emergency line",
  "Offline safety guides",
  "Shelter & hospital mapping",
  "Human-approved alerts",
] as const;

export const coreFeatures = [
  ...featurePillars,
  {
    title: "Command Operations",
    description:
      "Dispatchers, agencies, and admins coordinate incidents from role-specific consoles.",
    // CSS token (not a fixed hex) so navy flips to a legible indigo in dark mode.
    accent: "var(--nadaa-navy)",
  },
  {
    title: "Recovery Logistics",
    description:
      "Shelters, hospital capacity, road closures, and relief points help communities recover faster.",
    accent: nadaaBrand.colors.gold,
  },
] as const;

export const serviceLines = [
  {
    title: "Risk Intelligence",
    description:
      "Area risk checks, flood scoring, ML-assisted predictions, and safety guidance for communities and responders.",
    icon: "radar",
  },
  {
    title: "Alerting & Inclusive Access",
    description:
      "Human-approved alerts across web, mobile, SMS/USSD, WhatsApp, and low-literacy voice workflows.",
    icon: "bell",
  },
  {
    title: "Incident Reporting",
    description:
      "Citizen reports with location, media metadata, privacy controls, duplicate detection, and abuse review.",
    icon: "fileWarning",
  },
  {
    title: "Dispatch Command",
    description:
      "Dispatcher triage, verification, assignment, timelines, ML review, road closures, and response status.",
    icon: "route",
  },
  {
    title: "Agency Operations",
    description:
      "Agency-scoped incident work, responder updates, hospital capacity, shelters, and relief distribution.",
    icon: "building",
  },
  {
    title: "Governance & Audit",
    description:
      "Admin control over agencies, users, roles, MFA support, data sources, alert rules, and audit logs.",
    icon: "shield",
  },
] as const;

export const platformLanes = [
  {
    title: "Citizens",
    channels: "Web and mobile",
    summary:
      "Risk checks, incident reports, emergency alerts, offline guides, shelter support, and recovery services.",
    status: "Built across web and mobile foundations",
  },
  {
    title: "Dispatchers",
    channels: "Web and mobile",
    summary:
      "Triage queues, map-based command, incident verification, assignment, responder actions, and shift-friendly mobile workflows.",
    status: "Built for command center and mobile triage",
  },
  {
    title: "Agencies",
    channels: "Web",
    summary:
      "Assigned incident work, responder updates, hospital capacity, road context, shelters, and relief logistics.",
    status: "Built for agency-scoped operations",
  },
  {
    title: "Admins",
    channels: "Web",
    summary:
      "Platform governance, agencies, users, roles, MFA support, audit logs, data sources, and alert rules.",
    status: "Built for controlled administration",
  },
] as const;

export const benefits = [
  {
    audience: "For citizens",
    points: [
      "Know local risk before danger escalates.",
      "Report emergencies with location and media context.",
      "Find shelters, relief points, hospitals, and practical guidance.",
    ],
  },
  {
    audience: "For dispatchers",
    points: [
      "See reports, alerts, closures, and capacity in one operating picture.",
      "Reduce duplicate noise and assign the right responder faster.",
      "Keep every decision traceable through status and timeline events.",
    ],
  },
  {
    audience: "For agencies",
    points: [
      "Focus each team on assigned work without exposing unrelated operations.",
      "Update hospital capacity, shelter occupancy, and relief stock from the field.",
      "Coordinate recovery services after the immediate incident response.",
    ],
  },
  {
    audience: "For national leadership",
    points: [
      "Strengthen people-centered early warning and preparedness.",
      "Connect data from weather, hydrology, emergency services, and districts.",
      "Improve accountability with RBAC, MFA, audit logs, and approval gates.",
    ],
  },
] as const;

export const researchNotes = [
  {
    title: "One emergency access point",
    body: "Ghana's 112 emergency call centre connects citizens to police, fire, ambulance, NADMO, and relief agencies.",
    source: "ITU WSIS Stocktaking: Emergency Call Centre (112)",
    href: "https://www.itu.int/net4/wsis/archive/stocktaking/Project/Details?projectId=1487771718",
  },
  {
    title: "Flood resilience matters",
    body: "World Bank GARID financing highlights flood risk management for the Odaw River Basin and Greater Accra communities.",
    source: "World Bank GARID additional financing release",
    href: "https://www.worldbank.org/en/news/press-release/2023/05/25/world-bank-supports-ghana-to-improve-flood-resilience-for-2-5-million-people",
  },
  {
    title: "Early warning is a national priority",
    body: "GMet's EW4All roadmap emphasizes people-centered, timely, reliable, and actionable weather, water, and climate information.",
    source: "Ghana Meteorological Agency EW4All roadmap article",
    href: "https://www.meteo.gov.gh/news/ghana-validates-comprehensive-roadmap-for-early-warning-systems/",
  },
  {
    title: "Prevention and preparedness are public-facing",
    body: "NADMO public materials emphasize disaster profiles, hazards, risk drivers, and prevention-oriented communication.",
    source: "National Disaster Management Organisation",
    href: "https://www.nadmo.gov.gh/",
  },
] as const;

export const contactCards = [
  {
    title: "Emergency help",
    primary: "Call 112",
    detail:
      "Use Ghana's emergency line for immediate police, fire, ambulance, NADMO, or relief agency support.",
    href: "tel:112",
  },
  {
    title: "Partnerships and demos",
    primary: "Request a platform briefing",
    detail:
      "For agency onboarding, donor briefings, district pilots, or technical partnerships.",
    href: "mailto:partnerships@nadaa.gov.gh?subject=NADAA%20partnership%20request",
  },
  {
    title: "Operational rollout",
    primary: "Plan a deployment lane",
    detail:
      "Use the web, mobile, dispatch, agency, and admin apps as separate rollout tracks.",
    href: "/platforms",
  },
] as const;

export const platformPositioning =
  "One accountable system that carries a flood or fire from the first risk signal all the way to a warned citizen, a dispatched responder, and a recovered household — with every public-safety decision kept in human hands.";

export const responseLoop = [
  {
    step: "01",
    title: "Detect the risk",
    description:
      "Flood and fire risk scoring, ML flood-onset predictions, and governed partner feeds surface where danger is building — by area, saved zone, or live GPS. Predictions are decision-support and never auto-publish.",
  },
  {
    step: "02",
    title: "Warn the public",
    description:
      "An officer drafts an alert from a live risk or incident, targets it by district, region, community, radius, polygon, or nationally, and publishes only after a person approves it — reaching citizens across app, SMS, USSD, WhatsApp, voice, and cell broadcast in six languages.",
  },
  {
    step: "03",
    title: "Report what's happening",
    description:
      "Residents report floods, fires, crashes, collapses, blocked drains, or medical emergencies in seconds — with location, photos, video, urgency, and injuries, anonymously if they choose — feeding real ground truth into the command picture.",
  },
  {
    step: "04",
    title: "Coordinate the response",
    description:
      "Command staff see every incident on one live map, verify reports, merge duplicates of the same event into one traceable incident, clear false-report signals, then assign NADMO, Fire, Ambulance, Police, or the District Assembly with priority and instructions.",
  },
  {
    step: "05",
    title: "Respond with capacity in view",
    description:
      "Assigned agencies work their own incidents through a fixed status lifecycle with a mandatory note trail, routing casualties and evacuees using live shelter capacity, hospital beds and ICU availability, and active road-closure detours.",
  },
  {
    step: "06",
    title: "Recover and reunite",
    description:
      "After the water drops, households file property-damage claims, report or search for missing family, and find where relief is being distributed — while donors pledge aid to verified, anti-fraud-checked needs.",
  },
] as const;

export const roleSurfaces = [
  {
    role: "Citizens",
    icon: "citizen",
    accent: nadaaBrand.colors.green,
    tagline: "Know the risk. Reach safety.",
    audience: "Residents of flood- and disaster-prone communities",
    oneLiner:
      "Check flood and fire risk, get warnings, report incidents, and find shelter, safe routes, and relief — online or offline.",
    channels: ["Web PWA", "Mobile", "SMS", "USSD", "WhatsApp", "Voice 112"],
    status: "MVP",
    capabilities: [
      {
        title: "Check your area's risk",
        description:
          "Look up flood and fire risk by area, saved zone, or live GPS, with a plain-language reason and the nearest response facilities.",
      },
      {
        title: "Live warnings and urgent alerts",
        description:
          "See current warnings with severity, recommended action, and evacuation notices — with push alerts for urgent ones.",
      },
      {
        title: "Report an incident in seconds",
        description:
          "Send a flood, fire, crash, or medical report with location, photos, urgency, and injuries — anonymously if you choose.",
      },
      {
        title: "Find shelters, relief, and safe routes",
        description:
          "See nearby shelters and relief points, and plan a walking route to safety that avoids flooded and closed roads.",
      },
    ],
    benefits: [
      "Decide whether to travel or stay before conditions change.",
      "Reach responders fast with one-tap 112 and located reports.",
      "Stay informed offline — saved warnings and guides keep working.",
    ],
  },
  {
    role: "Command center",
    icon: "authority",
    // CSS token (not a fixed hex) so navy flips to a legible indigo in dark mode.
    accent: "var(--nadaa-navy)",
    tagline: "Command Ghana's disaster response.",
    audience: "NADMO authority and command-center staff",
    oneLiner:
      "Monitor incidents on a live map, verify reports, issue human-approved public alerts, and coordinate shelters, relief, and teams.",
    channels: ["Web"],
    status: "MVP",
    capabilities: [
      {
        title: "Live incident command map",
        description:
          "Every incident on a filterable map and queue — verify a report, move it through its lifecycle, and assign the right agency with priority and instructions.",
      },
      {
        title: "Public alert approval workflow",
        description:
          "Draft an alert, target it by district, region, radius, polygon, or nationally, and publish only after a person approves it.",
      },
      {
        title: "Duplicate merge and safety review",
        description:
          "Collapse many reports of one event into a single traceable incident, and clear abuse and false-report signals.",
      },
      {
        title: "Forecasting and flood simulation",
        description:
          "Per-district demand forecasts, agency staging advice, and a flood scenario runner — all advisory, with humans keeping deployment authority.",
      },
    ],
    benefits: [
      "One live picture of every incident, triaged by severity and district.",
      "Role- and MFA-gated approval reduces premature or unauthorized alerts.",
      "AI forecasts and image labels stay advisory — people decide.",
    ],
  },
  {
    role: "Dispatchers",
    icon: "dispatcher",
    accent: nadaaBrand.colors.red,
    tagline: "One console for disaster response.",
    audience: "Emergency dispatchers and incident-command officers",
    oneLiner:
      "Watch incidents on a live map, triage with AI you can override, dispatch the right agency, and issue approved alerts — at the desk or in the field.",
    channels: ["Web", "Mobile", "Push", "Voice 112"],
    status: "MVP",
    capabilities: [
      {
        title: "Live incident map and queue",
        description:
          "An operations map with road-closure and relief layers, a filterable queue, metric cards, and the full incident table.",
      },
      {
        title: "AI triage you control",
        description:
          "Per-incident severity, agency, and population suggestions with plain-language factors — accept or override, every decision logged.",
      },
      {
        title: "Capacity-aware casualty routing",
        description:
          "See nearby hospitals with live beds, ICU, and ambulance availability so casualties go where there is room.",
      },
      {
        title: "Field-ready, offline-first mobile",
        description:
          "Cached incident queue and capacity, MFA sign-in, status updates, and push escalation when the network drops.",
      },
    ],
    benefits: [
      "Go from citizen report to the right responder faster.",
      "Humans stay accountable for life-and-death calls.",
      "Duplicate-merge cuts through the noise of a flood surge.",
    ],
  },
  {
    role: "Response agencies",
    icon: "agency",
    accent: nadaaBrand.colors.gold,
    tagline: "Coordinate every assigned response.",
    audience: "Fire, ambulance, police, district, and NADMO responders",
    oneLiner:
      "Work the incidents dispatched to your agency, check nearby shelter and hospital capacity, and coordinate relief and donations — securely.",
    channels: ["Web"],
    status: "MVP",
    capabilities: [
      {
        title: "Assigned incident dashboard",
        description:
          "Only your agency's incidents, with live counts for assigned, en route, on scene, and priority, filterable by hazard and status.",
      },
      {
        title: "Controlled status workflow",
        description:
          "Advance each incident through a fixed lifecycle with a note trail and mandatory resolution notes before closing.",
      },
      {
        title: "Nearby capacity context",
        description:
          "Pull nearby shelters, hospital bed and ambulance availability, and road-closure detours for a selected incident.",
      },
      {
        title: "Relief and donation coordination",
        description:
          "Publish relief points, review partner pledges with anti-fraud clearing, and track pledged-versus-needed progress.",
      },
    ],
    benefits: [
      "Crews focus only on the incidents dispatched to them.",
      "A required, timestamped trail keeps every response auditable.",
      "Capacity-aware routing cuts wasted trips during a flood.",
    ],
  },
  {
    role: "Administrators",
    icon: "admin",
    // CSS token (not a fixed hex) so slate lightens for legibility in dark mode.
    accent: "var(--nadaa-slate)",
    tagline: "Accountable control for disaster response.",
    audience: "Platform administrators and governance leads",
    oneLiner:
      "Register agencies, provision MFA-protected users, and govern the audit trail, integration contracts, and alert-approval rules.",
    channels: ["Web"],
    status: "MVP",
    capabilities: [
      {
        title: "Agency and user governance",
        description:
          "Register response agencies and provision authority accounts with role, agency, +233 phone validation, and mandatory MFA setup.",
      },
      {
        title: "Role-based access policy",
        description:
          "A live role matrix maps seven authority roles to console access, alert-approval rights, and dispatcher operations.",
      },
      {
        title: "Secret-redacted audit trail",
        description:
          "A trace of every sensitive action — actor, role, target, request id, time — with tokens and secrets stripped from each snapshot.",
      },
      {
        title: "Integration and alert-rule governance",
        description:
          "Govern partner feeds (GMet, Hydrological Services, NADMO sync) and the approval, override, and targeting rules for mass alerts.",
      },
    ],
    benefits: [
      "Access is denied by default until role and MFA both pass.",
      "Every sensitive action leaves an auditable, redacted trace.",
      "Life-safety alerts stay under clear, governed approval rules.",
    ],
  },
] as const;

export const roleIconKeys = [
  "citizen",
  "authority",
  "dispatcher",
  "agency",
  "admin",
] as const;

export const differentiators = [
  {
    title: "Public-safety decisions stay human",
    description:
      "AI triage, ML flood forecasts, and computer-vision labels are advisory only. Alerts run a submit, approve, reject, and override workflow, sensitive results are held for human review, and every decision is logged.",
  },
  {
    title: "Inclusive access built for Ghana",
    description:
      "NADAA reaches people with no smartphone or data — over SMS, USSD, WhatsApp, voice on 112, and cell broadcast — and the citizen app is offline-first, caching warnings and guides for when the network drops.",
  },
  {
    title: "One flood surge, one clear incident",
    description:
      "When hundreds report the same flood, duplicate detection collapses them into a single incident, so responders act on verified events instead of noise, repeats, or false alarms.",
  },
  {
    title: "Capacity-aware routing, not blind dispatch",
    description:
      "Shelters show open or full status, hospitals show live beds and ambulances, and evacuation routes steer around severe-risk zones and closed roads — so people and casualties go where there is room.",
  },
  {
    title: "Governed, auditable, resilient",
    description:
      "MFA and a seven-role access matrix gate every console, a secret-redacted audit trail records who changed what, and fixture fallbacks keep command views usable when a backend service goes down mid-crisis.",
  },
] as const;

export const impactStats = [
  {
    value: "16",
    icon: "regions",
    label: "Regions across Ghana",
    detail: "National coverage, from Greater Accra to the Upper regions",
  },
  {
    value: "6",
    icon: "languages",
    label: "Ghanaian languages",
    detail: "English, Twi, Ga, Ewe, Dagbani, and Hausa",
  },
  {
    value: "7",
    icon: "channels",
    label: "Ways we reach you",
    detail: "App, SMS, USSD, WhatsApp, voice, cell broadcast, push",
  },
  {
    value: "112",
    icon: "emergency",
    label: "Ghana's emergency line",
    detail: "One number for police, fire, ambulance, and NADMO",
  },
] as const;

export const trustPoints = [
  "No public warning leaves NADAA without a person approving it — AI informs the decision, but never publishes on its own.",
  "Offline-first and low-bandwidth by design: warnings and guides reach citizens over SMS, USSD, WhatsApp, voice, and cell broadcast, and keep working on a cached app when networks fail.",
  "Built for who Ghana actually is — emergency guidance and voice alerts in six Ghanaian languages, with anonymous reporting and accessibility-need capture.",
  "Accountable end to end: MFA-gated roles, a secret-redacted audit trail on every sensitive action, and governed partner feeds with documented manual fallbacks.",
] as const;

export const complianceItems = [
  {
    title: "Data protection",
    description:
      "Personal data is handled under Ghana's Data Protection Act, 2012 (Act 843). NADAA collects the minimum needed, supports anonymous reporting, and never exposes a reporter's identity to the public.",
  },
  {
    title: "Official emergency response",
    description:
      "NADAA supports the National Disaster Management Organisation (NADMO) and Ghana's 112 emergency service. It complements official response and never replaces a call to 112.",
  },
  {
    title: "Human-approved warnings",
    description:
      "Every public warning is approved by an authorized officer before it is sent. Predictions and AI inform the decision but cannot publish on their own.",
  },
  {
    title: "Accessible by design",
    description:
      "Built to WCAG 2.1 AA guidance with scalable text, keyboard support, plain language, and guidance in six Ghanaian languages for a mixed-literacy audience.",
  },
] as const;

export const legalLinks = [
  { label: "Privacy & data protection", href: "#trust" },
  { label: "Accessibility", href: "#trust" },
  { label: "Terms of use", href: "#trust" },
  { label: "Emergency: call 112", href: "tel:112" },
] as const;
