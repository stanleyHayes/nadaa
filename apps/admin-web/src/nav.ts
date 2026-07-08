import { createElement, type ReactElement } from "react";
import {
  BellRing,
  DatabaseZap,
  KeyRound,
  LockKeyhole,
  ShieldCheck,
} from "lucide-react";
import type { AdminView } from "./data/types";

export const viewTabs: Array<{
  id: AdminView;
  label: string;
  icon: ReactElement;
}> = [
  {
    id: "overview",
    label: "Overview",
    icon: createElement(ShieldCheck, { size: 18 }),
  },
  {
    id: "access",
    label: "Access",
    icon: createElement(KeyRound, { size: 18 }),
  },
  {
    id: "audit",
    label: "Audit",
    icon: createElement(LockKeyhole, { size: 18 }),
  },
  {
    id: "integrations",
    label: "Data Sources",
    icon: createElement(DatabaseZap, { size: 18 }),
  },
  {
    id: "alertRules",
    label: "Alert Rules",
    icon: createElement(BellRing, { size: 18 }),
  },
];
