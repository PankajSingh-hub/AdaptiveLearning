import { Link } from "react-router-dom";
import {
  ArrowRight,
  Award,
  CheckCircle2,
  ClipboardList,
  Target,
} from "lucide-react";
import { useAuth } from "../hooks/useAuth";
import useFetch from "../hooks/useFetch";
import {
  candidateAnalytics,
  listAssessments,
} from "../services/candidateService";
import { Badge, Empty, ErrorMessage, PageLoader, Stat } from "../components/ui";
import { HBars, Trend } from "../components/Charts";
import { formatDate, formatLevel } from "../utils/format";
import { LEVEL_TONE } from "../utils/constants";

export default function CandidateDashboard() {
  const { user } = useAuth();
  const {
    data: a,
    loading,
    error,
    reload,
  } = useFetch(() => candidateAnalytics(user.id), [user.id]);
  const { data: assessments } = useFetch(() => listAssessments(), []);

  if (loading) return <PageLoader text="Loading dashboard..." />;
  if (error) return <ErrorMessage message={error} onRetry={reload} />;

  const trend = (a.trend ?? []).map((p) => ({
    label: formatDate(p.date),
    accuracy: p.accuracy,
  }));
  const topics = a.topicPerformance ?? [];

  return (
    <>
      <div className="page-head">
        <div>
          <h1>
            Welcome back,{" "}
            <span className="grad">{user.name.split(" ")[0]}</span>
          </h1>
          <p className="muted">Here is how your interviews are going.</p>
        </div>
        <Link to="/assessments" className="btn btn-primary">
          Start assessment <ArrowRight size={16} />
        </Link>
      </div>

      <div className="grid stats">
        <Stat
          label="Available assessments"
          value={assessments?.length ?? 0}
          icon={ClipboardList}
        />
        <Stat
          label="Completed"
          value={a.completedInterviews}
          icon={CheckCircle2}
        />
        <Stat
          label="Average score"
          value={a.avgScore}
          decimals={1}
          icon={Award}
        />
        <Stat
          label="Accuracy"
          value={a.accuracy}
          suffix="%"
          decimals={1}
          icon={Target}
        />
      </div>

      {a.completedInterviews === 0 ? (
        <div className="card">
          <Empty
            title="No interviews yet"
            hint="Take your first adaptive assessment to see analytics here."
            to="/assessments"
            action="Browse assessments"
          />
        </div>
      ) : (
        <>
          <div className="grid two">
            <div className="card">
              <h2>Accuracy over time</h2>
              <Trend data={trend} />
            </div>
            <div className="card">
              <h2>Topic performance</h2>
              <HBars data={topics} rag max={100} />
            </div>
          </div>
          <div className="grid two">
            <div className="card">
              <h2>Recommended topics</h2>
              {a.recommendedTopics?.length ? (
                <div className="chips">
                  {a.recommendedTopics.map((t) => (
                    <Badge key={t} tone="warn">
                      {t}
                    </Badge>
                  ))}
                </div>
              ) : (
                <p className="muted">No weak topics detected. Nice work.</p>
              )}
            </div>
            <div className="card">
              <h2>Recent interviews</h2>
              {a.recentReports?.map((r) => (
                <Link key={r.id} to={`/report/${r.id}`} className="list-row">
                  <div>
                    <div>{r.assessmentTitle}</div>
                    <div className="muted small">{formatDate(r.createdAt)}</div>
                  </div>
                  <div style={{ textAlign: "right" }}>
                    <strong>
                      {r.score}/{r.maxScore}
                    </strong>
                    <div>
                      <Badge tone={LEVEL_TONE[r.level]}>
                        {formatLevel(r.level)}
                      </Badge>
                    </div>
                  </div>
                </Link>
              ))}
            </div>
          </div>
        </>
      )}
    </>
  );
}
