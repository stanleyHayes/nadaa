import { useEffect, useMemo, useState } from "react";
import type {
  AidRequestListResponse,
  AidRequestRecord,
  CreateAidRequestRequest,
  CreateReliefPointRequest,
  HospitalCapacityRecord,
  HospitalCapacityResponse,
  HospitalCapacityUpdateRequest,
  HospitalCapacityUpdateResponse,
  IncidentListResponse,
  IncidentRecord,
  IncidentStatusUpdateRequest,
  ReliefPointListResponse,
  ReliefPointNearbyResponse,
  ReliefPointRecord,
  ReliefPointStockHistoryResponse,
  ReliefStockHistoryEntry,
  ReviewAidRequestRequest,
  RoadClosureListResponse,
  RoadClosureRecord,
  ShelterOccupancyUpdateRequest,
  ShelterRecord,
  ShelterUpdateResponse,
  UpdateReliefPointRequest,
} from "@nadaa/shared-types";
import {
  agencyHeaders,
  canManageShelterResources,
  type AgencySession,
} from "@/app/session";
import {
  INCIDENT_API_BASE,
  ROAD_CLOSURE_API_BASE,
  SHELTER_API_BASE,
} from "@/app/config";
import {
  initialAidRequestForm,
  initialHospitalCapacityForm,
  initialReliefPointForm,
  initialShelterOccupancyForm,
  initialStatusForm,
} from "./data";
import type {
  AidRequestFormState,
  HospitalCapacityFormState,
  IncidentFilterState,
  IncidentLoadState,
  ReliefPointFormState,
  ShelterOccupancyFormState,
  StatusFormState,
  UpdateLoadState,
} from "./types";
import {
  aidRequestToForm,
  hospitalToCapacityForm,
  matchesFilters,
  parseStockCategories,
  reliefPointToForm,
  shelterToOccupancyForm,
} from "./utils";

/** Default scene used to seed capacity context before an incident is picked. */
const DEFAULT_CAPACITY_LOCATION = { lat: 5.586, lng: -0.18 };

async function extractError(response: Response): Promise<string> {
  try {
    const body = (await response.json()) as {
      message?: string;
      error?: string | { message?: string };
    };
    if (body.message) return body.message;
    if (typeof body.error === "string") return body.error;
    if (body.error?.message) return body.error.message;
    return `Request failed (${response.status})`;
  } catch {
    return `Request failed (${response.status})`;
  }
}

async function fetchAssignedIncidents(): Promise<IncidentRecord[]> {
  const response = await fetch(
    `${INCIDENT_API_BASE}/incidents?assignedToMe=true`,
    { headers: agencyHeaders() },
  );
  if (!response.ok) {
    throw new Error(await extractError(response));
  }
  const payload = (await response.json()) as IncidentListResponse;
  return payload.incidents;
}

async function patchIncidentStatus(
  incidentId: string,
  request: IncidentStatusUpdateRequest,
): Promise<IncidentRecord> {
  const response = await fetch(
    `${INCIDENT_API_BASE}/incidents/${encodeURIComponent(incidentId)}/status`,
    {
      body: JSON.stringify(request),
      headers: agencyHeaders(),
      method: "PATCH",
    },
  );
  if (!response.ok) {
    throw new Error(await extractError(response));
  }
  return (await response.json()) as IncidentRecord;
}

async function fetchShelters(
  lat: number,
  lng: number,
): Promise<ShelterRecord[]> {
  try {
    const params = new URLSearchParams({
      lat: lat.toString(),
      lng: lng.toString(),
    });
    const response = await fetch(
      `${SHELTER_API_BASE}/shelters/nearby?${params}`,
    );
    if (!response.ok) {
      throw new Error(await extractError(response));
    }
    const payload = (await response.json()) as { shelters: ShelterRecord[] };
    return payload.shelters;
  } catch {
    return [];
  }
}

async function fetchHospitalCapacity(
  lat: number,
  lng: number,
): Promise<HospitalCapacityRecord[]> {
  try {
    const params = new URLSearchParams({
      includeStale: "true",
      lat: lat.toString(),
      limit: "6",
      lng: lng.toString(),
    });
    const response = await fetch(
      `${SHELTER_API_BASE}/hospitals/capacity?${params}`,
    );
    if (!response.ok) {
      throw new Error(await extractError(response));
    }
    const payload = (await response.json()) as HospitalCapacityResponse;
    return payload.facilities;
  } catch {
    return [];
  }
}

