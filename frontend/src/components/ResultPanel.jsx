import { CheckCircle2, Lightbulb, XCircle } from "lucide-react";
import { Badge } from "./ui";

const TONE = {
  ACCEPTED: "good",
  WRONG_ANSWER: "bad",
  RUNTIME_ERROR: "bad",
  TIME_LIMIT: "warn",
  COMPILE_ERROR: "warn",
};
const LABEL = {
  ACCEPTED: "Accepted",
  WRONG_ANSWER: "Wrong answer",
  RUNTIME_ERROR: "Runtime error",
  TIME_LIMIT: "Time limit exceeded",
  COMPILE_ERROR: "Compile error",
};

function CaseRow({ c }) {
  return (
    <div className="case">
      <div className="case-head">
        {c.passed ? (
          <CheckCircle2 size={16} color="#86d1a0" />
        ) : (
          <XCircle size={16} color="#e5846a" />
        )}
        <strong>{c.visible ? "Example" : "Hidden test"}</strong>
        {!c.visible && <span className="muted small">{c.tag}</span>}
        <span className="muted small" style={{ marginLeft: "auto" }}>
          {c.status} · {c.timeMs} ms
        </span>
      </div>
      {c.visible && (
        <>
          <div className="kv">Input: {c.input}</div>
          <div className="kv">Expected: {c.expected}</div>
          {!c.passed && (
            <div className="kv">Yours: {c.got || "(no output)"}</div>
          )}
        </>
      )}
    </div>
  );
}

export default function ResultPanel({ r }) {
  const runOk = r.mode === "run" && r.verdict === "ACCEPTED";
  return (
    <div className="card result">
      <div
        style={{
          display: "flex",
          gap: 10,
          alignItems: "center",
          flexWrap: "wrap",
        }}
      >
        <Badge tone={TONE[r.verdict]}>
          {runOk ? "Examples passed" : LABEL[r.verdict]}
        </Badge>
        {r.verdict !== "COMPILE_ERROR" && (
          <span className="muted small">
            {r.passed}/{r.total} tests passed ·{" "}
            {r.mode === "run" ? "examples only" : "full judge"}
          </span>
        )}
      </div>

      {r.compileOutput && <pre className="out">{r.compileOutput}</pre>}
      {r.detail && r.verdict === "RUNTIME_ERROR" && (
        <pre className="out">{r.detail}</pre>
      )}
      {r.analysis?.length > 0 && (
        <ul className="notes">
          {r.analysis.map((n, i) => (
            <li key={i}>{n}</li>
          ))}
        </ul>
      )}
      {runOk && (
        <p className="muted small" style={{ marginTop: 10 }}>
          Run only checks the examples and uses no hints. Press Submit to run
          every test.
        </p>
      )}

      {r.cases?.length > 0 && (
        <div style={{ marginTop: 12 }}>
          {r.cases.map((c, i) => (
            <CaseRow key={i} c={c} />
          ))}
        </div>
      )}

      {r.hint && (
        <div className="callout hintbox">
          <div
            style={{
              display: "flex",
              gap: 8,
              alignItems: "center",
              fontWeight: 600,
            }}
          >
            <Lightbulb size={18} color="#e8c46e" /> Hint {r.hintNumber} of 5
          </div>
          <p style={{ marginTop: 6 }}>{r.hint}</p>
        </div>
      )}

      {r.solved && r.feedback && (
        <div className="callout okbox">
          <strong>Solved · +{r.points} pts</strong>
          <div className="complexity">
            <Badge tone="accent">Optimal time {r.feedback.timeOptimal}</Badge>
            <Badge tone="accent">Optimal space {r.feedback.spaceOptimal}</Badge>
            <Badge>Slowest test {r.feedback.runtimeMs} ms</Badge>
          </div>
          <p className="small">
            <strong>Reference approach:</strong> {r.feedback.approach}
          </p>
          <ul className="notes">
            {r.feedback.notes.map((n, i) => (
              <li key={i}>{n}</li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}
