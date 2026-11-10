import { useEffect, useRef, useState } from "react";
import { Camera, CameraOff } from "lucide-react";
import { FilesetResolver, ObjectDetector } from "@mediapipe/tasks-vision";
import { Spinner } from "./ui";
import "../styles/proctor.css";

const WASM =
  "https://cdn.jsdelivr.net/npm/@mediapipe/tasks-vision@0.10.14/wasm";
const MODEL =
  "https://storage.googleapis.com/mediapipe-models/object_detector/efficientdet_lite0/float16/1/efficientdet_lite0.tflite";

// Tunable rules
const SAMPLE_MS = 500; // how often the frame is checked
const BAD_SAMPLES = 4; // 4 x 500 ms = a problem must last 2 seconds in a row
const COOLDOWN_MS = 5000; // grace period after a warning
const ALERT_MS = 2000; // how long the red label stays on screen
export const MAX_WARNINGS = 3;

const MESSAGES = {
  NO_PERSON: "No person detected in the camera frame",
  MULTIPLE_PEOPLE: "More than one person detected",
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

// Counts people in the webcam frame (body detection, no eye or gaze tracking).
// onViolation(type) must report the warning to the backend and resolve to { warnings }.
export default function Proctor({ initialWarnings = 0, onViolation }) {
  const video = useRef(null);
  const [status, setStatus] = useState("loading"); // loading | ready | denied | error
  const [attempt, setAttempt] = useState(0);
  const [people, setPeople] = useState(null);
  const [warnings, setWarnings] = useState(initialWarnings);
  const [alert, setAlert] = useState(null);

  const warnRef = useRef(initialWarnings);
  const violationRef = useRef(onViolation);
  violationRef.current = onViolation;
  const alertTimer = useRef(null);

  useEffect(() => {
    let stopped = false;
    let stream = null;
    let detector = null;
    let timer = null;
    let bad = 0;
    let quietUntil = 0;
    setStatus("loading");

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
      if (Date.now() < quietUntil || n === 1) {
        bad = 0;
        return;
      }
      bad += 1;
      if (bad >= BAD_SAMPLES) {
        bad = 0;
        quietUntil = Date.now() + COOLDOWN_MS;
        raise(n === 0 ? "NO_PERSON" : "MULTIPLE_PEOPLE");
      }
    };

    (async () => {
      try {
        stream = await navigator.mediaDevices.getUserMedia({
          video: { width: 640, height: 480, facingMode: "user" },
          audio: false,
        });
      } catch {
        if (!stopped) setStatus("denied");
        return;
      }
      if (stopped) {
        stream.getTracks().forEach((t) => t.stop());
        return;
      }
      video.current.srcObject = stream;
      try {
        await video.current.play();
      } catch {
        /* autoplay is allowed for muted video */
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
    };
  }, [attempt]);

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

  return (
    <>
      <div className={`proctor-cam ${ok ? "ok" : bad ? "bad" : ""}`}>
        <video ref={video} muted playsInline autoPlay />
        <div className="proctor-bar">
          <span
            style={{
              color: ok ? "var(--good)" : bad ? "var(--bad)" : "var(--muted)",
            }}
          >
            {label}
          </span>
          <span style={{ color: warnings > 0 ? "var(--bad)" : "var(--muted)" }}>
            Warnings {Math.min(warnings, MAX_WARNINGS)}/{MAX_WARNINGS}
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
                <h2>Starting camera check</h2>
                <p className="muted small">
                  Allow camera access when your browser asks. Video stays on
                  your device and is never uploaded or recorded.
                </p>
              </>
            )}
            {status === "denied" && (
              <>
                <CameraOff size={32} color="#f87171" />
                <h2 style={{ marginTop: 10 }}>Camera access is required</h2>
                <p className="muted small" style={{ marginBottom: 16 }}>
                  This assessment checks that one person is in front of the
                  camera. Allow camera access in your browser's address bar,
                  then try again. The timer keeps running.
                </p>
                <button
                  className="btn btn-primary"
                  onClick={() => setAttempt((a) => a + 1)}
                >
                  <Camera size={16} />
                  Try again
                </button>
              </>
            )}
            {status === "error" && (
              <>
                <CameraOff size={32} color="#f87171" />
                <h2 style={{ marginTop: 10 }}>Could not load the detector</h2>
                <p className="muted small" style={{ marginBottom: 16 }}>
                  The detection model is downloaded from the internet. Check
                  your connection and try again.
                </p>
                <button
                  className="btn btn-primary"
                  onClick={() => setAttempt((a) => a + 1)}
                >
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
