import { featurePillars, nadaaBrand } from "@nadaa/brand";

export const navItems = [
  { href: "#about", label: "About" },
  { href: "#features", label: "Features" },
  { href: "#platforms", label: "Platforms" },
  { href: "#services", label: "Services" },
  { href: "#benefits", label: "Benefits" },
  { href: "#contact", label: "Contact" },
] as const;

export const heroMetrics = [
  { label: "Platform surfaces", value: "6" },
  { label: "Emergency line", value: nadaaBrand.supportLine },
  { label: "Primary launch hazard", value: "Floods" },
] as const;

export const coreFeatures = [
  ...featurePillars,
  {
    title: "Command Operations",
    description:
      "Dispatchers, agencies, and admins coordinate incidents from role-specific consoles.",
    accent: nadaaBrand.colors.navy,
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
    href: "#platforms",
  },
] as const;
