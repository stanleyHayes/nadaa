import { useEffect, useMemo, useState } from "react";
import type {
  AreaRiskResponse,
  CitizenAlertFeedItem,
  EmergencyGuideRecord,
  NearbyShelterResponse,
  VolunteerProfile,
  VolunteerTaskRecord,
  VolunteerTaskStatus,
} from "@nadaa/shared-types";
import {
  fallbackRisk,
  fetchAlertFeed,
  fetchAreaRisk,
  fetchGuides,
  fetchNearbyShelters,
  fetchVolunteerTasks,
  registerVolunteerProfile,
  registerPushToken,
  submitIncidentDraft,
  submitVolunteerObservation as submitVolunteerObservationAPI,
  updateVolunteerTaskStatus,
} from "./api";
import {
  buildFallbackAlerts,
  buildFallbackGuides,
  initialPermissions,
  initialReportDraft,
  mobileAreaPresets,
  sampleShelters,
  sampleVolunteerProfile,
  sampleVolunteerTasks,
} from "./data";
import {
  createMemoryStorage,
  readGuideCache,
  readReportDraft,
  readSession,
  readVolunteerProfile,
  readVolunteerTasks,
  writeGuideCache,
  writeReportDraft,
  writeSession,
  writeVolunteerProfile,
  writeVolunteerTasks,
} from "./offline";
import { nextPermissionStatus, permissionMessage } from "./permissions";
import type {
  AlertView,
  MobileLoadState,
  MobilePermissionState,
  PushRegistrationState,
  ReportDraft,
  VolunteerObservationDraft,
} from "./types";

const storage = createMemoryStorage();