async function fetchRoadClosures(
  lat: number,
  lng: number,
): Promise<RoadClosureRecord[]> {
  try {
    const params = new URLSearchParams({
      lat: lat.toString(),
      limit: "6",
      lng: lng.toString(),
    });
    const response = await fetch(
      `${ROAD_CLOSURE_API_BASE}/road-closures?${params}`,
    );
    if (!response.ok) {
      throw new Error(await extractError(response));
    }
    const payload = (await response.json()) as RoadClosureListResponse;
    return payload.closures;
  } catch {
    return [];
  }
}

async function fetchNearbyReliefPoints(
  lat: number,
  lng: number,
): Promise<ReliefPointRecord[]> {
  try {
    const params = new URLSearchParams({
      lat: lat.toString(),
      limit: "6",
      lng: lng.toString(),
    });
    const response = await fetch(
      `${SHELTER_API_BASE}/relief-points/nearby?${params}`,
    );
    if (!response.ok) {
      throw new Error(await extractError(response));
    }
    const payload = (await response.json()) as ReliefPointNearbyResponse;
    return payload.reliefPoints;
  } catch {
    return [];
  }
}

async function fetchReliefPoints(): Promise<ReliefPointRecord[]> {
  const response = await fetch(`${SHELTER_API_BASE}/relief-points?limit=20`);
  if (!response.ok) {
    throw new Error(await extractError(response));
  }
  const payload = (await response.json()) as ReliefPointListResponse;
  return payload.reliefPoints;
}

async function fetchReliefPointHistory(
  reliefPointId: string,
): Promise<ReliefStockHistoryEntry[]> {
  const response = await fetch(
    `${SHELTER_API_BASE}/relief-points/${encodeURIComponent(reliefPointId)}/stock-history`,
  );
  if (!response.ok) {
    throw new Error(await extractError(response));
  }
  const payload = (await response.json()) as ReliefPointStockHistoryResponse;
  return payload.history;
}

async function createReliefPoint(
  request: CreateReliefPointRequest,
): Promise<ReliefPointRecord> {
  const response = await fetch(`${SHELTER_API_BASE}/relief-points`, {
    body: JSON.stringify(request),
    headers: agencyHeaders(),
    method: "POST",
  });
  if (!response.ok) {
    throw new Error(await extractError(response));
  }
  return (await response.json()) as ReliefPointRecord;
}

async function updateReliefPoint(
  reliefPointId: string,
  request: UpdateReliefPointRequest,
): Promise<ReliefPointRecord> {
  const response = await fetch(
    `${SHELTER_API_BASE}/relief-points/${encodeURIComponent(reliefPointId)}`,
    {
      body: JSON.stringify(request),
      headers: agencyHeaders(),
      method: "PATCH",
    },
  );
  if (!response.ok) {
    throw new Error(await extractError(response));
  }
  return (await response.json()) as ReliefPointRecord;
}

async function updateShelterOccupancy(
  shelterId: string,
  request: ShelterOccupancyUpdateRequest,
): Promise<ShelterRecord> {
  const response = await fetch(
    `${SHELTER_API_BASE}/shelters/${encodeURIComponent(shelterId)}/occupancy`,
    {
      body: JSON.stringify(request),
      headers: agencyHeaders(),
      method: "PATCH",
    },
  );
  if (!response.ok) {
    throw new Error(await extractError(response));
  }
  const payload = (await response.json()) as ShelterUpdateResponse;
  return payload.shelter;
}

async function updateHospitalCapacity(
  facilityId: string,
  request: HospitalCapacityUpdateRequest,
): Promise<HospitalCapacityRecord> {
  const response = await fetch(
    `${SHELTER_API_BASE}/hospitals/${encodeURIComponent(facilityId)}/capacity`,
    {
      body: JSON.stringify(request),
      headers: agencyHeaders(),
      method: "PATCH",
    },
  );
  if (!response.ok) {
    throw new Error(await extractError(response));
  }
  const payload = (await response.json()) as HospitalCapacityUpdateResponse;
  return payload.facility;
}

