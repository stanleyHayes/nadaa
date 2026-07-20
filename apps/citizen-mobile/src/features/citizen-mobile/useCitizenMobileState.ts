import { useEffect, useMemo, useRef, useState } from "react";
import AsyncStorage from "@react-native-async-storage/async-storage";
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
  fetchAlertFeed,
  fetchAreaRisk,
  fetchGuides,
  fetchNearbyShelters,
  fetchVolunteerTasks,
  isAuthError,
  loginCitizen,
  registerCitizen,
  registerVolunteerProfile,
  registerPushToken,
  requestCitizenLoginOtp,
  submitIncidentDraft,
  submitVolunteerObservation as submitVolunteerObservationAPI,
  updateVolunteerTaskStatus,
} from "./api";
import {
  buildFallbackGuides,
  emptyShelters,
  initialPermissions,
  initialReportDraft,
  initialSession,
  initialSignIn,
  initialVolunteerRegistration,
  mobileAreaPresets,
} from "./data";
import type { KeyValueStorage } from "./offline";
import {
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
import {
  permissionMessage,
  readDevicePosition,
  requestOSPermission,
} from "./permissions";
import {
  configureAlertNotifications,
  getAlertPermissionStatus,
  notifyNewAlerts,
  requestAlertPermission,
} from "./alertNotifications";
import type {
  AlertView,
  MobileLoadState,
  MobilePermissionState,
  MobileSession,
  PushRegistrationState,
  ReportDraft,
  SignInDraft,
  VolunteerObservationDraft,
  VolunteerRegistrationDraft,
} from "./types";

// Drafts, session, guide cache, and volunteer data persist across cold starts.
const storage: KeyValueStorage = AsyncStorage;

const E164_PHONE = /^\+[1-9]\d{7,14}$/;

function errorMessage(error: unknown, fallback: string) {
  return error instanceof Error ? error.message : fallback;
}

export function useCitizenMobileState() {
  const [alerts, setAlerts] = useState<CitizenAlertFeedItem[]>([]);
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
  // Live-only: risk starts empty and is loaded by refreshAll on startup — a
  // fixture is never rendered as if it were live data.
  const [risk, setRisk] = useState<AreaRiskResponse | null>(null);
  const [selectedArea, setSelectedArea] = useState(mobileAreaPresets[0]);
  const [session, setSession] = useState<MobileSession>(initialSession);
  const [shelters, setShelters] =
    useState<NearbyShelterResponse>(emptyShelters);
  const [signIn, setSignIn] = useState<SignInDraft>(initialSignIn);
  const [volunteerObservation, setVolunteerObservation] =
    useState<VolunteerObservationDraft>({
      escalationRequested: false,
      note: "",
      safetyStatus: "safe",
    });
  const [volunteerProfile, setVolunteerProfile] =
    useState<VolunteerProfile | null>(null);
  const [volunteerRegistration, setVolunteerRegistration] =
    useState<VolunteerRegistrationDraft>(initialVolunteerRegistration);
  const [volunteerTasks, setVolunteerTasks] = useState<VolunteerTaskRecord[]>(
    [],
  );

  const seenAlertIds = useRef<Set<string>>(new Set());
  const alertsSeeded = useRef(false);

  useEffect(() => {
    void hydrate();
    void configureAlertNotifications();
    void hydratePushPermission();
    // Load the live feed on cold start so the home screen never sits on
    // empty/placeholder data until a manual refresh.
    void refreshAll();
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

  // #31: sync the push permission from the OS so a persisted grant keeps
  // notifications working after an app restart.
  async function hydratePushPermission() {
    try {
      const status = await getAlertPermissionStatus();
      setPermissions((current) => ({ ...current, push: status }));
    } catch {
      // Leave push as "unknown" when the OS status cannot be read.
    }
  }

  async function refreshAll(area = selectedArea) {
    setLoadState({ status: "loading", message: "Refreshing citizen mobile" });
    try {
      const [nextRisk, nextAlerts, guideResult, nextShelters] =
        await Promise.all([
          fetchAreaRisk(area.lat, area.lng),
          fetchAlertFeed(),
          fetchGuides(session.preferredLanguage),
          fetchNearbyShelters(area.lat, area.lng),
        ]);
      setRisk(nextRisk);
      setAlerts(nextAlerts);
      // Notify for newly-arrived current alerts (never on the first load).
      // The OS notification channel handles Do-Not-Disturb; a level-5 emergency
      // overrides it. Seed seen ids when we can't/shouldn't notify.
      if (alertsSeeded.current && permissions.push === "granted") {
        void notifyNewAlerts(nextAlerts, seenAlertIds.current).catch(
          () => undefined,
        );
      } else {
        nextAlerts
          .filter((alert) => alert.status === "current")
          .forEach((alert) => seenAlertIds.current.add(alert.id));
      }
      alertsSeeded.current = true;
      setShelters(nextShelters);
      // Only overwrite the guide cache when the fetch was live. On failure the
      // cached/bundled guides stay on screen and the status message says the
      // guides are the offline cache — never a false "refreshed".
      if (guideResult.live) {
        setGuides(guideResult.guides);
        await writeGuideCache(storage, {
          cachedAt: new Date().toISOString(),
          guides: guideResult.guides,
          language: session.preferredLanguage,
        });
      }
      // Volunteer tasks require a signed-in citizen token; skip without one.
      if (volunteerProfile && session.accessToken) {
        const nextTasks = await fetchVolunteerTasks(
          volunteerProfile.id,
          session.accessToken,
        );
        setVolunteerTasks(nextTasks);
        await writeVolunteerTasks(storage, nextTasks);
      }
      setLoadState({
        status: "success",
        message: guideResult.live
          ? "Alerts, risk, guides, and shelters refreshed."
          : "Alerts, risk, and shelters refreshed. Guides are showing the offline cache — the guide service could not be reached.",
      });
    } catch (error) {
      setLoadState({
        status: "offline",
        message: errorMessage(
          error,
          "Network unavailable. Showing saved content.",
        ),
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
        message: `${errorMessage(error, "Could not submit.")} Draft saved for retry.`,
      });
    }
  }

  function saveSignInDraft(nextDraft: SignInDraft) {
    setSignIn(nextDraft);
  }

  async function requestSignInCode() {
    const phone = signIn.phone.trim();
    if (!E164_PHONE.test(phone)) {
      setLoadState({
        status: "error",
        message:
          "Enter your phone in international format, for example +233201234567.",
      });
      return;
    }
    setLoadState({ status: "loading", message: "Sending sign-in code" });
    try {
      let challenge = await requestCitizenLoginOtp(phone);
      if (!challenge) {
        const name = signIn.name.trim();
        if (!name) {
          setLoadState({
            status: "error",
            message:
              "This phone is not registered yet. Enter your name to create a citizen account.",
          });
          return;
        }
        challenge = await registerCitizen({
          contactPermission: true,
          name,
          phone,
          preferredLanguage: session.preferredLanguage,
        });
      }
      setSignIn({
        ...signIn,
        challengeId: challenge.challengeId,
        devOtp: challenge.devOtp,
        phone,
      });
      setLoadState({
        status: "success",
        message: challenge.devOtp
          ? `Sign-in code sent via ${challenge.otpDelivery}. Dev code: ${challenge.devOtp}.`
          : `Sign-in code sent via ${challenge.otpDelivery}.`,
      });
    } catch (error) {
      setLoadState({
        status: "error",
        message: errorMessage(error, "Could not send the sign-in code."),
      });
    }
  }

  async function verifySignIn() {
    const phone = signIn.phone.trim();
    const otp = signIn.otp.trim();
    if (!otp) {
      setLoadState({
        status: "error",
        message: "Enter the sign-in code we sent to your phone.",
      });
      return;
    }
    setLoadState({ status: "loading", message: "Verifying sign-in code" });
    try {
      const result = await loginCitizen(phone, otp);
      const nextSession: MobileSession = {
        accessToken: result.accessToken,
        contactPermission: result.user.contactPermission,
        isGuest: false,
        name: result.user.name,
        phone: result.user.phone,
        preferredLanguage:
          result.user.preferredLanguage || session.preferredLanguage,
        userId: result.user.id,
      };
      setSession(nextSession);
      await writeSession(storage, nextSession);
      setSignIn(initialSignIn);
      setLoadState({
        status: "success",
        message: `Signed in as ${result.user.name}.`,
      });
    } catch (error) {
      setLoadState({
        status: "error",
        message: errorMessage(error, "Could not verify the sign-in code."),
      });
    }
  }

  async function registerVolunteer() {
    if (!session.accessToken) {
      setLoadState({
        status: "error",
        message:
          "Sign in with your citizen phone number before registering as a volunteer.",
      });
      return;
    }
    if (
      volunteerRegistration.community.trim().length < 2 ||
      volunteerRegistration.district.trim().length < 2 ||
      volunteerRegistration.region.trim().length < 2
    ) {
      setLoadState({
        status: "error",
        message:
          "Enter your community, district, and region so assignments reach the right area.",
      });
      return;
    }
    setLoadState({
      status: "loading",
      message: "Registering community volunteer profile",
    });
    try {
      const response = await registerVolunteerProfile(
        session,
        volunteerRegistration,
      );
      setVolunteerProfile(response.volunteer);
      await writeVolunteerProfile(storage, response.volunteer);
      const tasks = await fetchVolunteerTasks(
        response.volunteer.id,
        session.accessToken,
      );
      setVolunteerTasks(tasks);
      await writeVolunteerTasks(storage, tasks);
      setLoadState({
        status: "success",
        message: `Volunteer group ready: ${response.volunteer.community}.`,
      });
    } catch (error) {
      setLoadState({
        status: "error",
        message: isAuthError(error)
          ? "Your sign-in has expired. Sign in again on the Support tab, then retry."
          : `${errorMessage(error, "Could not register the volunteer profile.")} Check your connection and retry.`,
      });
    }
  }

  async function refreshVolunteerTasks() {
    if (!volunteerProfile || !session.accessToken) {
      setLoadState({
        status: "error",
        message:
          "Sign in and register a volunteer profile to load assignments.",
      });
      return;
    }
    setLoadState({ status: "loading", message: "Refreshing volunteer tasks" });
    try {
      const tasks = await fetchVolunteerTasks(
        volunteerProfile.id,
        session.accessToken,
      );
      setVolunteerTasks(tasks);
      await writeVolunteerTasks(storage, tasks);
      setLoadState({
        status: "success",
        message: `${tasks.length} volunteer tasks ready.`,
      });
    } catch (error) {
      setLoadState({
        status: "error",
        message: isAuthError(error)
          ? "Your sign-in has expired. Sign in again on the Support tab, then retry."
          : `${errorMessage(error, "Could not load volunteer tasks.")} Check your connection and retry.`,
      });
    }
  }

  async function updateVolunteerStatus(
    taskId: string,
    status: Exclude<VolunteerTaskStatus, "assigned">,
  ) {
    if (!volunteerProfile || !session.accessToken) {
      setLoadState({
        status: "error",
        message:
          "Sign in with your citizen phone number to update volunteer tasks.",
      });
      return;
    }
    setLoadState({
      status: "loading",
      message: `Updating volunteer task to ${status}`,
    });
    try {
      const task = await updateVolunteerTaskStatus(
        taskId,
        {
          note: `Mobile volunteer status update: ${status}`,
          safetyStatus:
            status === "needs_escalation" ? "needs_authority" : "safe",
          status,
          volunteerId: volunteerProfile.id,
        },
        session.accessToken,
      );
      await replaceVolunteerTask(task);
      setLoadState({
        status: status === "needs_escalation" ? "offline" : "success",
        message:
          status === "needs_escalation"
            ? "Escalation marked. Call 112 if anyone is in danger."
            : `Volunteer task updated to ${status}.`,
      });
    } catch (error) {
      setLoadState({
        status: "error",
        message: `${errorMessage(error, "Could not update the volunteer task.")} Check your connection and retry.`,
      });
    }
  }

  async function saveVolunteerObservation(
    nextDraft: VolunteerObservationDraft,
  ) {
    setVolunteerObservation(nextDraft);
  }

  function saveVolunteerRegistration(nextDraft: VolunteerRegistrationDraft) {
    setVolunteerRegistration(nextDraft);
  }

  async function submitVolunteerObservation(taskId: string) {
    if (volunteerObservation.note.trim().length < 5) {
      setLoadState({
        status: "error",
        message: "Add a short field observation before submitting.",
      });
      return;
    }
    if (!volunteerProfile || !session.accessToken) {
      setLoadState({
        status: "error",
        message:
          "Sign in with your citizen phone number to submit observations.",
      });
      return;
    }
    setLoadState({
      status: "loading",
      message: "Submitting volunteer observation",
    });
    try {
      const task = await submitVolunteerObservationAPI(
        taskId,
        {
          escalationRequested: volunteerObservation.escalationRequested,
          observation: volunteerObservation.note.trim(),
          safetyStatus: volunteerObservation.safetyStatus,
          volunteerId: volunteerProfile.id,
        },
        session.accessToken,
      );
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
    } catch (error) {
      setLoadState({
        status: "error",
        message: `${errorMessage(error, "Could not submit the observation.")} Nothing was sent — check your connection and retry.`,
      });
    }
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
    if (key === "push") {
      // Pin the permission to the REAL OS answer — otherwise notifications are
      // scheduled but silently dropped.
      setLoadState({
        status: "loading",
        message: "Checking notification permission",
      });
      const granted = await requestAlertPermission();
      const resolved = granted ? "granted" : "denied";
      setPermissions((current) => ({ ...current, push: resolved }));
      setLoadState({
        status: granted ? "success" : "idle",
        message: permissionMessage("push", resolved),
      });
      setPushState(await registerPushToken(granted));
      return;
    }
    // Camera/location/media request the real OS permission and pin the
    // displayed state to the actual OS answer.
    setLoadState({
      status: "loading",
      message: `Requesting ${key} permission`,
    });
    const resolved = await requestOSPermission(key);
    setPermissions((current) => ({ ...current, [key]: resolved }));
    if (key === "location" && resolved === "granted") {
      // Use the real device position for the risk/shelter fetch and to prefill
      // the report draft. When no reading is available the draft coordinates
      // stay empty — they are never filled with a hardcoded city.
      const position = await readDevicePosition();
      if (position) {
        const area = {
          label: "Current location",
          lat: position.lat,
          lng: position.lng,
        };
        setSelectedArea(area);
        const nextDraft: ReportDraft = {
          ...reportDraft,
          lat: position.lat.toFixed(6),
          lng: position.lng.toFixed(6),
        };
        setReportDraft(nextDraft);
        await writeReportDraft(storage, nextDraft);
        await refreshAll(area);
        return;
      }
    }
    setLoadState({
      status: resolved === "granted" ? "success" : "idle",
      message: permissionMessage(key, resolved),
    });
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
      requestSignInCode,
      saveDraft,
      saveSignInDraft,
      saveVolunteerObservation,
      saveVolunteerRegistration,
      setAlertView,
      submitDraft,
      submitVolunteerObservation,
      togglePermission,
      updateVolunteerStatus,
      verifySignIn,
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
      signIn,
      visibleAlerts,
      volunteerObservation,
      volunteerProfile,
      volunteerRegistration,
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
