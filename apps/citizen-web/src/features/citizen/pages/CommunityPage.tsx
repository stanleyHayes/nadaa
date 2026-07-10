import { useState } from "react";
import type { ComponentType } from "react";
import { Box, Tab, Tabs } from "@mui/material";
import {
  ClipboardList,
  Database,
  HeartHandshake,
  Megaphone,
  Users,
} from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { PageBanner } from "../components/PageBanner";
import {
  DamageClaim,
  DonorPortal,
  MissingPersonsPanel,
  OpenDataPortal,
  PublicCampaignsPanel,
} from "../components";

type CommunityService = {
  /** Stable key used for tab/panel ids and React reconciliation. */
  key: string;
  label: string;
  icon: LucideIcon;
  Panel: ComponentType;
};

/**
 * The five recovery/community services, in the legacy `#community` order. Each
 * panel is self-contained (owns its own state, effects and data fetching), so
 * this page only has to arrange them.
 */
const COMMUNITY_SERVICES: CommunityService[] = [
  { key: "donate", label: "Donate", icon: HeartHandshake, Panel: DonorPortal },
  {
    key: "claim",
    label: "Damage claim",
    icon: ClipboardList,
    Panel: DamageClaim,
  },
  {
    key: "reunite",
    label: "Missing persons",
    icon: Users,
    Panel: MissingPersonsPanel,
  },
  { key: "data", label: "Open data", icon: Database, Panel: OpenDataPortal },
  {
    key: "campaigns",
    label: "Campaigns",
    icon: Megaphone,
    Panel: PublicCampaignsPanel,
  },
];

/**
 * Recovery & community hub. Migrated from the legacy `#community` section, which
 * rendered the donor portal, damage claim, missing-persons, open-data and public
 * campaigns panels in one long column. Presented here as MUI tabs so the five
 * heavy panels read as a coherent set rather than one long scroll. Every panel
 * stays mounted (inactive ones are visually hidden via the `hidden` attribute,
 * not unmounted) so each keeps its mount-once fetch behavior exactly as before.
 */
function CommunityHub() {
  const [active, setActive] = useState(0);

  return (
    <section aria-label="Community and recovery" className="citizen-section">
      <Box
        sx={{
          borderBottom: 1,
          borderColor: "divider",
          mb: { xs: 2.5, md: 3 },
        }}
      >
        <Tabs
          allowScrollButtonsMobile
          aria-label="Recovery and community services"
          onChange={(_event, value: number) => setActive(value)}
          scrollButtons="auto"
          value={active}
          variant="scrollable"
        >
          {COMMUNITY_SERVICES.map(({ key, label, icon: Icon }, index) => (
            <Tab
              aria-controls={`community-panel-${key}`}
              icon={<Icon aria-hidden="true" size={18} />}
              iconPosition="start"
              id={`community-tab-${key}`}
              key={key}
              label={label}
              sx={{ minHeight: 52, textTransform: "none" }}
              value={index}
            />
          ))}
        </Tabs>
      </Box>

      {COMMUNITY_SERVICES.map(({ key, Panel }, index) => (
        <div
          aria-labelledby={`community-tab-${key}`}
          hidden={active !== index}
          id={`community-panel-${key}`}
          key={key}
          role="tabpanel"
        >
          <Panel />
        </div>
      ))}
    </section>
  );
}

/** Community & recovery (route `/community`). Migrated from the legacy `#community` section. */
export function CommunityPage() {
  return (
    <>
      <PageBanner
        eyebrow="Recovery & community"
        subtitle="Donate, claim damage support, reunite families, and explore open disaster data."
        title="Community & recovery hub"
      />
      <div className="citizen-shell">
        <CommunityHub />
      </div>
    </>
  );
}

export default CommunityPage;