export function useCitizenMobileState() {
  const [alerts, setAlerts] = useState<CitizenAlertFeedItem[]>(() =>
    buildFallbackAlerts(),
  );
  const [alertView, setAlertView] = useState<AlertView>("current");
  const [guides, setGuides] = useState<EmergencyGuideRecord[]>(() =>
    buildFallbackGuides(),
  );
  const [loadState, setLoadState] = useState<MobileLoadState>({
    status: "idle",
    message: "Mobile foundation ready.",
  });
  const [permissions, setPermissions] =
    useState<MobilePermissionState>(initialPermissions);
  const [pushState, setPushState] = useState<PushRegistrationState>({
    status: "permission_needed",
    message: "Allow notifications to receive urgent NADAA warnings.",
  });
  const [reportDraft, setReportDraft] =
    useState<ReportDraft>(initialReportDraft);
  const [risk, setRisk] = useState<AreaRiskResponse>(() => fallbackRisk());
  const [selectedArea, setSelectedArea] = useState(mobileAreaPresets[0]);
  const [shelters, setShelters] =
    useState<NearbyShelterResponse>(sampleShelters);
  const [session, setSession] = useState({
    contactPermission: true,
    isGuest: true,
    name: "Guest citizen",
    phone: "+233200000000",
    preferredLanguage: "en",
    userId: "usr_mobile_guest",
  });
  const [volunteerObservation, setVolunteerObservation] =
    useState<VolunteerObservationDraft>({
      escalationRequested: false,
      note: "",
      safetyStatus: "safe",
    });
  const [volunteerProfile, setVolunteerProfile] = useState<VolunteerProfile>(
    sampleVolunteerProfile,
  );
  const [volunteerTasks, setVolunteerTasks] =
    useState<VolunteerTaskRecord[]>(sampleVolunteerTasks);

  useEffect(() => {
    void hydrate();
  }, []);

  const visibleAlerts = useMemo(
    () =>
      alerts.filter((alert) =>
        alertView === "all" ? true : alert.status === alertView,
      ),
    [alertView, alerts],
  );
  const currentAlertCount = useMemo(
    () => alerts.filter((alert) => alert.status === "current").length,
    [alerts],
  );
  const offlineGuideCount = useMemo(
    () => guides.filter((guide) => guide.offlineAvailable).length,
    [guides],
  );
  const activeVolunteerTaskCount = useMemo(
    () =>
      volunteerTasks.filter(
        (task) => !["completed", "cancelled"].includes(task.status),
      ).length,
    [volunteerTasks],
  );

  async function hydrate() {
    setLoadState({ status: "loading", message: "Loading saved mobile state" });
    const [cachedGuides, savedDraft, savedSession, savedVolunteer, savedTasks] =
      await Promise.all([
        readGuideCache(storage),
        readReportDraft(storage),
        readSession(storage),
        readVolunteerProfile(storage),
        readVolunteerTasks(storage),
      ]);
    setGuides(cachedGuides.guides);
    setReportDraft(savedDraft);
    setSession(savedSession);
    setVolunteerProfile(savedVolunteer);
    setVolunteerTasks(savedTasks);
    setLoadState({
      status: "success",
      message: `Offline guides ready from ${formatDate(cachedGuides.cachedAt)}.`,
    });
  }

  async function refreshAll() {
    setLoadState({ status: "loading", message: "Refreshing citizen mobile" });
    try {
      const [nextRisk, nextAlerts, nextGuides, nextShelters, nextTasks] =
        await Promise.all([
          fetchAreaRisk(selectedArea.lat, selectedArea.lng),
          fetchAlertFeed(),
          fetchGuides(session.preferredLanguage),
          fetchNearbyShelters(selectedArea.lat, selectedArea.lng),
          fetchVolunteerTasks(volunteerProfile.id),
        ]);
      setRisk(nextRisk);
      setAlerts(nextAlerts);
      setGuides(nextGuides);
      setShelters(nextShelters);
      setVolunteerTasks(nextTasks);
      await writeGuideCache(storage, {
        cachedAt: new Date().toISOString(),
        guides: nextGuides,
        language: session.preferredLanguage,
      });
      await writeVolunteerTasks(storage, nextTasks);
      setLoadState({
        status: "success",
        message:
          "Alerts, risk, guides, shelters, and volunteer tasks refreshed.",
      });
    } catch (error) {
      setLoadState({
        status: "offline",
        message:
          error instanceof Error
            ? error.message
            : "Network unavailable. Showing saved content.",
      });
    }
  }

  async function saveDraft(nextDraft: ReportDraft) {
    setReportDraft(nextDraft);
    await writeReportDraft(storage, nextDraft);
    setLoadState({ status: "success", message: "Report draft saved offline." });
  }

  async function submitDraft() {
    setLoadState({ status: "loading", message: "Submitting incident report" });
    try {
      const response = await submitIncidentDraft(reportDraft, session);
      setReportDraft(initialReportDraft);
      await writeReportDraft(storage, initialReportDraft);
      setLoadState({
        status: "success",
        message: `Report sent: ${response.reference}`,
      });
    } catch (error) {
      await writeReportDraft(storage, reportDraft);
      setLoadState({
        status: "error",
        message:
          error instanceof Error
            ? `${error.message} Draft saved for retry.`
            : "Could not submit. Draft saved for retry.",
      });
    }
  }

  async function updateSessionPhone(phone: string) {
    const nextSession = { ...session, isGuest: false, phone };
    setSession(nextSession);
    await writeSession(storage, nextSession);
  }

  async function registerVolunteer() {
    setLoadState({
      status: "loading",
      message: "Registering community volunteer profile",
    });
    const response = await registerVolunteerProfile(session);
    setVolunteerProfile(response.volunteer);
    await writeVolunteerProfile(storage, response.volunteer);
    const tasks = await fetchVolunteerTasks(response.volunteer.id);
    setVolunteerTasks(tasks);
    await writeVolunteerTasks(storage, tasks);
    setLoadState({
      status: "success",
      message: `Volunteer group ready: ${response.volunteer.community}.`,
    });
  }

  async function refreshVolunteerTasks() {
    setLoadState({ status: "loading", message: "Refreshing volunteer tasks" });
    const tasks = await fetchVolunteerTasks(volunteerProfile.id);
    setVolunteerTasks(tasks);
    await writeVolunteerTasks(storage, tasks);
    setLoadState({
      status: "success",
      message: `${tasks.length} volunteer tasks ready.`,
    });
  }

  async function updateVolunteerStatus(
    taskId: string,
    status: Exclude<VolunteerTaskStatus, "assigned">,
  ) {
    setLoadState({
      status: "loading",
      message: `Updating volunteer task to ${status}`,
    });
    const task = await updateVolunteerTaskStatus(taskId, {
      note: `Mobile volunteer status update: ${status}`,
      safetyStatus: status === "needs_escalation" ? "needs_authority" : "safe",
      status,
      volunteerId: volunteerProfile.id,
    });
    await replaceVolunteerTask(task);
    setLoadState({
      status: status === "needs_escalation" ? "offline" : "success",
      message:
        status === "needs_escalation"
          ? "Escalation marked. Call 112 if anyone is in danger."
          : `Volunteer task updated to ${status}.`,
    });
  }

  async function saveVolunteerObservation(
    nextDraft: VolunteerObservationDraft,
  ) {
    setVolunteerObservation(nextDraft);
  }

  async function submitVolunteerObservation(taskId: string) {
    if (volunteerObservation.note.trim().length < 5) {
      setLoadState({
        status: "error",
        message: "Add a short field observation before submitting.",
      });
      return;
    }
    setLoadState({
      status: "loading",
      message: "Submitting volunteer observation",
    });
    const task = await submitVolunteerObservationAPI(taskId, {
      escalationRequested: volunteerObservation.escalationRequested,
      observation: volunteerObservation.note.trim(),
      safetyStatus: volunteerObservation.safetyStatus,
      volunteerId: volunteerProfile.id,
    });
    await replaceVolunteerTask(task);
    setVolunteerObservation({
      escalationRequested: false,
      note: "",
      safetyStatus: "safe",
    });
    setLoadState({
      status: task.escalationRequired ? "offline" : "success",
      message: task.escalationRequired
        ? "Observation submitted with authority escalation."
        : "Observation submitted.",
    });
  }

  async function replaceVolunteerTask(task: VolunteerTaskRecord) {
    const nextTasks = [
      task,
      ...volunteerTasks.filter((item) => item.id !== task.id),
    ];
    setVolunteerTasks(nextTasks);
    await writeVolunteerTasks(storage, nextTasks);
  }

  async function togglePermission(key: keyof MobilePermissionState) {
    const nextStatus = nextPermissionStatus(permissions[key]);
    const nextPermissions = { ...permissions, [key]: nextStatus };
    setPermissions(nextPermissions);
    setLoadState({
      status: nextStatus === "granted" ? "success" : "idle",
      message: permissionMessage(key, nextStatus),
    });
    if (key === "push") {
      setPushState(await registerPushToken(nextStatus === "granted"));
    }
  }

  function chooseArea(index: number) {
    const area = mobileAreaPresets[index] ?? mobileAreaPresets[0];
    setSelectedArea(area);
  }

  return {
    actions: {
      chooseArea,
      refreshAll,
      refreshVolunteerTasks,
      registerVolunteer,
      saveDraft,
      saveVolunteerObservation,
      setAlertView,
      submitDraft,
      submitVolunteerObservation,
      togglePermission,
      updateSessionPhone,
      updateVolunteerStatus,
    },
    state: {
      activeVolunteerTaskCount,
      alertView,
      alerts,
      currentAlertCount,
      guides,
      loadState,
      offlineGuideCount,
      permissions,
      pushState,
      reportDraft,
      risk,
      selectedArea,
      session,
      shelters,
      visibleAlerts,
      volunteerObservation,
      volunteerProfile,
      volunteerTasks,
    },
  };
}

function formatDate(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return new Intl.DateTimeFormat("en-GH", {
    day: "numeric",
    month: "short",
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
}
