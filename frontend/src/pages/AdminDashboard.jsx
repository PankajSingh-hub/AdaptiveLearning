import { Link } from "react-router-dom";
import {
  Award,
  CheckCircle2,
  ClipboardList,
  Library,
  Target,
  Users,
} from "lucide-react";
import useFetch from "../hooks/useFetch";
import { adminAnalytics } from "../services/adminService";
import { Badge, Empty, ErrorMessage, PageLoader, Stat } from "../components/ui";
import { HBars } from "../components/Charts";
import { formatDate, formatLevel } from "../utils/format";
import { LEVEL_TONE } from "../utils/constants";

export default function AdminDashboard() {
  const {
    data: a,
    loading,
    error,
    reload,
  } = useFetch(() => adminAnalytics(), []);
  if (loading) return <PageLoader text="Loading overview..." />;
  if (error) return <ErrorMessage message={error} onRetry={reload} />;

  return (
    <>
      <div className="page-head">
        <div>
          <h1>Overview</h1>
          <p className="muted">Live numbers from your database.</p>
        </div>
      </div>
      <div className="grid stats">
        <Stat label="Candidates" value={a.totalCandidates} icon={Users} />
        <Stat
          label="Assessments"
          value={a.totalAssessments}
          icon={ClipboardList}
        />
        <Stat label="Questions" value={a.totalQuestions} icon={Library} />
        <Stat
          label="Completed interviews"
          value={a.completedInterviews}
          icon={CheckCircle2}
        />
        <Stat label="Avg score" value={a.avgScore} decimals={1} icon={Award} />
        <Stat
          label="Avg accuracy"
          value={a.avgAccuracy}
          suffix="%"
          decimals={1}
          icon={Target}
        />
      </div>

      <div className="grid two">
        <div className="card">
          <h2>Questions by difficulty</h2>
          <HBars
            data={a.difficultyDistribution}
            nameKey="label"
            dataKey="count"
            unit=""
          />
        </div>
        <div className="card">
          <h2>Questions by topic</h2>
          <HBars
            data={a.topicDistribution}
            nameKey="label"
            dataKey="count"
            unit=""
          />
        </div>
      </div>

      <div className="grid two">
        <div className="card">
          <h2>Recent interviews</h2>
          {a.recentInterviews.length === 0 ? (
            <Empty title="No interviews yet" />
          ) : (
            a.recentInterviews.map((r) => (
              <Link key={r.id} to={`/report/${r.id}`} className="list-row">
                <div>
                  <div>{r.candidateName}</div>
                  <div className="muted small">
                    {r.assessmentTitle} · {formatDate(r.createdAt)}
                  </div>
                </div>
                <div style={{ textAlign: "right" }}>
                  <strong>{r.accuracy}%</strong>
                  <div>
                    <Badge tone={LEVEL_TONE[r.level]}>
                      {formatLevel(r.level)}
                    </Badge>
                  </div>
                </div>
              </Link>
            ))
          )}
        </div>
        <div className="card">
          <h2>Top candidates</h2>
          {a.topCandidates.length === 0 ? (
            <Empty title="No data yet" />
          ) : (
            a.topCandidates.map((c, i) => (
              <div key={c.candidateId} className="list-row">
                <div>
                  <span className="muted">#{i + 1}</span> {c.name}
                  <div className="muted small">{c.interviews} interview(s)</div>
                </div>
                <strong>{c.avgAccuracy}%</strong>
              </div>
            ))
          )}
        </div>
      </div>
    </>
  );
}
