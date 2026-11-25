import { useEffect, useRef, useState } from "react";
import {
  Camera,
  CameraOff,
  Mic,
  ShieldCheck,
  Smartphone,
  Volume2,
} from "lucide-react";
import { FilesetResolver, ObjectDetector } from "@mediapipe/tasks-vision";
import { Spinner } from "./ui";
import "../styles/proctor.css";

const WASM =
  "https://cdn.jsdelivr.net/npm/@mediapipe/tasks-vision@0.10.14/wasm";
const MODEL =
  "https://storage.googleapis.com/mediapipe-models/object_detector/efficientdet_lite0/float16/1/efficientdet_lite0.tflite";

// Tunable rules
const SAMPLE_MS = 500; // how often the frame and cumulative sound are evaluated
const AUDIO_POLL_MS = 50; // continuous audio sampling every 50ms so no speech/noise is missed
const NO_PERSON_SAMPLES = 5; // 5 x 500 ms = 2.5 seconds without a person in camera
const MULTIPLE_PEOPLE_SAMPLES = 2; // 2 x 500 ms = 1 second when 2+ people are in camera
const PHONE_SAMPLES = 2; // 2 x 500 ms = 1 second when mobile phone is detected
const LOUD_SAMPLES = 2; // 2 x 500 ms = 1 second of noise or talking detected
const LOUD_FLOOR = 0.022; // sensitive noise floor: catches normal talking and background voices
const LOUD_MARGIN = 0.012; // margin above room background level
const LOUD_CAP = 0.04; // cap on noise floor to prevent ambient noise from masking speech
const COOLDOWN_MS = 5000; // grace period after any warning
const ALERT_MS = 2500; // how long the red label stays on screen
const MAX_WARNINGS = 3;

// Detection thresholds
const PERSON_MIN_SCORE = 0.2; // allows detection when candidate is partially/half in frame
const SECOND_PERSON_MIN_SCORE = 0.38; // confident threshold to prevent false-positive extra person
const PHONE_MIN_SCORE = 0.28; // reliable threshold for detecting mobile phones

// One generic message for every kind of violation, so candidates are not told what triggered it.
const WARNING_MESSAGE =
  "Unusual activity detected. Stay alone in the frame, keep your surroundings quiet and keep phones out of view.";

function boxIoU(b1, b2) {
  if (!b1 || !b2) return 0;
  const x1 = Math.max(b1.originX, b2.originX);
  const y1 = Math.max(b1.originY, b2.originY);
  const x2 = Math.min(b1.originX + b1.width, b2.originX + b2.width);
  const y2 = Math.min(b1.originY + b1.height, b2.originY + b2.height);
  const w = Math.max(0, x2 - x1);
  const h = Math.max(0, y2 - y1);
  const inter = w * h;
  if (inter <= 0) return 0;
  const area1 = b1.width * b1.height;
  const area2 = b2.width * b2.height;
  return inter / (area1 + area2 - inter);
}

function isCenterInside(inner, outer) {
  if (!inner || !outer) return false;
  const cx = inner.originX + inner.width / 2;
  const cy = inner.originY + inner.height / 2;
  return (
    cx >= outer.originX &&
    cx <= outer.originX + outer.width &&
    cy >= outer.originY &&
    cy <= outer.originY + outer.height
  );
}

async function createDetector() {
  const fileset = await FilesetResolver.forVisionTasks(WASM);
  const options = (delegate) => ({
    baseOptions: { modelAssetPath: MODEL, delegate },
    runningMode: "VIDEO",
    scoreThreshold: 0.2,
    maxResults: 8,
    categoryAllowlist: ["person", "cell phone"],
  });
  try {
    return await ObjectDetector.createFromOptions(fileset, options("GPU"));
  } catch {
    return await ObjectDetector.createFromOptions(fileset, options("CPU"));
  }
}

const rms = (buf) => {
  let sum = 0;
  for (let i = 0; i < buf.length; i++) {
    const v = (buf[i] - 128) / 128;
    sum += v * v;
  }
  return Math.sqrt(sum / buf.length);
};