async function fetchAidRequests(): Promise<AidRequestRecord[]> {
  const response = await fetch(
    `${SHELTER_API_BASE}/aid-requests?includePrivate=true&limit=30`,
    { headers: agencyHeaders() },
  );
  if (response.status === 403) {
    // Private aid listing is limited to AidRequestViewRoles; responders and
    // viewers get a clear reason instead of a raw API error.
    throw new Error(
      "Your role does not have access to private aid requests.",
    );
  }
  if (!response.ok) {
    throw new Error(await extractError(response));
  }
  const payload = (await response.json()) as AidRequestListResponse;
  return payload.aidRequests;
}

async function createAidRequest(
  request: CreateAidRequestRequest,
): Promise<AidRequestRecord> {
  const response = await fetch(`${SHELTER_API_BASE}/aid-requests`, {
    body: JSON.stringify(request),
    headers: agencyHeaders(),
    method: "POST",
  });
  if (!response.ok) {
    throw new Error(await extractError(response));
  }
  return (await response.json()) as AidRequestRecord;
}

async function reviewAidRequest(
  aidRequestId: string,
  request: ReviewAidRequestRequest,
): Promise<AidRequestRecord> {
  const response = await fetch(
    `${SHELTER_API_BASE}/aid-requests/${encodeURIComponent(aidRequestId)}/review`,
    {
      body: JSON.stringify(request),
      headers: agencyHeaders(),
      method: "PATCH",
    },
  );
  if (!response.ok) {
    throw new Error(await extractError(response));
  }
  return (await response.json()) as AidRequestRecord;
}

/**
 * Central agency-operations state container. Holds every incident, capacity,
 * relief, and aid workflow so the shell can mount data once and route between
 * views without losing state. Collections start empty and are populated from
 * the live services; failures surface a loading/error state per view.
 */
