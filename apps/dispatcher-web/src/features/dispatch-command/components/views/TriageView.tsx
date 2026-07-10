import { MenuItem, Stack } from "@mui/material";
import { Crosshair } from "lucide-react";
import type { DispatchData } from "../../useDispatchData";
import { AITriageSuggestionPanel } from "../AITriageSuggestionPanel";
import { CommandSelect } from "../shared";
import { SectionCard } from "../primitives";
import {
  hazardLabel,
  severityLabel,
  triagePopulationError,
  triageReasonError,
} from "../../utils";

export function TriageView({ data }: { data: DispatchData }) {
  const {
    filteredIncidents,
    selectedIncident,
    selectedIncidentId,
    setSelectedIncidentId,
    triageSuggestion,
    triageLoadState,
    triageMessage,
    triageBusy,
    triageFeedback,
    triageForm,
    updateTriageForm,
    acceptTriageSuggestion,
    overrideTriageSuggestion,
    refreshTriage,
  } = data;

  const canAccept =
    !!triageSuggestion &&
    !!selectedIncident &&
    (selectedIncident.source === "api"
      ? triageLoadState === "ready"
      : triageLoadState === "fallback");
  const canOverride =
    !!triageSuggestion &&
    selectedIncident?.source === "api" &&
    triageLoadState === "ready";

  return (
    <Stack spacing={2.5}>
      <SectionCard
        title="Focused incident"
        eyebrow="AI triage runs on the selected report"
        icon={Crosshair}
        accent="info"
      >
        <CommandSelect
          label="Incident"
          value={selectedIncident ? selectedIncidentId : ""}
          onChange={(event) => setSelectedIncidentId(event.target.value)}
        >
          {filteredIncidents.length === 0 ? (
            <MenuItem value="" disabled>
              No incidents in the current filter
            </MenuItem>
          ) : null}
          {filteredIncidents.map((incident) => (
            <MenuItem value={incident.id} key={incident.id}>
              {incident.reference} · {hazardLabel(incident.type)} ·{" "}
              {severityLabel(incident.severity)} · {incident.district}
            </MenuItem>
          ))}
        </CommandSelect>
      </SectionCard>

      <AITriageSuggestionPanel
        busy={triageBusy}
        canAccept={canAccept}
        canOverride={canOverride}
        feedback={triageFeedback}
        form={triageForm}
        incident={triageSuggestion}
        loadMessage={triageMessage}
        loadState={triageLoadState}
        onAccept={() => void acceptTriageSuggestion()}
        onOverride={() => void overrideTriageSuggestion()}
        onRefresh={() => {
          if (selectedIncident?.source === "api") {
            void refreshTriage(selectedIncident.id);
          }
        }}
        onUpdateForm={updateTriageForm}
        populationError={triagePopulationError(triageForm.affectedPopulation)}
        reasonError={triageReasonError(triageForm.reason)}
        suggestion={triageSuggestion}
      />
    </Stack>
  );
}