// Checks (1) exactly one person is in the camera frame (body detection, no eye or gaze tracking) and
// (2) the room is not loud (microphone VOLUME only: nothing is recorded, stored, uploaded or transcribed).
// onViolation(type) must report the warning to the backend and resolve to { warnings }.
export default function Proctor({
  initialWarnings = 0,
  onViolation,
  compact = false,
}) {
  const video = useRef(null);
  const [started, setStarted] = useState(false);
  const [status, setStatus] = useState("loading"); // loading | ready | denied | error
  const [attempt, setAttempt] = useState(0);
  const [people, setPeople] = useState(null);
  const [hasPhone, setHasPhone] = useState(false);
  const [micOn, setMicOn] = useState(false);
  const [loud, setLoud] = useState(false);
  const [warnings, setWarnings] = useState(initialWarnings);
  const [alert, setAlert] = useState(null);

  const warnRef = useRef(initialWarnings);
  const violationRef = useRef(onViolation);
  useEffect(() => {
    violationRef.current = onViolation;
  });
  const alertTimer = useRef(null);

  useEffect(() => {
    if (!started) return;
    let stopped = false;
    let stream = null;
    let detector = null;
    let audioCtx = null;
    let analyser = null;
    let timer = null;
    let audioTimer = null;
    let noPersonRun = 0;
    let multiplePeopleRun = 0;
    let phoneRun = 0;
    let loudRun = 0;
    let calib = [];
    let floor = LOUD_FLOOR;
    let quietUntil = 0;
    let peakAudio = 0;
    const samples = new Uint8Array(1024);

    const raise = (type) => {
      warnRef.current += 1;
      setWarnings(warnRef.current);
      setAlert({
        message: WARNING_MESSAGE,
        n: Math.min(warnRef.current, MAX_WARNINGS),
      });
      clearTimeout(alertTimer.current);
      alertTimer.current = setTimeout(() => setAlert(null), ALERT_MS);
      Promise.resolve(violationRef.current?.(type))
        .then((r) => {
          if (r && typeof r.warnings === "number") {
            warnRef.current = r.warnings;
            setWarnings(r.warnings);
          }
        })
        .catch(() => {});
    };

    const tick = () => {
      const el = video.current;
      if (!el || !detector || el.readyState < 2) return;
      const quiet = Date.now() < quietUntil;

      // --- sound level ---
      let soundLevel = peakAudio;
      if (analyser) {
        analyser.getByteTimeDomainData(samples);
        const instantLevel = rms(samples);
        if (instantLevel > soundLevel) soundLevel = instantLevel;
      }
      peakAudio = 0;
      const isLoud = soundLevel > floor;
      setLoud(isLoud);

      // --- vision detections (people & cell phones) ---
      const track = stream?.getVideoTracks()[0];
      let rawDetections = [];
      if (track && track.readyState === "live") {
        try {
          rawDetections =
            detector.detectForVideo(el, performance.now()).detections || [];
        } catch {
          return;
        }
      }

      const personBoxes = [];
      let phoneFound = false;

      for (const d of rawDetections) {
        const cat = d.categories?.[0];
        if (!cat) continue;
        const score = cat.score;

        if (cat.categoryName === "cell phone" && score >= PHONE_MIN_SCORE) {
          phoneFound = true;
        } else if (cat.categoryName === "person" && score >= PERSON_MIN_SCORE) {
          personBoxes.push({ box: d.boundingBox, score });
        }
      }

      // Cluster overlapping person boxes so a single person (e.g. half portion / partial body) is not double-counted
      const distinctPeople = [];
      for (const p of personBoxes) {
        let isMerged = false;
        for (const existing of distinctPeople) {
          if (
            boxIoU(p.box, existing.box) > 0.2 ||
            isCenterInside(p.box, existing.box) ||
            isCenterInside(existing.box, p.box)
          ) {
            existing.score = Math.max(existing.score, p.score);
            isMerged = true;
            break;
          }
        }
        if (!isMerged) {
          distinctPeople.push({ ...p });
        }
      }

      // If at least one person box is detected, the candidate is in frame (even if half portion with score >= PERSON_MIN_SCORE)
      let n = 0;
      if (distinctPeople.length > 0) {
        n = 1;
        // Count additional people only if they meet the confident threshold to avoid false positives
        for (let i = 1; i < distinctPeople.length; i++) {
          if (distinctPeople[i].score >= SECOND_PERSON_MIN_SCORE) {
            n++;
          }
        }
      }

      setPeople(n);
      setHasPhone(phoneFound);

      const resetRuns = () => {
        noPersonRun = 0;
        multiplePeopleRun = 0;
        phoneRun = 0;
        loudRun = 0;
      };

      if (quiet) {
        resetRuns();
        return;
      }

      // Check phone run
      if (phoneFound) {
        phoneRun++;
      } else {
        phoneRun = Math.max(0, phoneRun - 1);
      }

      // Check sound run
      if (isLoud) {
        loudRun++;
      } else {
        loudRun = Math.max(0, loudRun - 1);
      }

      // Check people runs
      if (n === 0) {
        noPersonRun++;
        multiplePeopleRun = 0;
      } else if (n >= 2) {
        multiplePeopleRun++;
        noPersonRun = 0;
      } else {
        noPersonRun = 0;
        multiplePeopleRun = 0;
      }

      // Prioritize violation alerts
      if (phoneRun >= PHONE_SAMPLES) {
        resetRuns();
        quietUntil = Date.now() + COOLDOWN_MS;
        raise("PHONE_DETECTED");
      } else if (multiplePeopleRun >= MULTIPLE_PEOPLE_SAMPLES) {
        resetRuns();
        quietUntil = Date.now() + COOLDOWN_MS;
        raise("MULTIPLE_PEOPLE");
      } else if (noPersonRun >= NO_PERSON_SAMPLES) {
        resetRuns();
        quietUntil = Date.now() + COOLDOWN_MS;
        raise("NO_PERSON");
      } else if (loudRun >= LOUD_SAMPLES) {
        resetRuns();
        quietUntil = Date.now() + COOLDOWN_MS;
        raise("LOUD_VOICE");
      }
    };

    (async () => {
      try {
        stream = await navigator.mediaDevices.getUserMedia({
          video: { width: 640, height: 480, facingMode: "user" },
          audio: {
            echoCancellation: false,
            noiseSuppression: false,
            autoGainControl: false,
          },
        });
      } catch {
        try {
          stream = await navigator.mediaDevices.getUserMedia({
            video: { width: 640, height: 480, facingMode: "user" },
            audio: true,
          });
        } catch {
          try {
            stream = await navigator.mediaDevices.getUserMedia({
              video: { width: 640, height: 480, facingMode: "user" },
              audio: false,
            });
          } catch {
            if (!stopped) setStatus("denied");
            return;
          }
        }
      }
      if (stopped) {
        stream.getTracks().forEach((t) => t.stop());
        return;
      }

      if (stream.getAudioTracks().length > 0) {
        try {
          audioCtx = new (window.AudioContext || window.webkitAudioContext)();
          const src = audioCtx.createMediaStreamSource(
            new MediaStream(stream.getAudioTracks()),
          );
          analyser = audioCtx.createAnalyser();
          analyser.fftSize = 1024;
          src.connect(analyser); // measured only: never connected to the speakers, never recorded
          if (audioCtx.state === "suspended") await audioCtx.resume();
          setMicOn(true);

          audioTimer = setInterval(() => {
            if (!analyser || stopped) return;
            analyser.getByteTimeDomainData(samples);
            const level = rms(samples);
            if (level > peakAudio) peakAudio = level;

            if (calib.length < 10) {
              calib.push(level);
              if (calib.length === 10) {
                const sorted = [...calib].sort((a, b) => a - b);
                const base = sorted[Math.floor(sorted.length * 0.3)];
                floor = Math.min(
                  LOUD_CAP,
                  Math.max(LOUD_FLOOR, base + LOUD_MARGIN),
                );
              }
            } else {
              setLoud(level > floor);
            }
          }, AUDIO_POLL_MS);
        } catch {
          analyser = null;
        }
      }

      video.current.srcObject = stream;
      try {
        await video.current.play();
      } catch {
        /* muted video may autoplay */
      }
      try {
        detector = await createDetector();
      } catch {
        if (!stopped) setStatus("error");
        return;
      }
      if (stopped) {
        detector.close();
        return;
      }
      setStatus("ready");
      timer = setInterval(tick, SAMPLE_MS);
    })();

    return () => {
      stopped = true;
      clearInterval(timer);
      clearInterval(audioTimer);
      clearTimeout(alertTimer.current);
      stream?.getTracks().forEach((t) => t.stop());
      detector?.close();
      audioCtx?.close().catch(() => {});
    };
  }, [started, attempt]);

  const retry = () => {
    setStatus("loading");
    setAttempt((a) => a + 1);
  };

  const ok = status === "ready" && people === 1 && !hasPhone;
  const bad =
    (status === "ready" && people !== null && people !== 1) || hasPhone;
  const label =
    people === null
      ? "..."
      : people === 0
        ? "No one"
        : people === 1
          ? "1 person"
          : `${people} people`;

  // ---- before-you-start screen ----
  if (!started) {
    return (
      <div className="proctor-block">
        <div className="card">
          <ShieldCheck size={32} color="var(--accent)" />
          <h2 style={{ marginTop: 10 }}>Before you start</h2>
          <ul className="prep">
            <li>
              <Volume2 size={16} />
              <span>
                <strong>Find a quiet place.</strong> Any noise or talking will
                trigger a warning.
              </span>
            </li>
            <li>
              <Camera size={16} />
              <span>
                <strong>Stay in the frame.</strong> Keep your face in view.
                Another person in view will trigger a warning.
              </span>
            </li>
            <li>
              <Smartphone size={16} />
              <span>
                <strong>No mobile phones.</strong> Phones in camera view will
                trigger a cheating warning.
              </span>
            </li>
          </ul>
          <p className="muted small" style={{ marginBottom: 6 }}>
            Nothing is recorded, stored or uploaded, and no speech is
            recognised.
          </p>
          <p className="muted small" style={{ marginBottom: 16 }}>
            3 warnings end the assessment. Your timer is already running.
          </p>
          <button
            className="btn btn-primary"
            style={{ width: "100%" }}
            onClick={() => setStarted(true)}
          >
            I am ready, start
          </button>
        </div>
      </div>
    );
  }

  return (
    <>
      <div
        className={`proctor-cam ${compact ? "compact" : ""} ${ok && !loud && !hasPhone ? "ok" : bad || loud || hasPhone ? "bad" : ""}`}
      >
        <video ref={video} muted playsInline autoPlay />
        <div className="proctor-bar">
          <span
            style={{
              color: ok ? "var(--good)" : bad ? "var(--bad)" : "var(--muted)",
            }}
          >
            {label}
          </span>
          {hasPhone && (
            <span
              style={{
                color: "var(--bad)",
                display: "inline-flex",
                gap: 3,
                alignItems: "center",
                fontWeight: 700,
              }}
            >
              <Smartphone size={12} /> Phone
            </span>
          )}
          <span
            style={{
              color: loud ? "var(--bad)" : "var(--muted)",
              display: "inline-flex",
              gap: 4,
              alignItems: "center",
            }}
          >
            {micOn ? <Mic size={12} /> : null}
            {micOn ? (loud ? "Loud" : "Quiet") : "No mic"}
          </span>
          <span style={{ color: warnings > 0 ? "var(--bad)" : "var(--muted)" }}>
            {Math.min(warnings, MAX_WARNINGS)}/{MAX_WARNINGS}
          </span>
        </div>
      </div>

      {alert && (
        <div className="proctor-alert" role="alert">
          <div>
            {alert.message}
            <small>
              Warning {alert.n} of {MAX_WARNINGS}
              {alert.n >= MAX_WARNINGS ? " · ending the assessment" : ""}
            </small>
          </div>
        </div>
      )}

      {status !== "ready" && (
        <div className="proctor-block">
          <div className="card">
            {status === "loading" && (
              <>
                <div className="center" style={{ padding: 10 }}>
                  <Spinner />
                </div>
                <h2>Starting camera and microphone check</h2>
                <p className="muted small">
                  Allow camera and microphone access when your browser asks.
                  Nothing is recorded or uploaded.
                </p>
              </>
            )}
            {status === "denied" && (
              <>
                <CameraOff size={32} color="var(--bad)" />
                <h2 style={{ marginTop: 10 }}>Camera access is required</h2>
                <p className="muted small" style={{ marginBottom: 16 }}>
                  This assessment checks that one person is in front of the
                  camera. Allow camera access in your browser's address bar,
                  then try again. The timer keeps running.
                </p>
                <button className="btn btn-primary" onClick={retry}>
                  <Camera size={16} />
                  Try again
                </button>
              </>
            )}
            {status === "error" && (
              <>
                <CameraOff size={32} color="var(--bad)" />
                <h2 style={{ marginTop: 10 }}>Could not load the detector</h2>
                <p className="muted small" style={{ marginBottom: 16 }}>
                  The detection model is downloaded from the internet. Check
                  your connection and try again.
                </p>
                <button className="btn btn-primary" onClick={retry}>
                  Try again
                </button>
              </>
            )}
          </div>
        </div>
      )}
    </>
  );
}
