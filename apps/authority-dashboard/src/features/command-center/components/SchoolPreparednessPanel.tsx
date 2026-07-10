import { type ChangeEvent, useEffect, useMemo, useRef, useState } from "react";
import {
  Alert,
  Box,
  Button,
  Chip,
  Divider,
  FormControlLabel,
  Grid,
  MenuItem,
  Paper,
  Stack,
  Switch,
  Tab,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  Tabs,
  TextField,
  Typography,
} from "@mui/material";
import type { SelectChangeEvent } from "@mui/material/Select";
import L from "leaflet";
import "leaflet/dist/leaflet.css";
import {
  CheckCircle2,
  GraduationCap,
  Loader2,
  MapPin,
  RefreshCw,
  ShieldAlert,
  Users,
} from "lucide-react";
import { nadaaBrand, semantic } from "@nadaa/brand";
import type {
  DrillRecord,
  ReadinessCheck,
  RiskLevel,
  SchoolDetailResponse,
  SchoolListResponse,
  SchoolProfile,
  SchoolReadinessStatus,
  SchoolSummary,
} from "@nadaa/shared-types";
import { SCHOOL_API_BASE } from "@/app/config";
import { authorityHeaders } from "@/app/session";
import { CommandSelect, ScrollableTable, SeverityChip } from "./shared";
import type {
  DrillFormState,
  ReadinessFormState,
  SchoolDetailLoadState,
  SchoolFormState,
  SchoolPanelView,
} from "../types";

const readinessStatusOrder: Record<SchoolReadinessStatus, number> = {
  not_ready: 4,
  needs_improvement: 3,
  not_assessed: 2,
  ready: 1,
};

const drillTypeOptions = [
  "fire",
  "flood",
  "storm",
  "earthquake",
  "lockdown",
  "evacuation",
  "medical",
];

const readinessStatusOptions: SchoolReadinessStatus[] = [
  "ready",
  "needs_improvement",
  "not_ready",
  "not_assessed",
];

const riskLevelOptions: RiskLevel[] = [
  "low",
  "moderate",
  "high",
  "severe",
  "emergency",
];

const defaultChecklist = [
  { label: "Emergency contacts updated", checked: false, category: "admin" },
  { label: "Evacuation routes marked", checked: false, category: "planning" },
  { label: "First aid kits stocked", checked: false, category: "equipment" },
  {
    label: "Teachers trained on alarm protocol",
    checked: false,
    category: "training",
  },
  {
    label: "Students with special needs mapped",
    checked: false,
    category: "planning",
  },
];

function readinessColor(status: SchoolReadinessStatus) {
  switch (status) {
    case "ready":
      return nadaaBrand.colors.green;
    case "needs_improvement":
      return nadaaBrand.colors.gold;
    case "not_ready":
      return nadaaBrand.colors.red;
    default:
      return nadaaBrand.colors.slate;
  }
}

// escapeHtml guards Leaflet popups, which render their string content as HTML,
// against injection from API-provided values.
function escapeHtml(value: string): string {
  return value
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#39;");
}

function readinessLabel(status: SchoolReadinessStatus) {
  switch (status) {
    case "ready":
      return "Ready";
    case "needs_improvement":
      return "Needs improvement";
    case "not_ready":
      return "Not ready";
    case "not_assessed":
      return "Not assessed";
    default:
      return status;
  }
}

function buildDefaultSchoolForm(school?: SchoolProfile): SchoolFormState {
  return {
    name: school?.name ?? "",
    address: school?.address ?? "",
    region: school?.region ?? "Greater Accra",
    district: school?.district ?? "",
    latitude: school ? `${school.location.lat}` : "5.5600",
    longitude: school ? `${school.location.lng}` : "-0.2000",
    studentPopulation: school ? `${school.studentPopulation}` : "",
    emergencyContacts:
      school?.emergencyContacts
        .map((c) => `${c.name}, ${c.role}, ${c.phone ?? ""}`)
        .join("\n") ?? "",
    hazards: school?.hazards.join(", ") ?? "",
    evacuationPoints:
      school?.evacuationPoints
        .map((p) => `${p.label}, ${p.location.lat}, ${p.location.lng}`)
        .join("\n") ?? "",
  };
}

