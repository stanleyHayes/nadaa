export const nadaaBrand = {
  name: "NADAA",
  fullName: "National Disaster Alert & Response Platform",
  country: "Ghana",
  slogan: "Be Aware. Be Prepared. Be Safe.",
  supportLine: "112",
  colors: {
    navy: "#1A1A1A",
    charcoal: "#2B2B2B",
    silver: "#8E8E8E",
    lightSilver: "#E6E6E6",
    green: "#118D4E",
    red: "#E53935",
    gold: "#F4C20D",
    slate: "#5A5A5A",
    white: "#FFFFFF",
    mist: "#F4F4F5",
    ink: "#0D0D0D",
  },
  meanings: {
    navy: "Authority & Focus",
    charcoal: "Depth & Structure",
    silver: "Clarity & Precision",
    green: "Safety & Growth",
    red: "Alert & Urgency",
    gold: "Hope & Optimism",
    slate: "Stability & Strength",
  },
} as const;

export const featurePillars = [
  {
    title: "Know Your Risk",
    description: "Check risks in any area anytime",
    accent: nadaaBrand.colors.green,
  },
  {
    title: "Get Alerts",
    description: "Receive timely warnings that save lives",
    accent: nadaaBrand.colors.red,
  },
  {
    title: "Report Incidents",
    description: "Report disasters and accidents",
    accent: nadaaBrand.colors.gold,
  },
  {
    title: "Get Help",
    description: "Connect to emergency services fast",
    accent: "#42A5F5",
  },
  {
    title: "Stay Informed",
    description: "Learn how to prepare, respond, and recover",
    accent: nadaaBrand.colors.green,
  },
  {
    title: "Stronger Together",
    description: "Building resilient communities",
    accent: nadaaBrand.colors.red,
  },
] as const;

export const hazardPalette = {
  flood: "#0B6FB8",
  fire: nadaaBrand.colors.red,
  medical: nadaaBrand.colors.green,
  geological: "#9A5A23",
  road: "#4C5563",
  storm: "#3E8ED0",
  disease: "#7C3AED",
} as const;
