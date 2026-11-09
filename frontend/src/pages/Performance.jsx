import { useState } from "react";
import { Link } from "react-router-dom";
import { useAuth } from "../hooks/useAuth";
import useFetch from "../hooks/useFetch";
import { candidateAnalytics, listReports } from "../services/candidateService";
import {
  Badge,
  Empty,
  ErrorMessage,
  PageLoader,
  Pager,
} from "../components/ui";
import { HBars } from "../components/Charts";
import { formatDate, formatLevel } from "../utils/format";
import { LEVEL_TONE } from "../utils/constants";

export default function Performance() {
  const { user } = useAuth();
  const [page, setPage] = useState(0);
  const analytics = useFetch(() => candidateAnalytics(user.id), [user.id]);
  const reports = useFetch(() => listReports(page, 8), [page]);

  if (analytics.loading && !analytics.data)
    return <PageLoader text="Loading performance..." />;
  if (analytics.error)
    return (
      <ErrorMessage message={analytics.error} onRetry={analytics.reload} />
    );
  const a = analytics.data;

  return (
    <>
      <div className="page-head">
        <div>
          <h1>Performance</h1>
          <p className="muted">
            {a.questionsAttempted} questions answered · {a.avgResponseSec}s
            average response
          </p>
        </div>
      </div>
      <div className="grid two">
        <div className="card">
          <h2>By difficulty</h2>
          {a.difficultyPerformance?.length ? (
            <HBars
              data={a.difficultyPerformance}
              nameKey="difficulty"
              rag
              max={100}
            />
          ) : (
            <p className="muted">No data yet.</p>
          )}
        </div>
        <div className="card">
          <h2>All topics</h2>
          {a.topicPerformance?.length ? (
            <HBars data={a.topicPerformance} rag max={100} />
          ) : (
            <p className="muted">No data yet.</p>
          )}
        </div>
      </div>

      <div className="card">
        <h2>Interview history</h2>
        {reports.loading ? (
          <PageLoader />
        ) : reports.error ? (
          <ErrorMessage message={reports.error} onRetry={reports.reload} />
        ) : reports.data.items.length === 0 ? (
          <Empty
            title="No interviews yet"
            to="/assessments"
            action="Start one"
          />
        ) : (
          <>
            <div className="table-wrap">
              <table>
                <thead>
                  <tr>
                    <th>Assessment</th>
                    <th>Date</th>
                    <th>Score</th>
                    <th>Accuracy</th>
                    <th>Level</th>
                    <th />
                  </tr>
                </thead>
                <tbody>
                  {reports.data.items.map((r) => (
                    <tr key={r.id}>
                      <td>{r.assessmentTitle}</td>
                      <td>{formatDate(r.createdAt)}</td>
                      <td>
                        {r.score}/{r.maxScore}
                      </td>
                      <td>{r.accuracy}%</td>
                      <td>
                        <Badge tone={LEVEL_TONE[r.level]}>
                          {formatLevel(r.level)}
                        </Badge>
                      </td>
                      <td>
                        <Link to={`/report/${r.id}`} className="grad">
                          View
                        </Link>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            <Pager
              page={page}
              totalPages={reports.data.totalPages}
              onChange={setPage}
            />
          </>
        )}
      </div>
    </>
  );
}
