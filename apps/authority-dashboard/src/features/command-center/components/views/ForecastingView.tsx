import { Stack } from "@mui/material";
import { Cpu } from "lucide-react";
import { ResourcePositioningPanel } from "../ResourcePositioningPanel";
import { FloodSimulationPanel } from "../FloodSimulationPanel";
import { SectionCard } from "../primitives";
import { StatusLine } from "../shared";

export function ForecastingView() {
  return (
    <Stack spacing={2.5}>
      <SectionCard
        title="ML review gate"
        eyebrow="Human-in-the-loop"
        icon={Cpu}
        accent="info"
      >
        <Stack spacing={1.25}>
          <StatusLine label="Model outputs" value="Advisory only" color="warning" />
          <StatusLine
            label="Public alerts"
            value="Require human approval"
            color="warning"
          />
          <StatusLine label="Audit trail" value="Every decision logged" color="success" />
        </Stack>
      </SectionCard>

      <ResourcePositioningPanel />
      <FloodSimulationPanel />
    </Stack>
  );
}
