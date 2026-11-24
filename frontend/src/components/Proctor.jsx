import { useEffect, useRef, useState } from "react";
import { Camera, CameraOff, Mic, ShieldCheck, Volume2 } from "lucide-react";
import { FilesetResolver, ObjectDetector } from "@mediapipe/tasks-vision";
import { Spinner } from "./ui";
import "../styles/proctor.css";

const WASM =
  "https://cdn.jsdelivr.net/npm/@mediapipe/tasks-vision@0.10.14/wasm";
const MODEL =
  "https://storage.googleapis.com/mediapipe-models/object_detector/efficientdet_lite0/float16/1/efficientdet_lite0.tflite";

// Tunable rules
const SAMPLE_MS = 500; // how often the frame and the sound level are checked
const BAD_SAMPLES = 4; // 4 x 500 ms = a camera problem must last 2 seconds in a row
const LOUD_SAMPLES = 5; // 5 x 500 ms = loud sound must last 2.5 seconds in a row
const CALIBRATE_SAMPLES = 2; // first second: measure the room's background level
const LOUD_MARGIN = 0.07; // volume (RMS, 0..1) above the room's background that counts as loud
const LOUD_FLOOR = 0.09; // never treat anything below this absolute volume as loud
const COOLDOWN_MS = 5000; // grace period after any warning
const ALERT_MS = 2000; // how long the red label stays on screen
const MAX_WARNINGS = 3;

const MESSAGES = {
  NO_PERSON: "No person detected in the camera frame",
  MULTIPLE_PEOPLE: "More than one person detected",
  LOUD_VOICE: "Too much noise or talking detected",
};

async function createDetector() {
  const fileset = await FilesetResolver.forVisionTasks(WASM);
  const options = (delegate) => ({
    baseOptions: { modelAssetPath: MODEL, delegate },
    runningMode: "VIDEO",
    scoreThreshold: 0.5,
    maxResults: 5,
    categoryAllowlist: ["person"],
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
    let bad = 0;
    let loudRun = 0;
    let calib = [];
    let floor = LOUD_FLOOR;
    let quietUntil = 0;
    const samples = new Uint8Array(1024);

    const raise = (type) => {
      warnRef.current += 1;
      setWarnings(warnRef.current);
      setAlert({
        message: MESSAGES[type],
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
      if (analyser) {
        analyser.getByteTimeDomainData(samples);
        const level = rms(samples);
        if (calib.length < CALIBRATE_SAMPLES) {
          calib.push(level);
          if (calib.length === CALIBRATE_SAMPLES) {
            const base = calib.reduce((a, b) => a + b, 0) / calib.length;
            floor = Math.max(LOUD_FLOOR, base + LOUD_MARGIN);
          }
        } else {
          const isLoud = level > floor;
          setLoud(isLoud);
          loudRun = isLoud && !quiet ? loudRun + 1 : 0;
        }
      }

      // --- people in frame ---
      const track = stream?.getVideoTracks()[0];
      let n = 0;
      if (track && track.readyState === "live") {
        try {
          n = detector.detectForVideo(el, performance.now()).detections.length;
        } catch {
          return;
        }
      }
      setPeople(n);

      if (quiet) {
        bad = 0;
        loudRun = 0;
        return;
      }
      bad = n === 1 ? 0 : bad + 1;
      if (bad >= BAD_SAMPLES) {
        bad = 0;
        loudRun = 0;
        quietUntil = Date.now() + COOLDOWN_MS;
        raise(n === 0 ? "NO_PERSON" : "MULTIPLE_PEOPLE");
      } else if (loudRun >= LOUD_SAMPLES) {
        bad = 0;
        loudRun = 0;
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
        // Camera is required. Retry without audio so a missing microphone is not fatal.
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

  const ok = status === "ready" && people === 1;
  const bad = status === "ready" && people !== null && people !== 1;
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
                <strong>Find a quiet place.</strong> Loud noise or talking can
                trigger a warning.
              </span>
            </li>
            <li>
              <Camera size={16} />
              <span>
                <strong>Stay in the frame.</strong> Exactly one person, your
                face and shoulders visible, in good light.
              </span>
            </li>
            <li>
              <Mic size={16} />
              <span>
                <strong>Camera and microphone</strong> are used only to count
                people and measure volume.
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
        className={`proctor-cam ${compact ? "compact" : ""} ${ok && !loud ? "ok" : bad || loud ? "bad" : ""}`}
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