export function useAgencyData(session: AgencySession) {
  const [incidents, setIncidents] = useState<IncidentRecord[]>([]);
  const [incidentLoadState, setIncidentLoadState] =
    useState<IncidentLoadState>("loading");
  const [incidentError, setIncidentError] = useState<string | null>(null);
  const [selectedIncidentId, setSelectedIncidentId] = useState<string | null>(
    null,
  );
  const [filters, setFilters] = useState<IncidentFilterState>({
    hazard: "all",
    severity: "all",
    status: "all",
  });
  const [statusForm, setStatusForm] =
    useState<StatusFormState>(initialStatusForm);
  const [statusUpdateState, setStatusUpdateState] =
    useState<UpdateLoadState>("idle");
  const [statusUpdateError, setStatusUpdateError] = useState<string | null>(
    null,
  );
  const [shelters, setShelters] = useState<ShelterRecord[]>([]);
  const [shelterForm, setShelterForm] = useState<ShelterOccupancyFormState>(
    initialShelterOccupancyForm,
  );
  const [selectedShelterId, setSelectedShelterId] = useState<string | null>(
    null,
  );
  const [hospitals, setHospitals] = useState<HospitalCapacityRecord[]>([]);
  const [hospitalForm, setHospitalForm] = useState<HospitalCapacityFormState>(
    initialHospitalCapacityForm,
  );
  const [selectedHospitalId, setSelectedHospitalId] = useState<string | null>(
    null,
  );
  const [capacityUpdateState, setCapacityUpdateState] =
    useState<UpdateLoadState>("idle");
  const [capacityUpdateError, setCapacityUpdateError] = useState<
    string | null
  >(null);
  const [capacityLoadState, setCapacityLoadState] =
    useState<IncidentLoadState>("loading");
  const [roadClosures, setRoadClosures] = useState<RoadClosureRecord[]>([]);
  const [nearbyReliefPoints, setNearbyReliefPoints] = useState<
    ReliefPointRecord[]
  >([]);
  const [reliefPoints, setReliefPoints] = useState<ReliefPointRecord[]>([]);
  const [selectedReliefPointId, setSelectedReliefPointId] = useState<
    string | null
  >(null);
  const [reliefForm, setReliefForm] = useState<ReliefPointFormState>(
    initialReliefPointForm,
  );
  const [reliefHistory, setReliefHistory] = useState<ReliefStockHistoryEntry[]>(
    [],
  );
  const [reliefLoadState, setReliefLoadState] =
    useState<IncidentLoadState>("loading");
  const [reliefUpdateState, setReliefUpdateState] =
    useState<UpdateLoadState>("idle");
  const [reliefError, setReliefError] = useState<string | null>(null);
  const [aidRequests, setAidRequests] = useState<AidRequestRecord[]>([]);
  const [selectedAidRequestId, setSelectedAidRequestId] = useState<
    string | null
  >(null);
  const [aidForm, setAidForm] = useState<AidRequestFormState>(
    initialAidRequestForm,
  );
  const [aidLoadState, setAidLoadState] =
    useState<IncidentLoadState>("loading");
  const [aidUpdateState, setAidUpdateState] = useState<UpdateLoadState>("idle");
  const [aidError, setAidError] = useState<string | null>(null);

  async function loadIncidents() {
    setIncidentLoadState("loading");
    setIncidentError(null);
    try {
      const data = await fetchAssignedIncidents();
      setIncidents(data);
      setIncidentLoadState(data.length > 0 ? "ready" : "empty");
    } catch (error) {
      setIncidents([]);
      setIncidentError(
        error instanceof Error
          ? error.message
          : "Could not load assigned incidents.",
      );
      setIncidentLoadState("error");
    }
  }

  async function loadReliefPoints() {
    setReliefLoadState("loading");
    setReliefError(null);
    try {
      const data = await fetchReliefPoints();
      setReliefPoints(data);
      setReliefLoadState(data.length > 0 ? "ready" : "empty");
      setSelectedReliefPointId((current) => current ?? data[0]?.id ?? null);
    } catch (error) {
      setReliefPoints([]);
      setReliefError(
        error instanceof Error
          ? error.message
          : "Could not load relief distribution points.",
      );
      setReliefLoadState("error");
    }
  }

  async function loadReliefHistory(reliefPointId: string) {
    try {
      const history = await fetchReliefPointHistory(reliefPointId);
      setReliefHistory(history);
    } catch (error) {
      setReliefHistory([]);
      setReliefError(
        error instanceof Error
          ? error.message
          : "Could not load relief stock history.",
      );
    }
  }

  async function loadAidRequests() {
    setAidLoadState("loading");
    setAidError(null);
    try {
      const data = await fetchAidRequests();
      setAidRequests(data);
      setAidLoadState(data.length > 0 ? "ready" : "empty");
      setSelectedAidRequestId((current) => current ?? data[0]?.id ?? null);
    } catch (error) {
      setAidRequests([]);
      setAidError(
        error instanceof Error ? error.message : "Could not load aid requests.",
      );
      setAidLoadState("error");
    }
  }

  async function loadCapacity(lat: number, lng: number) {
    setCapacityLoadState("loading");
    const [nearbyShelters, nearbyHospitals, nearbyClosures, nearbyRelief] =
      await Promise.all([
        fetchShelters(lat, lng),
        fetchHospitalCapacity(lat, lng),
        fetchRoadClosures(lat, lng),
        fetchNearbyReliefPoints(lat, lng),
      ]);
    setShelters(nearbyShelters);
    setHospitals(nearbyHospitals);
    setRoadClosures(nearbyClosures);
    setNearbyReliefPoints(nearbyRelief);
    setCapacityLoadState(
      nearbyShelters.length > 0 ||
        nearbyHospitals.length > 0 ||
        nearbyClosures.length > 0 ||
        nearbyRelief.length > 0
        ? "ready"
        : "empty",
    );
  }

  useEffect(() => {
    void loadIncidents();
    void loadReliefPoints();
    void loadAidRequests();
    void loadCapacity(
      DEFAULT_CAPACITY_LOCATION.lat,
      DEFAULT_CAPACITY_LOCATION.lng,
    );
    // Load once on mount; the shell only mounts this hook after sign-in.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const filteredIncidents = useMemo(
    () => incidents.filter((incident) => matchesFilters(incident, filters)),
    [incidents, filters],
  );

  const selectedIncident = useMemo(
    () =>
      incidents.find((incident) => incident.id === selectedIncidentId) ?? null,
    [incidents, selectedIncidentId],
  );

  const selectedReliefPoint = useMemo(
    () =>
      reliefPoints.find((point) => point.id === selectedReliefPointId) ?? null,
    [reliefPoints, selectedReliefPointId],
  );

  const selectedAidRequest = useMemo(
    () =>
      aidRequests.find((request) => request.id === selectedAidRequestId) ??
      null,
    [aidRequests, selectedAidRequestId],
  );

  useEffect(() => {
    if (!selectedIncident) return;
    setStatusForm({ ...initialStatusForm, status: selectedIncident.status });
    void loadCapacity(
      selectedIncident.location.lat,
      selectedIncident.location.lng,
    );
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedIncident?.id]);

  useEffect(() => {
    if (!selectedReliefPoint) {
      setReliefHistory([]);
      return;
    }
    setReliefForm(reliefPointToForm(selectedReliefPoint));
    void loadReliefHistory(selectedReliefPoint.id);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedReliefPoint?.id]);

  useEffect(() => {
    if (!selectedAidRequest) {
      return;
    }
    setAidForm(aidRequestToForm(selectedAidRequest));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedAidRequest?.id]);

  function selectIncident(incidentId: string) {
    setSelectedIncidentId(incidentId);
    setStatusUpdateState("idle");
    setStatusUpdateError(null);
  }

  function deselectIncident() {
    setSelectedIncidentId(null);
    setStatusUpdateState("idle");
    setStatusUpdateError(null);
  }

  function selectReliefPoint(reliefPointId: string) {
    setSelectedReliefPointId(reliefPointId);
    setReliefUpdateState("idle");
    setReliefError(null);
  }

  function deselectReliefPoint() {
    setSelectedReliefPointId(null);
    setReliefUpdateState("idle");
    setReliefError(null);
  }

  function selectAidRequest(aidRequestId: string) {
    setSelectedAidRequestId(aidRequestId);
    setAidUpdateState("idle");
    setAidError(null);
  }

  function deselectAidRequest() {
    setSelectedAidRequestId(null);
    setAidUpdateState("idle");
    setAidError(null);
  }

  async function handleStatusUpdate() {
    if (!selectedIncident) return;
    setStatusUpdateState("loading");
    setStatusUpdateError(null);
    try {
      const updated = await patchIncidentStatus(selectedIncident.id, {
        note: statusForm.note,
        resolutionNotes: statusForm.resolutionNotes,
        status: statusForm.status,
      });
      setIncidents((current) =>
        current.map((incident) =>
          incident.id === updated.id ? updated : incident,
        ),
      );
      setStatusForm({ ...initialStatusForm, status: updated.status });
      setStatusUpdateState("success");
    } catch (error) {
      setStatusUpdateError(
        error instanceof Error ? error.message : "Status update failed.",
      );
      setStatusUpdateState("error");
    }
  }

  function selectShelterTarget(shelterId: string) {
    setSelectedShelterId(shelterId);
    setCapacityUpdateState("idle");
    setCapacityUpdateError(null);
    const shelter = shelters.find((item) => item.id === shelterId);
    if (shelter) {
      setShelterForm(shelterToOccupancyForm(shelter));
    }
  }

  function selectHospitalTarget(facilityId: string) {
    setSelectedHospitalId(facilityId);
    setCapacityUpdateState("idle");
    setCapacityUpdateError(null);
    const facility = hospitals.find((item) => item.id === facilityId);
    if (facility) {
      setHospitalForm(hospitalToCapacityForm(facility));
    }
  }

  /** Parse a numeric form field; empty strings stay undefined (no change). */
  function optionalCount(value: string): number | undefined {
    const trimmed = value.trim();
    if (!trimmed) {
      return undefined;
    }
    const parsed = Number.parseInt(trimmed, 10);
    return Number.isNaN(parsed) ? undefined : parsed;
  }

  /** Returns true when the update succeeded so the view can close the dialog. */
  async function handleShelterOccupancyUpdate(): Promise<boolean> {
    if (!selectedShelterId) {
      setCapacityUpdateError("Choose a shelter to update.");
      setCapacityUpdateState("error");
      return false;
    }
    if (!canManageShelterResources(session)) {
      setCapacityUpdateError(
        "Your role does not have permission to update shelter occupancy.",
      );
      setCapacityUpdateState("error");
      return false;
    }
    setCapacityUpdateState("loading");
    setCapacityUpdateError(null);
    try {
      const request: ShelterOccupancyUpdateRequest = {
        status: shelterForm.status as ShelterRecord["status"],
      };
      const capacity = optionalCount(shelterForm.capacity);
      const currentOccupancy = optionalCount(shelterForm.currentOccupancy);
      if (capacity !== undefined) {
        request.capacity = capacity;
      }
      if (currentOccupancy !== undefined) {
        request.currentOccupancy = currentOccupancy;
      }
      if (shelterForm.notes.trim()) {
        request.notes = shelterForm.notes.trim();
      }
      const updated = await updateShelterOccupancy(selectedShelterId, request);
      setShelters((current) =>
        current.map((shelter) =>
          shelter.id === updated.id ? updated : shelter,
        ),
      );
      setCapacityUpdateState("success");
      return true;
    } catch (error) {
      // Surfaces the service's 400 invalid_occupancy message when the merged
      // occupancy would exceed capacity.
      setCapacityUpdateError(
        error instanceof Error ? error.message : "Shelter update failed.",
      );
      setCapacityUpdateState("error");
      return false;
    }
  }

  /** Returns true when the update succeeded so the view can close the dialog. */
  async function handleHospitalCapacityUpdate(): Promise<boolean> {
    if (!selectedHospitalId) {
      setCapacityUpdateError("Choose a hospital to update.");
      setCapacityUpdateState("error");
      return false;
    }
    if (!canManageShelterResources(session)) {
      setCapacityUpdateError(
        "Your role does not have permission to update hospital capacity.",
      );
      setCapacityUpdateState("error");
      return false;
    }
    setCapacityUpdateState("loading");
    setCapacityUpdateError(null);
    try {
      const request: HospitalCapacityUpdateRequest = {
        emergencyCapacity: hospitalForm.emergencyCapacity,
        oxygenAvailable: hospitalForm.oxygenAvailable,
        source: "manual",
      };
      const totalBeds = optionalCount(hospitalForm.totalBeds);
      const availableBeds = optionalCount(hospitalForm.availableBeds);
      const icuBedsAvailable = optionalCount(hospitalForm.icuBedsAvailable);
      const maternityBedsAvailable = optionalCount(
        hospitalForm.maternityBedsAvailable,
      );
      const pediatricBedsAvailable = optionalCount(
        hospitalForm.pediatricBedsAvailable,
      );
      const isolationBedsAvailable = optionalCount(
        hospitalForm.isolationBedsAvailable,
      );
      const ambulancesAvailable = optionalCount(
        hospitalForm.ambulancesAvailable,
      );
      if (totalBeds !== undefined) {
        request.totalBeds = totalBeds;
      }
      if (availableBeds !== undefined) {
        request.availableBeds = availableBeds;
      }
      if (icuBedsAvailable !== undefined) {
        request.icuBedsAvailable = icuBedsAvailable;
      }
      if (maternityBedsAvailable !== undefined) {
        request.maternityBedsAvailable = maternityBedsAvailable;
      }
      if (pediatricBedsAvailable !== undefined) {
        request.pediatricBedsAvailable = pediatricBedsAvailable;
      }
      if (isolationBedsAvailable !== undefined) {
        request.isolationBedsAvailable = isolationBedsAvailable;
      }
      if (ambulancesAvailable !== undefined) {
        request.ambulancesAvailable = ambulancesAvailable;
      }
      if (hospitalForm.emergencyUnitStatus.trim()) {
        request.emergencyUnitStatus =
          hospitalForm.emergencyUnitStatus as HospitalCapacityUpdateRequest["emergencyUnitStatus"];
      }
      if (hospitalForm.notes.trim()) {
        request.notes = hospitalForm.notes.trim();
      }
      const updated = await updateHospitalCapacity(
        selectedHospitalId,
        request,
      );
      setHospitals((current) =>
        current.map((facility) =>
          facility.id === updated.id ? { ...facility, ...updated } : facility,
        ),
      );
      setCapacityUpdateState("success");
      return true;
    } catch (error) {
      setCapacityUpdateError(
        error instanceof Error ? error.message : "Hospital update failed.",
      );
      setCapacityUpdateState("error");
      return false;
    }
  }

  async function handleSaveReliefPoint() {
    if (!canManageShelterResources(session)) {
      setReliefError(
        "Your role does not have permission to manage relief points.",
      );
      setReliefUpdateState("error");
      return;
    }
    setReliefUpdateState("loading");
    setReliefError(null);
    try {
      const location = {
        lat: Number.parseFloat(reliefForm.lat),
        lng: Number.parseFloat(reliefForm.lng),
      };
      const stockCategories = parseStockCategories(reliefForm.stockCategories);
      const request = {
        address: reliefForm.address.trim(),
        contact: reliefForm.contact.trim(),
        district: reliefForm.district.trim(),
        eligibility: reliefForm.eligibility.trim(),
        location,
        name: reliefForm.name.trim(),
        operatingHours: reliefForm.operatingHours.trim(),
        region: reliefForm.region.trim(),
        schedule: reliefForm.schedule.trim(),
        sourceRef: session.agencyId,
        status: reliefForm.status,
        stockCategories,
        type: reliefForm.type,
      };
      const saved = selectedReliefPoint
        ? await updateReliefPoint(
            selectedReliefPoint.id,
            request satisfies UpdateReliefPointRequest,
          )
        : await createReliefPoint({
            ...request,
            source: "manual",
          } satisfies CreateReliefPointRequest);
      setReliefPoints((current) =>
        current.some((point) => point.id === saved.id)
          ? current.map((point) => (point.id === saved.id ? saved : point))
          : [saved, ...current],
      );
      setSelectedReliefPointId(saved.id);
      setReliefForm(reliefPointToForm(saved));
      await loadReliefHistory(saved.id);
      setReliefUpdateState("success");
    } catch (error) {
      setReliefError(
        error instanceof Error ? error.message : "Relief point update failed.",
      );
      setReliefUpdateState("error");
    }
  }

  function handleNewReliefPoint() {
    setSelectedReliefPointId(null);
    setReliefForm(initialReliefPointForm);
    setReliefHistory([]);
    setReliefUpdateState("idle");
    setReliefError(null);
  }

  async function handleCreateAidRequest() {
    if (!canManageShelterResources(session)) {
      setAidError("Your role does not have permission to create aid requests.");
      setAidUpdateState("error");
      return;
    }
    setAidUpdateState("loading");
    setAidError(null);
    try {
      const quantityNeeded = Number.parseInt(aidForm.quantityNeeded, 10);
      const neededBy = new Date(aidForm.neededBy).toISOString();
      const request = {
        category: aidForm.category,
        contact: aidForm.contact.trim(),
        description: aidForm.description.trim(),
        district: aidForm.district.trim(),
        location: {
          lat: Number.parseFloat(aidForm.lat),
          lng: Number.parseFloat(aidForm.lng),
        },
        neededBy,
        priority: aidForm.priority,
        quantityNeeded,
        quantityUnit: aidForm.quantityUnit.trim(),
        receivingOrganization: aidForm.receivingOrganization.trim(),
        region: aidForm.region.trim(),
        sourceReliefPointId: aidForm.sourceReliefPointId.trim() || undefined,
        title: aidForm.title.trim(),
        visibility: aidForm.visibility,
      } satisfies CreateAidRequestRequest;
      const created = await createAidRequest(request);
      setAidRequests((current) => [created, ...current]);
      setSelectedAidRequestId(created.id);
      setAidForm(aidRequestToForm(created));
      setAidUpdateState("success");
    } catch (error) {
      setAidError(
        error instanceof Error ? error.message : "Aid request create failed.",
      );
      setAidUpdateState("error");
    }
  }

  async function handleReviewAidRequest(
    status: ReviewAidRequestRequest["status"],
  ) {
    if (!selectedAidRequest) return;
    if (!canManageShelterResources(session)) {
      setAidError("Your role does not have permission to review aid requests.");
      setAidUpdateState("error");
      return;
    }
    setAidUpdateState("loading");
    setAidError(null);
    try {
      const reviewed = await reviewAidRequest(selectedAidRequest.id, {
        antiFraudNotes:
          status === "approved" || status === "open"
            ? "Receiving organization, contact, and category checked by agency operator."
            : "Reviewed by agency operator.",
        approvalNotes:
          status === "approved" || status === "open"
            ? "Approved for partner/public aid listing."
            : "Status updated by agency operator.",
        status,
      });
      setAidRequests((current) =>
        current.map((request) =>
          request.id === reviewed.id ? reviewed : request,
        ),
      );
      setAidForm(aidRequestToForm(reviewed));
      setAidUpdateState("success");
    } catch (error) {
      setAidError(
        error instanceof Error ? error.message : "Aid request review failed.",
      );
      setAidUpdateState("error");
    }
  }

  function handleNewAidRequest() {
    setSelectedAidRequestId(null);
    // Computed now, at dialog-open time, so the default neededBy never expires.
    setAidForm(initialAidRequestForm());
    setAidUpdateState("idle");
    setAidError(null);
  }

  async function handleAidExport() {
    if (!canManageShelterResources(session)) {
      setAidError("Your role does not have permission to export aid reports.");
      return;
    }
    setAidError(null);
    try {
      const response = await fetch(
        `${SHELTER_API_BASE}/aid-requests/report.csv`,
        { headers: agencyHeaders() },
      );
      if (!response.ok) {
        throw new Error(await extractError(response));
      }
      const blob = await response.blob();
      const href = URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = href;
      link.download = "nadaa-aid-requests.csv";
      document.body.appendChild(link);
      link.click();
      // Revoke asynchronously: a synchronous revoke can cancel the download
      // before the browser has consumed the blob.
      window.setTimeout(() => {
        URL.revokeObjectURL(href);
        link.remove();
      }, 0);
    } catch (error) {
      setAidError(
        error instanceof Error ? error.message : "Aid request export failed.",
      );
    }
  }

  const metrics = useMemo(() => {
    const isOpen = (status: IncidentRecord["status"]) =>
      !["closed", "false_report"].includes(status);
    return {
      assigned: incidents.filter((incident) => incident.status === "assigned")
        .length,
      enRoute: incidents.filter(
        (incident) => incident.status === "response_en_route",
      ).length,
      onScene: incidents.filter((incident) => incident.status === "on_scene")
        .length,
      contained: incidents.filter((incident) => incident.status === "contained")
        .length,
      recovery: incidents.filter(
        (incident) => incident.status === "recovery_ongoing",
      ).length,
      priority: incidents.filter(
        (incident) => incident.priorityReview && isOpen(incident.status),
      ).length,
      open: incidents.filter((incident) => isOpen(incident.status)).length,
      reliefOpen: reliefPoints.filter((point) =>
        ["open", "limited"].includes(point.status),
      ).length,
      aidOpen: aidRequests.filter((request) =>
        ["approved", "open", "partially_matched"].includes(request.status),
      ).length,
      aidPending: aidRequests.filter(
        (request) => request.status === "pending_review",
      ).length,
    };
  }, [aidRequests, incidents, reliefPoints]);

  const sheltersCritical = useMemo(
    () =>
      shelters.filter(
        (shelter) =>
          shelter.status === "full" ||
          (shelter.capacity > 0 &&
            shelter.currentOccupancy / shelter.capacity >= 0.9),
      ).length,
    [shelters],
  );

  return {
    session,
    // Incidents
    incidents,
    filteredIncidents,
    incidentLoadState,
    incidentError,
    filters,
    setFilters,
    selectedIncidentId,
    selectedIncident,
    selectIncident,
    deselectIncident,
    loadIncidents,
    // Status update
    statusForm,
    setStatusForm,
    statusUpdateState,
    statusUpdateError,
    handleStatusUpdate,
    // Capacity
    shelters,
    hospitals,
    roadClosures,
    nearbyReliefPoints,
    capacityLoadState,
    sheltersCritical,
    shelterForm,
    setShelterForm,
    hospitalForm,
    setHospitalForm,
    selectedShelterId,
    selectShelterTarget,
    selectedHospitalId,
    selectHospitalTarget,
    capacityUpdateState,
    capacityUpdateError,
    handleShelterOccupancyUpdate,
    handleHospitalCapacityUpdate,
    // Relief
    reliefPoints,
    selectedReliefPointId,
    selectedReliefPoint,
    reliefForm,
    setReliefForm,
    reliefHistory,
    reliefLoadState,
    reliefUpdateState,
    reliefError,
    loadReliefPoints,
    selectReliefPoint,
    deselectReliefPoint,
    handleSaveReliefPoint,
    handleNewReliefPoint,
    // Aid
    aidRequests,
    selectedAidRequestId,
    selectedAidRequest,
    aidForm,
    setAidForm,
    aidLoadState,
    aidUpdateState,
    aidError,
    loadAidRequests,
    selectAidRequest,
    deselectAidRequest,
    handleCreateAidRequest,
    handleReviewAidRequest,
    handleNewAidRequest,
    handleAidExport,
    // Derived
    metrics,
  };
}

export type AgencyData = ReturnType<typeof useAgencyData>;