function buildDefaultDrillForm(): DrillFormState {
  return {
    date: new Date().toISOString().slice(0, 16),
    type: "fire",
    participants: "",
    notes: "",
    completed: true,
  };
}

function buildDefaultReadinessForm(): ReadinessFormState {
  return {
    checkDate: new Date().toISOString().slice(0, 10),
    riskLevel: "moderate",
    areaRiskRef: "",
    overallStatus: "needs_improvement",
    notes: "",
    checklistItems: defaultChecklist
      .map((item) => `${item.checked ? "[x]" : "[ ]"} ${item.label}`)
      .join("\n"),
  };
}

function parseContacts(
  value: string,
): { name: string; role: string; phone?: string }[] {
  return value
    .split("\n")
    .map((line) => line.trim())
    .filter(Boolean)
    .map((line) => {
      const parts = line.split(",").map((p) => p.trim());
      return {
        name: parts[0] ?? "",
        role: parts[1] ?? "contact",
        phone: parts[2] || undefined,
      };
    })
    .filter((c) => c.name);
}

function parseEvacuationPoints(
  value: string,
): { label: string; lat: number; lng: number }[] {
  return value
    .split("\n")
    .map((line) => line.trim())
    .filter(Boolean)
    .map((line) => {
      const parts = line.split(",").map((p) => p.trim());
      const lat = Number(parts[1]);
      const lng = Number(parts[2]);
      return {
        label: parts[0] ?? "",
        lat: Number.isFinite(lat) ? lat : 0,
        lng: Number.isFinite(lng) ? lng : 0,
      };
    })
    .filter((p) => p.label);
}

function parseChecklistItems(
  value: string,
): { label: string; checked: boolean; category: string }[] {
  return value
    .split("\n")
    .map((line) => line.trim())
    .filter(Boolean)
    .map((line) => {
      const checked = line.startsWith("[x]") || line.startsWith("[X]");
      const label = line.replace(/^\[[xX ]\]\s*/, "").trim();
      return { label, checked, category: "general" };
    })
    .filter((item) => item.label);
}

