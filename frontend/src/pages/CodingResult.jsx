import { Link, useParams } from "react-router-dom";
import { ArrowLeft } from "lucide-react";
import useFetch from "../hooks/useFetch";
import { getCodingSession } from "../services/candidateService";
import {
  Badge,
  DifficultyBadge,
  ErrorMessage,
  PageLoader,
  Stat,
} from "../components/ui";

const TONE = { SOLVED: "good", FAILED: "bad", SKIPPED: "", OPEN: "warn" };
const LANG_LABEL = { cpp: "C++", java: "Java", python: "Python" };

export default function CodingResult() {
  const { id } = useParams();
  const {
    data: s,
    loading,
    error,
    reload,
  } = useFetch(() => getCodingSession(id), [id]);
  if (loading) return <PageLoader text="Loading results..." />;
  if (error) return <ErrorMessage message={error} onRetry={reload} />;

  const solved = s.problems.filter((p) => p.status === "SOLVED").length;
  const hintsUsed = s.problems.reduce((sum, p) => sum + p.failures, 0);

  return (
    <>
      <Link
        to="/assessments"
        className="muted small"
        style={{ display: "inline-flex", gap: 6, alignItems: "center" }}
      >
        <ArrowLeft size={14} /> Back to assessments
      </Link>
      <div className="page-head">
        <div>
          <h1>{s.assessmentTitle}</h1>
          <p className="muted">Coding results</p>
        </div>
        {s.status === "IN_PROGRESS" ? (
          <Link to={`/coding/${id}`} className="btn btn-primary">
            Continue session
          </Link>
        ) : (
          <Badge tone={s.status === "COMPLETED" ? "good" : "warn"}>
            {s.status}
          </Badge>
        )}
      </div>

      <div className="grid stats">
        <div className="card">
          <div className="muted small">Score</div>
          <div className="big-score grad">
            {s.score}
            <span className="muted" style={{ fontSize: "1.2rem" }}>
              /{s.maxScore}
            </span>
          </div>
        </div>
        <Stat label="Solved" value={solved} />
        <Stat label="Problems" value={s.problems.length} />
        <Stat label="Hints used" value={hintsUsed} />
      </div>

      <div className="card">
        <h2>Problems</h2>
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>#</th>
                <th>Problem</th>
                <th>Level</th>
                <th>Result</th>
                <th>Hints</th>
                <th>Points</th>
                <th>Language</th>
                <th>Slowest test</th>
                <th>Optimal time</th>
              </tr>
            </thead>
            <tbody>
              {s.problems.map((p) => (
                <tr key={p.order}>
                  <td>{p.order}</td>
                  <td>{p.title}</td>
                  <td>
                    <DifficultyBadge level={p.difficulty} />
                  </td>
                  <td>
                    <Badge tone={TONE[p.status]}>{p.status}</Badge>
                  </td>
                  <td>{p.failures}/5</td>
                  <td>{p.points}</td>
                  <td>{LANG_LABEL[p.language] || "—"}</td>
                  <td>{p.status === "SOLVED" ? `${p.runtimeMs} ms` : "—"}</td>
                  <td className="mono">{p.timeOptimal}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        <p className="muted small" style={{ marginTop: 12 }}>
          Points: 10 / 20 / 30 for Easy / Medium / Hard, minus 10% for each
          failed submission. "Optimal time" is the reference complexity for the
          problem, not a measurement of your code.
        </p>
      </div>
    </>
  );
}
