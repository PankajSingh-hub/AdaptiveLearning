import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Clock, Code2, ListChecks, Play } from "lucide-react";
import useFetch from "../hooks/useFetch";
import {
  listAssessments,
  startCoding,
  startInterview,
} from "../services/candidateService";
import { errMsg } from "../services/api";
import { useToast } from "../context/toastCtx";
import { Badge, Empty, ErrorMessage, PageLoader } from "../components/ui";

export default function AssessmentList() {
  const { data, loading, error, reload } = useFetch(
    () => listAssessments(),
    [],
  );
  const [startingId, setStartingId] = useState(null);
  const navigate = useNavigate();
  const toast = useToast();

  const start = async (a) => {
    if (startingId) return;
    setStartingId(a.id);
    try {
      if (a.type === "CODING") {
        const res = await startCoding(a.id);
        navigate(`/coding/${res.sessionId}`);
      } else {
        const res = await startInterview(a.id);
        navigate(`/interview/${res.sessionId}`);
      }
    } catch (e) {
      toast.error(errMsg(e));
      setStartingId(null);
    }
  };

  if (loading) return <PageLoader text="Loading assessments..." />;
  if (error) return <ErrorMessage message={error} onRetry={reload} />;

  return (
    <>
      <div className="page-head">
        <div>
          <h1>Assessments</h1>
          <p className="muted">
            Quizzes adapt to your answers; coding challenges give progressive
            hints.
          </p>
        </div>
      </div>
      {data.length === 0 ? (
        <div className="card">
          <Empty
            title="No assessments available"
            hint="Ask an admin to publish one."
          />
        </div>
      ) : (
        <div className="grid two">
          {data.map((a) => {
            const coding = a.type === "CODING";
            return (
              <div key={a.id} className="card hover">
                <div
                  style={{
                    display: "flex",
                    justifyContent: "space-between",
                    gap: 10,
                  }}
                >
                  <h2>{a.title}</h2>
                  <Badge tone={coding ? "accent" : ""}>
                    {coding ? "Coding" : "Quiz"}
                  </Badge>
                </div>
                <p className="muted" style={{ minHeight: 44 }}>
                  {a.description}
                </p>
                <div className="chips" style={{ margin: "14px 0 18px" }}>
                  <Badge>
                    {coding ? (
                      <Code2 size={12} style={{ verticalAlign: -2 }} />
                    ) : (
                      <ListChecks size={12} style={{ verticalAlign: -2 }} />
                    )}{" "}
                    {a.totalQuestions} {coding ? "problems" : "questions"}
                  </Badge>
                  <Badge>
                    <Clock size={12} style={{ verticalAlign: -2 }} />{" "}
                    {a.durationMinutes} min
                  </Badge>
                </div>
                <button
                  className="btn btn-primary"
                  disabled={startingId !== null}
                  onClick={() => start(a)}
                >
                  <Play size={16} />
                  {startingId === a.id
                    ? "Starting..."
                    : coding
                      ? "Start challenge"
                      : "Start interview"}
                </button>
              </div>
            );
          })}
        </div>
      )}
    </>
  );
}
