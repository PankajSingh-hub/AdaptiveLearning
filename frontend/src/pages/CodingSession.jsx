import { useCallback, useEffect, useRef, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import {
  ArrowRight,
  Clock,
  Lightbulb,
  Lock,
  Play,
  RotateCcw,
  Send,
  SkipForward,
  Sparkles,
} from "lucide-react";
import {
  finishCoding,
  getCodingProblem,
  getCodingSession,
  runCode,
  skipCoding,
  submitCode,
  warnCoding,
} from "../services/candidateService";
import { errMsg } from "../services/api";
import { useToast } from "../context/ToastContext";
import {
  Badge,
  DifficultyBadge,
  ErrorMessage,
  PageLoader,
} from "../components/ui";
import ConfirmDialog from "../components/ConfirmDialog";
import CodeEditor from "../components/CodeEditor";
import ResultPanel from "../components/ResultPanel";
import Proctor from "../components/Proctor";
import { formatClock } from "../utils/format";
import { APP_NAME } from "../utils/constants";
import "../styles/coding.css";

const LANG_LABEL = { cpp: "C++", java: "Java", python: "Python" };
const PREF_KEY = "aap_lang";
const draftKey = (sid, order, lang) => `aap_code_${sid}_${order}_${lang}`;
const read = (k) => {
  try {
    return localStorage.getItem(k);
  } catch {
    return null;
  }
};
const write = (k, v) => {
  try {
    localStorage.setItem(k, v);
  } catch {
    /* storage unavailable */
  }
};

export default function CodingSession() {
  const { id } = useParams();
  const navigate = useNavigate();
  const toast = useToast();

  const [session, setSession] = useState(null);
  const [order, setOrder] = useState(0);
  const [problem, setProblem] = useState(null);
  const [lang, setLang] = useState(() => read(PREF_KEY) || "python");
  const [code, setCode] = useState("");
  const [result, setResult] = useState(null);
  const [busy, setBusy] = useState(null); // null | 'run' | 'submit' | 'skip'
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [left, setLeft] = useState(0);
  const [advance, setAdvance] = useState(null); // { next, left } after the 5th failed submit
  const [confirmEnd, setConfirmEnd] = useState(false);
  const [confirmSkip, setConfirmSkip] = useState(false);
  const deadline = useRef(0);
  const finishing = useRef(false);

  const toResult = useCallback(
    () => navigate(`/coding/${id}/result`, { replace: true }),
    [id, navigate],
  );
  const goNext = useCallback(
    (next) => (next ? setOrder(next) : toResult()),
    [toResult],
  );

  // Load the session. The server is the source of truth for remaining time.
  useEffect(() => {
    getCodingSession(id)
      .then((s) => {
        if (s.status !== "IN_PROGRESS" || !s.current) return toResult();
        setSession(s);
        deadline.current = Date.now() + s.remainingSec * 1000;
        setLeft(s.remainingSec);
        setOrder(s.current);
      })
      .catch((e) => setError(errMsg(e)))
      .finally(() => setLoading(false));
  }, [id, toResult]);

  // Load the selected problem.
  useEffect(() => {
    if (!order) return;
    let cancelled = false;
    setProblem(null);
    setResult(null);
    setAdvance(null);
    getCodingProblem(id, order)
      .then((p) => {
        if (cancelled) return;
        setProblem(p);
        const langs = Object.keys(p.starter);
        setLang((l) => (langs.includes(l) ? l : langs[0]));
      })
      .catch((e) => !cancelled && toast.error(errMsg(e)));
    return () => {
      cancelled = true;
    };
  }, [id, order, toast]);

  // Load the draft (or the starter code) whenever the problem or language changes.
  useEffect(() => {
    if (!problem || !lang) return;
    setCode(
      read(draftKey(id, problem.order, lang)) ?? problem.starter[lang] ?? "",
    );
  }, [problem, lang, id]);

  const endNow = useCallback(async () => {
    if (finishing.current) return;
    finishing.current = true;
    try {
      await finishCoding(id);
    } catch (e) {
      finishing.current = false;
      toast.error(errMsg(e));
      return;
    }
    toResult();
  }, [id, toResult, toast]);

  // Camera check reports each warning to the backend; the third one ends the session.
  const onViolation = useCallback(
    async (type) => {
      const r = await warnCoding(id, type);
      if (r.terminated) {
        toast.error("Coding session ended after 3 proctoring warnings");
        toResult();
      }
      return r;
    },
    [id, toResult, toast],
  );

  // Countdown. At zero the session is closed; the backend enforces expiry on its own as well.
  useEffect(() => {
    if (!session) return;
    const timer = setInterval(() => {
      const s = Math.max(0, Math.ceil((deadline.current - Date.now()) / 1000));
      setLeft(s);
      if (s === 0) {
        clearInterval(timer);
        toast.error("Time's up");
        endNow();
      }
    }, 1000);
    return () => clearInterval(timer);
  }, [session, endNow, toast]);

  // Auto-advance after the 5th failed submit.
  useEffect(() => {
    if (!advance) return;
    if (advance.left <= 0) {
      goNext(advance.next);
      return;
    }
    const t = setTimeout(
      () => setAdvance((a) => a && { ...a, left: a.left - 1 }),
      1000,
    );
    return () => clearTimeout(t);
  }, [advance, goNext]);

  const onCode = (v) => {
    setCode(v);
    write(draftKey(id, order, lang), v);
  };
  const onLang = (l) => {
    setLang(l);
    write(PREF_KEY, l);
    setResult(null);
  };
  const resetCode = () => onCode(problem.starter[lang]);

  const skip = async () => {
    if (busy) return;
    setBusy("skip");
    try {
      const r = await skipCoding(id, order);
      setConfirmSkip(false);
      getCodingSession(id)
        .then(setSession)
        .catch(() => {});
      goNext(r.nextOrder);
    } catch (e) {
      setConfirmSkip(false);
      toast.error(errMsg(e));
    } finally {
      setBusy(null);
    }
  };

  const judge = async (kind) => {
    if (busy || !problem || problem.status !== "OPEN") return;
    if (!code.trim()) return toast.error("Write some code first");
    setBusy(kind);
    setResult(null);
    try {
      const r = await (kind === "run" ? runCode : submitCode)(
        id,
        order,
        lang,
        code,
      );
      setResult(r);
      if (kind === "submit") {
        setProblem((p) => ({
          ...p,
          failures: r.failures,
          status: r.solved ? "SOLVED" : r.exhausted ? "FAILED" : p.status,
          hints: r.hint ? [...p.hints, r.hint] : p.hints,
        }));
        getCodingSession(id)
          .then(setSession)
          .catch(() => {});
        if (r.exhausted) setAdvance({ next: r.nextOrder, left: 10 });
      }
    } catch (e) {
      toast.error(errMsg(e));
    } finally {
      setBusy(null);
    }
  };

  if (loading) return <PageLoader text="Loading coding session..." />;
  if (error)
    return (
      <div className="interview">
        <ErrorMessage message={error} />
      </div>
    );
  if (!session) return null;

  const open = problem?.status === "OPEN";

  return (
    <>
      <Proctor
        initialWarnings={session.warnings ?? 0}
        onViolation={onViolation}
      />

      <div className="iv-top cs-top">
        <div className="brand" style={{ padding: 0 }}>
          <Sparkles size={20} color="#e8b86a" />
          <span className="grad">{APP_NAME}</span>
        </div>
        <div className="pchips">
          {session.problems.map((p) => (
            <button
              key={p.order}
              className={`pchip ${p.status} ${p.order === order ? "cur" : ""}`}
              disabled={!!busy}
              title={`${p.order}. ${p.title}`}
              onClick={() => setOrder(p.order)}
            >
              {p.order}
            </button>
          ))}
        </div>
        <div className={`timer mono ${left <= 300 ? "low" : ""}`}>
          <Clock size={18} />
          {formatClock(left)}
        </div>
        <button
          className="btn"
          disabled={!!busy}
          onClick={() => setConfirmEnd(true)}
        >
          End
        </button>
      </div>

      {!session.sandboxed && (
        <div className="demo-banner">
          Local demo mode: your code runs on the host machine without a sandbox.
          Never expose this deployment publicly.
        </div>
      )}

      <div className={`cs-body ${!session.sandboxed ? "with-banner" : ""}`}>
        {/* LEFT: problem */}
        <div className="cs-pane">
          {!problem ? (
            <PageLoader text="Loading problem..." />
          ) : (
            <>
              <div className="card">
                <div
                  style={{
                    display: "flex",
                    gap: 10,
                    alignItems: "center",
                    flexWrap: "wrap",
                    marginBottom: 12,
                  }}
                >
                  <h2 style={{ margin: 0 }}>
                    {problem.order}. {problem.title}
                  </h2>
                  <DifficultyBadge level={problem.difficulty} />
                  {!open && (
                    <Badge tone={problem.status === "SOLVED" ? "good" : "bad"}>
                      {problem.status}
                    </Badge>
                  )}
                </div>
                <div className="desc">{problem.description}</div>
                {problem.examples.map((ex, i) => (
                  <div
                    key={i}
                    className="example"
                  >{`Example ${i + 1}\nInput:  ${ex.input}\nOutput: ${ex.output}`}</div>
                ))}
              </div>

              <div className="card">
                <h2 style={{ display: "flex", alignItems: "center", gap: 8 }}>
                  <Lightbulb size={18} color="#e8c46e" /> Hints
                  <span className="muted small" style={{ marginLeft: "auto" }}>
                    {problem.failures} of {problem.maxFailures} used
                  </span>
                </h2>
                {Array.from({ length: problem.maxFailures }, (_, i) =>
                  problem.hints[i] ? (
                    <div key={i} className="hint on">
                      <Lightbulb size={16} color="#e8c46e" />
                      <span>{problem.hints[i]}</span>
                    </div>
                  ) : (
                    <div key={i} className="hint off">
                      <Lock size={16} />
                      <span>
                        Hint {i + 1} unlocks after{" "}
                        {i === 0
                          ? "your first failed submission"
                          : `failed submission ${i + 1}`}
                      </span>
                    </div>
                  ),
                )}
                <p className="muted small" style={{ marginTop: 8 }}>
                  After the 5th failed submission the problem closes and you
                  move on. You can also skip a problem. Compile errors and Run
                  do not use a hint.
                </p>
              </div>
            </>
          )}
        </div>

        {/* RIGHT: editor and results */}
        <div className="cs-pane">
          {problem && (
            <div className="card">
              <div className="code-bar">
                <select
                  className="input"
                  value={lang}
                  onChange={(e) => onLang(e.target.value)}
                  disabled={!!busy}
                >
                  {Object.keys(problem.starter).map((l) => (
                    <option key={l} value={l}>
                      {LANG_LABEL[l] || l}
                    </option>
                  ))}
                </select>
                <button
                  className="btn"
                  onClick={resetCode}
                  disabled={!open || !!busy}
                >
                  <RotateCcw size={15} />
                  Reset
                </button>
              </div>

              <CodeEditor
                value={code}
                onChange={onCode}
                disabled={!open}
                language={lang}
              />

              {open ? (
                <div className="run-actions">
                  <button
                    className="btn"
                    disabled={!!busy}
                    onClick={() => judge("run")}
                  >
                    <Play size={16} />
                    {busy === "run" ? "Running..." : "Run"}
                  </button>
                  <button
                    className="btn btn-primary"
                    disabled={!!busy}
                    onClick={() => judge("submit")}
                  >
                    <Send size={16} />
                    {busy === "submit" ? "Judging..." : "Submit"}
                  </button>
                  <button
                    className="btn"
                    disabled={!!busy}
                    onClick={() => setConfirmSkip(true)}
                  >
                    <SkipForward size={16} />
                    Skip
                  </button>
                  <span className="muted small">
                    Run checks the examples for free. A failed Submit unlocks
                    the next hint.
                  </span>
                </div>
              ) : (
                <p className="muted small" style={{ margin: "12px 0" }}>
                  This problem is finished. Pick an open problem from the
                  numbers above.
                </p>
              )}

              {result && <ResultPanel r={result} />}

              {result?.solved && (
                <div className="run-actions">
                  <button
                    className="btn btn-primary"
                    onClick={() => goNext(result.nextOrder)}
                  >
                    {result.sessionDone ? "See results" : "Next problem"}{" "}
                    <ArrowRight size={16} />
                  </button>
                </div>
              )}
              {advance && (
                <div className="callout advance">
                  <span>
                    Out of hints.{" "}
                    {advance.next
                      ? `Moving to the next problem in ${advance.left}s...`
                      : `Finishing in ${advance.left}s...`}
                  </span>
                  <button className="btn" onClick={() => goNext(advance.next)}>
                    {advance.next ? "Next now" : "See results"}
                  </button>
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      {confirmEnd && (
        <ConfirmDialog
          title="End coding session?"
          confirmText="End and see results"
          message="Open problems will be marked skipped. You cannot resume after ending."
          onCancel={() => setConfirmEnd(false)}
          onConfirm={endNow}
        />
      )}

      {confirmSkip && (
        <ConfirmDialog
          title="Skip this problem?"
          confirmText="Skip"
          busy={busy === "skip"}
          message="You will get no points for it and cannot come back to it."
          onCancel={() => setConfirmSkip(false)}
          onConfirm={skip}
        />
      )}
    </>
  );
}
