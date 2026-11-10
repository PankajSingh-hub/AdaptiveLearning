import { useCallback, useEffect, useRef, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import {
  ArrowRight,
  CheckCircle2,
  Clock,
  Sparkles,
  XCircle,
} from "lucide-react";
import {
  finishInterview,
  getInterview,
  submitAnswer,
  warnInterview,
} from "../services/candidateService";
import { errMsg } from "../services/api";
import { useToast } from "../context/ToastContext";
import {
  DifficultyBadge,
  Badge,
  ErrorMessage,
  PageLoader,
  Progress,
} from "../components/ui";
import ConfirmDialog from "../components/ConfirmDialog";
import { formatClock } from "../utils/format";
import { APP_NAME } from "../utils/constants";
import Proctor from "../components/Proctor";
export default function InterviewSession() {
  const { id } = useParams();
  const navigate = useNavigate();
  const toast = useToast();

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [ready, setReady] = useState(false);
  const [question, setQuestion] = useState(null);
  const [selected, setSelected] = useState(null);
  const [result, setResult] = useState(null);
  const [busy, setBusy] = useState(false);
  const [left, setLeft] = useState(0);
  const [score, setScore] = useState(0);
  const [warnings, setWarnings] = useState(0);
  const [confirmEnd, setConfirmEnd] = useState(false);
  const deadline = useRef(0);
  const finishing = useRef(false);

  // Load (or resume) the session. The server is the source of truth for the remaining time.
  useEffect(() => {
    getInterview(id)
      .then((s) => {
        if (s.status !== "IN_PROGRESS")
          return navigate(`/report/${s.reportId}`, { replace: true });
        if (!s.question)
          return setError("No active question found for this interview.");
        setQuestion(s.question);
        setScore(s.totalScore);
        deadline.current = Date.now() + s.remainingSec * 1000;
        setLeft(s.remainingSec);
        setReady(true);
        setWarnings(s.warnings ?? 0);
        setReady(true);
      })
      .catch((e) => setError(errMsg(e)))
      .finally(() => setLoading(false));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id]);

  const endNow = useCallback(async () => {
    if (finishing.current) return;
    finishing.current = true;
    setBusy(true);
    try {
      const r = await finishInterview(id);
      navigate(`/report/${r.reportId}`, { replace: true });
    } catch (e) {
      finishing.current = false;
      setBusy(false);
      toast.error(errMsg(e));
    }
  }, [id, navigate, toast]);
  const onViolation = useCallback(
    async (type) => {
      const r = await warnInterview(id, type);
      if (r.terminated) {
        toast.error("Interview ended after 3 proctoring warnings");
        if (r.reportId) navigate(`/report/${r.reportId}`, { replace: true });
        else endNow();
      }
      return r;
    },
    [id, navigate, toast, endNow],
  );

  // Countdown. At zero the interview is submitted; the backend checks expiry again anyway.
  useEffect(() => {
    if (!ready) return;
    const timer = setInterval(() => {
      const s = Math.max(0, Math.ceil((deadline.current - Date.now()) / 1000));
      setLeft(s);
      if (s === 0) {
        clearInterval(timer);
        toast.error("Time's up — submitting your interview");
        endNow();
      }
    }, 1000);
    return () => clearInterval(timer);
  }, [ready, endNow, toast]);

  const submit = async () => {
    if (!selected || busy || result) return;
    setBusy(true);
    try {
      const r = await submitAnswer(id, question.sessionQuestionId, selected);
      setScore(r.totalScore);
      if (r.expired) {
        toast.error("Time expired");
        return r.reportId
          ? navigate(`/report/${r.reportId}`, { replace: true })
          : endNow();
      }
      setResult(r);
    } catch (e) {
      toast.error(errMsg(e));
    } finally {
      setBusy(false);
    }
  };

  const next = () => {
    if (result.completed)
      return result.reportId
        ? navigate(`/report/${result.reportId}`, { replace: true })
        : endNow();
    setQuestion(result.nextQuestion);
    setResult(null);
    setSelected(null);
  };

  const optionClass = (o) => {
    let c = "option";
    if (!result) return selected === o.id ? `${c} sel` : c;
    const isRight = result.correctOptionId
      ? o.id === result.correctOptionId
      : o.id === selected && result.correct;
    if (isRight) return `${c} right`;
    return o.id === selected ? `${c} wrong` : c;
  };

  if (loading) return <PageLoader text="Loading interview..." />;
  if (error)
    return (
      <div className="interview">
        <ErrorMessage message={error} />
      </div>
    );

  const answered = result ? question.number : question.number - 1;

  return (
    <>
      <Proctor initialWarnings={warnings} onViolation={onViolation} />

      <div className="iv-top">
        <div className="brand" style={{ padding: 0 }}>
          <Sparkles size={20} color="#6d8dff" />
          <span className="grad">{APP_NAME}</span>
        </div>
        <div className={`timer mono ${left <= 60 ? "low" : ""}`}>
          <Clock size={18} />
          {formatClock(left)}
        </div>
        <button
          className="btn"
          disabled={busy}
          onClick={() => setConfirmEnd(true)}
        >
          End
        </button>
      </div>

      <div className="interview">
        <div
          className="small muted"
          style={{
            display: "flex",
            justifyContent: "space-between",
            marginBottom: 8,
          }}
        >
          <span>
            Question {question.number} of {question.total}
          </span>
          <span>Score {score}</span>
        </div>
        <Progress value={(answered / question.total) * 100} />

        <div
          key={question.sessionQuestionId}
          className={result ? "locked" : ""}
        >
          <div className="q-meta">
            <DifficultyBadge level={question.difficulty} />
            <Badge tone="accent">{question.topic}</Badge>
          </div>
          <div className="q-text">{question.text}</div>
          {question.options.map((o, i) => (
            <div
              key={o.id}
              className={optionClass(o)}
              style={{ animationDelay: `${i * 60}ms` }}
              onClick={() => !result && setSelected(o.id)}
            >
              <span className="dot" />
              <span>{o.text}</span>
            </div>
          ))}
        </div>

        {result && (
          <div className={`feedback ${result.correct ? "ok" : "no"}`}>
            <div
              style={{
                display: "flex",
                gap: 8,
                alignItems: "center",
                fontWeight: 600,
              }}
            >
              {result.correct ? (
                <CheckCircle2 size={20} color="#34d399" />
              ) : (
                <XCircle size={20} color="#f87171" />
              )}
              {result.correct
                ? `Correct · +${result.pointsAwarded} pts`
                : "Incorrect"}
            </div>
            <div className="small muted" style={{ marginTop: 6 }}>
              {result.difficultyBefore !== result.difficultyAfter
                ? `Difficulty adjusted: ${result.difficultyBefore} → ${result.difficultyAfter}`
                : `Difficulty stays ${result.difficultyAfter}`}
            </div>
            {result.explanation && (
              <p className="small" style={{ marginTop: 10 }}>
                {result.explanation}
              </p>
            )}
          </div>
        )}

        <div className="actions">
          {!result ? (
            <button
              className="btn btn-primary"
              disabled={!selected || busy}
              onClick={submit}
            >
              {busy ? "Submitting answer..." : "Submit answer"}
            </button>
          ) : (
            <button className="btn btn-primary" disabled={busy} onClick={next}>
              {result.completed ? "View report" : "Next question"}{" "}
              <ArrowRight size={16} />
            </button>
          )}
        </div>
      </div>

      {confirmEnd && (
        <ConfirmDialog
          title="End interview?"
          confirmText="End and see report"
          busy={busy}
          message="Unanswered questions will not count. You cannot resume after ending."
          onCancel={() => setConfirmEnd(false)}
          onConfirm={endNow}
        />
      )}
    </>
  );
}
