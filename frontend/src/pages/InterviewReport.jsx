import { Link, useParams } from "react-router-dom";
import { ArrowLeft } from "lucide-react";
import useFetch from "../hooks/useFetch";
import { getReport } from "../services/candidateService";
import { useAuth } from "../hooks/useAuth";
import { Badge, ErrorMessage, PageLoader, Stat } from "../components/ui";
import { Donut, HBars, Steps } from "../components/Charts";
import { formatDate, formatLevel } from "../utils/format";
import { LEVEL_TONE } from "../utils/constants";

export default function InterviewReport() {
  const { id } = useParams();
  const { isAdmin } = useAuth();
  const {
    data: r,
    loading,
    error,
    reload,
  } = useFetch(() => getReport(id), [id]);

  if (loading) return <PageLoader text="Generating report..." />;
  if (error) return <ErrorMessage message={error} onRetry={reload} />;

  return (
    <>
      <Link
        to={isAdmin ? "/admin/candidates" : "/dashboard"}
        className="muted small"
        style={{ display: "inline-flex", gap: 6, alignItems: "center" }}
      >
        <ArrowLeft size={14} /> Back
      </Link>
      <div className="page-head">
        <div>
          <h1>Interview performance report</h1>
          <p className="muted">
            {r.candidateName} · {r.assessmentTitle} · {formatDate(r.date)}
          </p>
        </div>
        <div className="chips">
          {r.terminated && <Badge tone="bad">Ended after 3 warnings</Badge>}
          <Badge tone={LEVEL_TONE[r.performanceLevel]}>
            {formatLevel(r.performanceLevel)}
          </Badge>
        </div>
      </div>

      <div className="grid stats">
        <div className="card">
          <div className="muted small">Score</div>
          <div className="big-score grad">
            {r.score}
            <span className="muted" style={{ fontSize: "1.2rem" }}>
              /{r.maxScore}
            </span>
          </div>
        </div>
        <Stat label="Accuracy" value={r.accuracy} suffix="%" decimals={1} />
        <Stat label="Attempted" value={r.attempted} />
        <Stat label="Correct" value={r.correct} />
        <Stat label="Incorrect" value={r.incorrect} />
        <Stat label="Avg response (s)" value={r.avgResponseSec} decimals={1} />
        <Stat label="Proctor warnings" value={r.warnings} />
      </div>

      <div className="grid two">
        <div className="card">
          <h2>Correct vs incorrect</h2>
          {r.attempted > 0 ? (
            <Donut correct={r.correct} incorrect={r.incorrect} />
          ) : (
            <p className="muted">No answers recorded.</p>
          )}
        </div>
        <div className="card">
          <h2>Difficulty progression</h2>
          {r.difficultyProgression?.length ? (
            <Steps levels={r.difficultyProgression} />
          ) : (
            <p className="muted">No data.</p>
          )}
        </div>
      </div>

      <div className="grid two">
        <div className="card">
          <h2>Topic accuracy</h2>
          {r.topicBreakdown?.length ? (
            <HBars data={r.topicBreakdown} rag max={100} />
          ) : (
            <p className="muted">No data.</p>
          )}
        </div>
        <div className="card">
          <h2>By difficulty</h2>
          {r.difficultyBreakdown?.length ? (
            <HBars
              data={r.difficultyBreakdown}
              nameKey="difficulty"
              rag
              max={100}
            />
          ) : (
            <p className="muted">No data.</p>
          )}
        </div>
      </div>

      <div className="grid two">
        <div className="card">
          <h2>Strong topics</h2>
          {r.strongTopics?.length ? (
            <div className="chips">
              {r.strongTopics.map((t) => (
                <Badge key={t} tone="good">
                  {t}
                </Badge>
              ))}
            </div>
          ) : (
            <p className="muted">None yet.</p>
          )}
          <h2 style={{ marginTop: 22 }}>Weak topics</h2>
          {r.weakTopics?.length ? (
            <div className="chips">
              {r.weakTopics.map((t) => (
                <Badge key={t} tone="bad">
                  {t}
                </Badge>
              ))}
            </div>
          ) : (
            <p className="muted">None. Great.</p>
          )}
        </div>
        <div className="card">
          <h2>Recommendations</h2>
          {r.recommendations?.map((t, i) => (
            <div key={i} className="list-row small">
              {t}
            </div>
          ))}
        </div>
      </div>
    </>
  );
}