export function SchoolPreparednessPanel() {
  const [schools, setSchools] = useState<SchoolSummary[]>([]);
  const [loadState, setLoadState] = useState<SchoolDetailLoadState>("loading");
  const [feedback, setFeedback] = useState("Loading school preparedness data");
  const [districtFilter, setDistrictFilter] = useState("all");
  const [selectedSchoolId, setSelectedSchoolId] = useState("");
  const [view, setView] = useState<SchoolPanelView>("list");
  const [detail, setDetail] = useState<SchoolProfile | undefined>(undefined);
  const [detailLoadState, setDetailLoadState] =
    useState<SchoolDetailLoadState>("loading");
  const [drills, setDrills] = useState<DrillRecord[]>([]);
  const [readiness, setReadiness] = useState<ReadinessCheck | undefined>(
    undefined,
  );
  const [schoolForm, setSchoolForm] = useState<SchoolFormState>(
    buildDefaultSchoolForm(),
  );
  const [drillForm, setDrillForm] = useState<DrillFormState>(
    buildDefaultDrillForm(),
  );
  const [readinessForm, setReadinessForm] = useState<ReadinessFormState>(
    buildDefaultReadinessForm(),
  );
  const [busy, setBusy] = useState(false);
  const mapRef = useRef<HTMLDivElement | null>(null);
  const leafletMapRef = useRef<L.Map | null>(null);
  const layerRef = useRef<L.LayerGroup | null>(null);

  const districtOptions = useMemo(
    () => Array.from(new Set(schools.map((s) => s.district))).sort(),
    [schools],
  );

  const filteredSchools = useMemo(() => {
    if (districtFilter === "all") {
      return schools;
    }
    return schools.filter((s) => s.district === districtFilter);
  }, [districtFilter, schools]);

  const selectedSchool = useMemo(
    () => schools.find((s) => s.id === selectedSchoolId) ?? filteredSchools[0],
    [schools, selectedSchoolId, filteredSchools],
  );

  const refreshSchools = async (signal?: AbortSignal) => {
    setLoadState("loading");
    setFeedback("Loading school preparedness data");
    try {
      const response = await fetch(`${SCHOOL_API_BASE}/schools`, {
        headers: authorityHeaders(),
        signal,
      });
      if (!response.ok) {
        throw new Error(`school API returned ${response.status}`);
      }
      const payload = (await response.json()) as SchoolListResponse;
      setSchools(payload.schools);
      setLoadState("ready");
      setFeedback(
        payload.schools.length
          ? "School preparedness API connected."
          : "No schools are currently registered.",
      );
    } catch (error) {
      if (signal?.aborted) {
        return;
      }
      setSchools([]);
      setLoadState("error");
      setFeedback(
        "School preparedness unavailable. Reconnect the school-service.",
      );
    }
  };

  useEffect(() => {
    const controller = new AbortController();
    void refreshSchools(controller.signal);
    return () => controller.abort();
  }, []);

  const loadSchoolDetail = async (schoolId: string, signal?: AbortSignal) => {
    if (!schoolId) {
      setDetail(undefined);
      return;
    }
    setDetailLoadState("loading");
    try {
      const [profileResponse, drillsResponse, readinessResponse] =
        await Promise.all([
          fetch(`${SCHOOL_API_BASE}/schools/${schoolId}`, {
            headers: authorityHeaders(),
            signal,
          }),
          fetch(`${SCHOOL_API_BASE}/schools/${schoolId}/drills`, {
            headers: authorityHeaders(),
            signal,
          }),
          fetch(`${SCHOOL_API_BASE}/schools/${schoolId}/readiness`, {
            headers: authorityHeaders(),
            signal,
          }),
        ]);
      if (!profileResponse.ok) {
        throw new Error(`school API returned ${profileResponse.status}`);
      }
      const profilePayload =
        (await profileResponse.json()) as SchoolDetailResponse;
      setDetail(profilePayload.school);
      setSchoolForm(buildDefaultSchoolForm(profilePayload.school));

      if (drillsResponse.ok) {
        const drillsPayload = (await drillsResponse.json()) as {
          drills: DrillRecord[];
        };
        setDrills(drillsPayload.drills);
      } else {
        setDrills([]);
      }

      if (readinessResponse.ok) {
        const readinessPayload = (await readinessResponse.json()) as {
          readiness?: ReadinessCheck;
        };
        setReadiness(readinessPayload.readiness);
        if (readinessPayload.readiness) {
          setReadinessForm({
            checkDate: readinessPayload.readiness.checkDate.slice(0, 10),
            riskLevel: readinessPayload.readiness.riskLevel,
            areaRiskRef: readinessPayload.readiness.areaRiskRef ?? "",
            overallStatus: readinessPayload.readiness.overallStatus,
            notes: readinessPayload.readiness.notes ?? "",
            checklistItems: readinessPayload.readiness.checklistItems
              .map((item) => `${item.checked ? "[x]" : "[ ]"} ${item.label}`)
              .join("\n"),
          });
        }
      } else {
        setReadiness(undefined);
      }
      setDetailLoadState("ready");
    } catch (error) {
      if (signal?.aborted) {
        return;
      }
      setDetailLoadState("error");
      setFeedback("School detail unavailable. Reconnect the school-service.");
    }
  };

  useEffect(() => {
    if (!selectedSchool) {
      return;
    }
    const controller = new AbortController();
    void loadSchoolDetail(selectedSchool.id, controller.signal);
    return () => controller.abort();
  }, [selectedSchool?.id]);

  useEffect(() => {
    if (!mapRef.current || leafletMapRef.current) {
      return;
    }
    const map = L.map(mapRef.current, {
      center: [5.586, -0.18],
      zoom: 11,
      zoomControl: true,
      scrollWheelZoom: false,
    });
    L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
      attribution:
        '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>',
      maxZoom: 19,
    }).addTo(map);
    leafletMapRef.current = map;
    layerRef.current = L.layerGroup().addTo(map);
    return () => {
      map.remove();
      leafletMapRef.current = null;
      layerRef.current = null;
    };
  }, []);

  useEffect(() => {
    const layer = layerRef.current;
    const map = leafletMapRef.current;
    if (!layer || !map) {
      return;
    }
    layer.clearLayers();
    if (!filteredSchools.length) {
      return;
    }
    const bounds = L.latLngBounds([]);
    filteredSchools.forEach((school) => {
      const color = readinessColor(school.readinessStatus);
      const marker = L.circleMarker(
        [school.location.lat, school.location.lng],
        {
          radius: 8,
          fillColor: color,
          color: nadaaBrand.colors.navy,
          weight: 1,
          opacity: 1,
          fillOpacity: 0.85,
        },
      );
      marker.bindPopup(
        `<strong>${escapeHtml(school.name)}</strong><br/>${escapeHtml(school.district)}<br/>${escapeHtml(readinessLabel(school.readinessStatus))}`,
      );
      marker.on("click", () => {
        setSelectedSchoolId(school.id);
        setView("detail");
      });
      marker.addTo(layer);
      bounds.extend([school.location.lat, school.location.lng]);
    });
    if (bounds.isValid()) {
      map.fitBounds(bounds, { padding: [20, 20], maxZoom: 14 });
    }
  }, [filteredSchools]);

  const updateSchoolForm =
    (key: keyof SchoolFormState) =>
    (
      event:
        ChangeEvent<HTMLInputElement | HTMLTextAreaElement> | SelectChangeEvent,
    ) => {
      setSchoolForm((current) => ({ ...current, [key]: event.target.value }));
    };

  const updateDrillForm =
    (key: keyof DrillFormState) =>
    (
      event:
        ChangeEvent<HTMLInputElement | HTMLTextAreaElement> | SelectChangeEvent,
    ) => {
      const value =
        "checked" in event.target && typeof event.target.checked === "boolean"
          ? event.target.checked
          : event.target.value;
      setDrillForm((current) => ({ ...current, [key]: value }));
    };

  const updateReadinessForm =
    (key: keyof ReadinessFormState) =>
    (
      event:
        ChangeEvent<HTMLInputElement | HTMLTextAreaElement> | SelectChangeEvent,
    ) => {
      setReadinessForm((current) => ({
        ...current,
        [key]: event.target.value,
      }));
    };

  const saveSchoolProfile = async () => {
    const latitude = Number(schoolForm.latitude);
    const longitude = Number(schoolForm.longitude);
    const studentPopulation = Number(schoolForm.studentPopulation);
    if (
      !schoolForm.name.trim() ||
      !Number.isFinite(latitude) ||
      !Number.isFinite(longitude) ||
      latitude < -90 ||
      latitude > 90 ||
      longitude < -180 ||
      longitude > 180 ||
      !Number.isFinite(studentPopulation) ||
      studentPopulation < 0
    ) {
      setFeedback(
        "School name and valid latitude, longitude, and student population are required.",
      );
      return;
    }
    const payload = {
      name: schoolForm.name.trim(),
      location: { lat: latitude, lng: longitude },
      region: schoolForm.region.trim(),
      district: schoolForm.district.trim(),
      address: schoolForm.address.trim(),
      studentPopulation,
      emergencyContacts: parseContacts(schoolForm.emergencyContacts).map(
        (c) => ({
          ...c,
          isPrimary: false,
          email: undefined,
        }),
      ),
      hazards: schoolForm.hazards
        .split(",")
        .map((h) => h.trim())
        .filter(Boolean),
      evacuationPoints: parseEvacuationPoints(schoolForm.evacuationPoints),
    };

    setBusy(true);
    setFeedback("");
    try {
      const response = await fetch(
        `${SCHOOL_API_BASE}/schools/${selectedSchool?.id ?? ""}`,
        {
          method: "PUT",
          headers: authorityHeaders(),
          body: JSON.stringify(payload),
        },
      );
      if (!response.ok) {
        throw new Error(`school API returned ${response.status}`);
      }
      const updated = (await response.json()) as SchoolProfile;
      setDetail(updated);
      setSchools((current) =>
        current.map((s) =>
          s.id === updated.id
            ? {
                ...s,
                name: updated.name,
                district: updated.district,
                location: updated.location,
                studentPopulation: updated.studentPopulation,
                updatedAt: updated.updatedAt,
              }
            : s,
        ),
      );
      setFeedback(`${updated.name} profile updated.`);
      setView("detail");
    } catch (error) {
      setFeedback(
        "School profile update needs the school-service API and authority session.",
      );
    } finally {
      setBusy(false);
    }
  };

  const addDrill = async () => {
    if (!selectedSchool) {
      return;
    }
    const participants = Number(drillForm.participants);
    if (!Number.isFinite(participants) || participants < 0) {
      setFeedback("Drill participants must be a valid number.");
      return;
    }
    const payload = {
      date: new Date(drillForm.date).toISOString(),
      type: drillForm.type,
      participants,
      notes: drillForm.notes.trim(),
      completed: drillForm.completed,
    };

    setBusy(true);
    setFeedback("");
    try {
      const response = await fetch(
        `${SCHOOL_API_BASE}/schools/${selectedSchool.id}/drills`,
        {
          method: "POST",
          headers: authorityHeaders(),
          body: JSON.stringify(payload),
        },
      );
      if (!response.ok) {
        throw new Error(`school API returned ${response.status}`);
      }
      const drill = (await response.json()) as DrillRecord;
      setDrills((current) => [drill, ...current]);
      setDrillForm(buildDefaultDrillForm());
      setFeedback(`Drill recorded for ${selectedSchool.name}.`);
      setView("detail");
    } catch (error) {
      setFeedback(
        "Drill save needs the school-service API and authority session.",
      );
    } finally {
      setBusy(false);
    }
  };

  const submitReadiness = async () => {
    if (!selectedSchool) {
      return;
    }
    const payload = {
      checkDate: new Date(readinessForm.checkDate).toISOString(),
      riskLevel: readinessForm.riskLevel,
      areaRiskRef: readinessForm.areaRiskRef.trim(),
      overallStatus: readinessForm.overallStatus,
      notes: readinessForm.notes.trim(),
      checklistItems: parseChecklistItems(readinessForm.checklistItems),
    };

    setBusy(true);
    setFeedback("");
    try {
      const response = await fetch(
        `${SCHOOL_API_BASE}/schools/${selectedSchool.id}/readiness`,
        {
          method: "POST",
          headers: authorityHeaders(),
          body: JSON.stringify(payload),
        },
      );
      if (!response.ok) {
        throw new Error(`school API returned ${response.status}`);
      }
      const check = (await response.json()) as ReadinessCheck;
      setReadiness(check);
      setSchools((current) =>
        current.map((s) =>
          s.id === selectedSchool.id
            ? { ...s, readinessStatus: check.overallStatus }
            : s,
        ),
      );
      setFeedback(`Readiness check submitted for ${selectedSchool.name}.`);
      setView("detail");
    } catch (error) {
      setFeedback(
        "Readiness check needs the school-service API and authority session.",
      );
    } finally {
      setBusy(false);
    }
  };

  return (
    <Paper className="surface school-panel">
      <Stack
        direction={{ xs: "column", sm: "row" }}
        spacing={1}
        className="section-heading"
        sx={{
          justifyContent: "space-between",
          alignItems: { xs: "stretch", sm: "center" }
        }}>
        <Stack direction="row" spacing={1} sx={{
          alignItems: "center"
        }}>
          <GraduationCap size={21} color="var(--nadaa-navy)" />
          <Box>
            <Typography variant="h6">School preparedness</Typography>
            <Typography variant="caption" sx={{
              color: "text.secondary"
            }}>
              Manage school plans, drills, and readiness checks
            </Typography>
          </Box>
        </Stack>
        <Button
          type="button"
          variant="outlined"
          size="small"
          startIcon={
            loadState === "loading" ? (
              <Loader2 size={16} className="spin-icon" />
            ) : (
              <RefreshCw size={16} />
            )
          }
          onClick={() => void refreshSchools()}
          disabled={loadState === "loading"}
        >
          Refresh
        </Button>
      </Stack>
      {feedback ? (
        <Alert
          severity={loadState === "error" ? "error" : "info"}
          className="feed-alert"
        >
          {feedback}
        </Alert>
      ) : null}
      <Tabs
        value={view}
        onChange={(_, next) => setView(next as SchoolPanelView)}
        textColor="secondary"
        indicatorColor="secondary"
      >
        <Tab value="list" label="School list" />
        <Tab value="detail" label="Profile" />
        <Tab value="drill" label="Drill" />
        <Tab value="readiness" label="Readiness" />
      </Tabs>
      {view === "list" && (
        <Stack spacing={2} sx={{ mt: 2 }}>
          <Grid container spacing={1.5}>
            <Grid size={{ xs: 12, md: 4 }}>
              <CommandSelect
                label="District"
                value={districtFilter}
                onChange={(event) => setDistrictFilter(event.target.value)}
              >
                <MenuItem value="all">All districts</MenuItem>
                {districtOptions.map((district) => (
                  <MenuItem value={district} key={district}>
                    {district}
                  </MenuItem>
                ))}
              </CommandSelect>
            </Grid>
            <Grid size={{ xs: 12, md: 8 }}>
              <Box
                ref={mapRef}
                sx={{
                  height: 280,
                  width: "100%",
                  borderRadius: 1,
                  border: `1px solid ${semantic.border}`,
                  backgroundColor: semantic.surface,
                }}
                aria-label="School preparedness map"
              />
            </Grid>
          </Grid>

          {filteredSchools.length ? (
            <ScrollableTable label="School list table">
              <Table size="small">
                <TableHead>
                  <TableRow>
                    <TableCell>School</TableCell>
                    <TableCell>District</TableCell>
                    <TableCell>Students</TableCell>
                    <TableCell>Readiness</TableCell>
                    <TableCell>Last drill</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {filteredSchools
                    .slice()
                    .sort(
                      (a, b) =>
                        readinessStatusOrder[a.readinessStatus] -
                        readinessStatusOrder[b.readinessStatus],
                    )
                    .map((school) => (
                      <TableRow
                        key={school.id}
                        hover
                        selected={school.id === selectedSchool?.id}
                        onClick={() => {
                          setSelectedSchoolId(school.id);
                          setView("detail");
                        }}
                        className="incident-row"
                      >
                        <TableCell>
                          <Typography variant="subtitle2">
                            {school.name}
                          </Typography>
                        </TableCell>
                        <TableCell>{school.district}</TableCell>
                        <TableCell>
                          {school.studentPopulation.toLocaleString()}
                        </TableCell>
                        <TableCell>
                          <Chip
                            size="small"
                            label={readinessLabel(school.readinessStatus)}
                            style={{
                              backgroundColor: readinessColor(
                                school.readinessStatus,
                              ),
                              color: "#fff",
                            }}
                          />
                        </TableCell>
                        <TableCell>
                          {school.lastDrillDate
                            ? new Date(
                                school.lastDrillDate,
                              ).toLocaleDateString()
                            : "—"}
                        </TableCell>
                      </TableRow>
                    ))}
                </TableBody>
              </Table>
            </ScrollableTable>
          ) : (
            <Alert severity="info">
              No schools match the selected district.
            </Alert>
          )}
        </Stack>
      )}
      {view === "detail" && (
        <Stack spacing={2} sx={{ mt: 2 }}>
          {detailLoadState === "loading" && (
            <Loader2 size={20} className="spin-icon" />
          )}
          {!detail && detailLoadState !== "loading" && (
            <Alert severity="info">
              Select a school from the list to view its profile.
            </Alert>
          )}
          {detail && (
            <>
              <Stack direction="row" spacing={1} sx={{
                alignItems: "center"
              }}>
                <MapPin size={18} color="var(--nadaa-navy)" />
                <Typography variant="h6">{detail.name}</Typography>
              </Stack>
              <Grid container spacing={1}>
                <Grid size={{ xs: 6 }}>
                  <Stack direction="row" spacing={0.5} sx={{
                    alignItems: "center"
                  }}>
                    <Users size={16} />
                    <Typography variant="body2">
                      {detail.studentPopulation.toLocaleString()} students
                    </Typography>
                  </Stack>
                </Grid>
                <Grid size={{ xs: 6 }}>
                  <Typography variant="body2" sx={{
                    color: "text.secondary"
                  }}>
                    {detail.district}
                  </Typography>
                </Grid>
              </Grid>

              <Divider />

              <Typography variant="subtitle2">Emergency contacts</Typography>
              {detail.emergencyContacts.length ? (
                <Stack spacing={0.5}>
                  {detail.emergencyContacts.map((contact, index) => (
                    <Typography variant="body2" key={index}>
                      {contact.name} — {contact.role}
                      {contact.phone ? ` · ${contact.phone}` : ""}
                    </Typography>
                  ))}
                </Stack>
              ) : (
                <Typography variant="body2" sx={{
                  color: "text.secondary"
                }}>
                  No contacts recorded.
                </Typography>
              )}

              <Typography variant="subtitle2">Hazards</Typography>
              <Stack direction="row" spacing={0.5} sx={{
                flexWrap: "wrap"
              }}>
                {detail.hazards.length ? (
                  detail.hazards.map((hazard) => (
                    <Chip key={hazard} size="small" label={hazard} />
                  ))
                ) : (
                  <Typography variant="body2" sx={{
                    color: "text.secondary"
                  }}>
                    No hazards recorded.
                  </Typography>
                )}
              </Stack>

              <Typography variant="subtitle2">Evacuation points</Typography>
              {detail.evacuationPoints.length ? (
                <Stack spacing={0.5}>
                  {detail.evacuationPoints.map((point, index) => (
                    <Typography variant="body2" key={index}>
                      {point.label}
                      {point.capacity ? ` (capacity ${point.capacity})` : ""}
                    </Typography>
                  ))}
                </Stack>
              ) : (
                <Typography variant="body2" sx={{
                  color: "text.secondary"
                }}>
                  No evacuation points recorded.
                </Typography>
              )}

              <Divider />

              <Typography variant="subtitle2">Drill history</Typography>
              {drills.length ? (
                <Stack spacing={0.5}>
                  {drills.slice(0, 3).map((drill) => (
                    <Typography variant="body2" key={drill.id}>
                      {new Date(drill.date).toLocaleDateString()} · {drill.type}{" "}
                      · {drill.participants} participants
                      {drill.completed ? " · completed" : ""}
                    </Typography>
                  ))}
                </Stack>
              ) : (
                <Typography variant="body2" sx={{
                  color: "text.secondary"
                }}>
                  No drills recorded.
                </Typography>
              )}

              <Typography variant="subtitle2">Latest readiness</Typography>
              {readiness ? (
                <Stack spacing={0.5}>
                  <Stack direction="row" spacing={1} sx={{
                    alignItems: "center"
                  }}>
                    <SeverityChip severity={readiness.riskLevel} />
                    <Chip
                      size="small"
                      label={readinessLabel(readiness.overallStatus)}
                      style={{
                        backgroundColor: readinessColor(
                          readiness.overallStatus,
                        ),
                        color: "#fff",
                      }}
                    />
                  </Stack>
                  <Typography variant="body2" sx={{
                    color: "text.secondary"
                  }}>
                    Checked {new Date(readiness.checkDate).toLocaleDateString()}
                    {readiness.areaRiskRef
                      ? ` · Risk ref ${readiness.areaRiskRef}`
                      : ""}
                  </Typography>
                  {readiness.notes ? (
                    <Typography variant="body2">{readiness.notes}</Typography>
                  ) : null}
                </Stack>
              ) : (
                <Typography variant="body2" sx={{
                  color: "text.secondary"
                }}>
                  No readiness check recorded.
                </Typography>
              )}

              <Stack direction="row" spacing={1}>
                <Button
                  variant="outlined"
                  size="small"
                  onClick={() => setView("drill")}
                >
                  Add drill
                </Button>
                <Button
                  variant="outlined"
                  size="small"
                  onClick={() => setView("readiness")}
                >
                  Submit readiness
                </Button>
              </Stack>
            </>
          )}
        </Stack>
      )}
      {view === "drill" && (
        <Stack spacing={2} sx={{ mt: 2 }}>
          <Typography variant="h6">
            Record drill {selectedSchool ? `— ${selectedSchool.name}` : ""}
          </Typography>
          <Grid container spacing={1.5}>
            <Grid size={{ xs: 6 }}>
              <TextField
                label="Date"
                type="datetime-local"
                size="small"
                fullWidth
                value={drillForm.date}
                onChange={updateDrillForm("date")}
              />
            </Grid>
            <Grid size={{ xs: 6 }}>
              <CommandSelect
                label="Drill type"
                value={drillForm.type}
                onChange={updateDrillForm("type")}
              >
                {drillTypeOptions.map((type) => (
                  <MenuItem value={type} key={type}>
                    {type}
                  </MenuItem>
                ))}
              </CommandSelect>
            </Grid>
            <Grid size={{ xs: 6 }}>
              <TextField
                label="Participants"
                size="small"
                fullWidth
                value={drillForm.participants}
                onChange={updateDrillForm("participants")}
                slotProps={{
                  htmlInput: { inputMode: "numeric" }
                }}
              />
            </Grid>
            <Grid size={{ xs: 6 }}>
              <FormControlLabel
                control={
                  <Switch
                    checked={drillForm.completed}
                    onChange={updateDrillForm("completed")}
                  />
                }
                label="Completed"
              />
            </Grid>
            <Grid size={{ xs: 12 }}>
              <TextField
                label="Notes"
                size="small"
                fullWidth
                multiline
                minRows={2}
                value={drillForm.notes}
                onChange={updateDrillForm("notes")}
              />
            </Grid>
          </Grid>
          <Button
            variant="contained"
            startIcon={<CheckCircle2 size={17} />}
            onClick={() => void addDrill()}
            disabled={busy || !selectedSchool}
          >
            {busy ? "Saving" : "Save drill"}
          </Button>
        </Stack>
      )}
      {view === "readiness" && (
        <Stack spacing={2} sx={{ mt: 2 }}>
          <Typography variant="h6">
            Readiness check {selectedSchool ? `— ${selectedSchool.name}` : ""}
          </Typography>
          <Grid container spacing={1.5}>
            <Grid size={{ xs: 6 }}>
              <TextField
                label="Check date"
                type="date"
                size="small"
                fullWidth
                value={readinessForm.checkDate}
                onChange={updateReadinessForm("checkDate")}
              />
            </Grid>
            <Grid size={{ xs: 6 }}>
              <CommandSelect
                label="Overall status"
                value={readinessForm.overallStatus}
                onChange={updateReadinessForm("overallStatus")}
              >
                {readinessStatusOptions.map((status) => (
                  <MenuItem value={status} key={status}>
                    {readinessLabel(status)}
                  </MenuItem>
                ))}
              </CommandSelect>
            </Grid>
            <Grid size={{ xs: 6 }}>
              <CommandSelect
                label="Area risk level"
                value={readinessForm.riskLevel}
                onChange={updateReadinessForm("riskLevel")}
              >
                {riskLevelOptions.map((level) => (
                  <MenuItem value={level} key={level}>
                    {level}
                  </MenuItem>
                ))}
              </CommandSelect>
            </Grid>
            <Grid size={{ xs: 6 }}>
              <TextField
                label="Area risk reference"
                size="small"
                fullWidth
                value={readinessForm.areaRiskRef}
                onChange={updateReadinessForm("areaRiskRef")}
                helperText="Optional risk API reference"
              />
            </Grid>
            <Grid size={{ xs: 12 }}>
              <TextField
                label="Checklist (one per line, [x] to mark checked)"
                size="small"
                fullWidth
                multiline
                minRows={4}
                value={readinessForm.checklistItems}
                onChange={updateReadinessForm("checklistItems")}
              />
            </Grid>
            <Grid size={{ xs: 12 }}>
              <TextField
                label="Notes"
                size="small"
                fullWidth
                multiline
                minRows={2}
                value={readinessForm.notes}
                onChange={updateReadinessForm("notes")}
              />
            </Grid>
          </Grid>
          <Button
            variant="contained"
            startIcon={<ShieldAlert size={17} />}
            onClick={() => void submitReadiness()}
            disabled={busy || !selectedSchool}
          >
            {busy ? "Saving" : "Submit readiness check"}
          </Button>
        </Stack>
      )}
    </Paper>
  );
}
