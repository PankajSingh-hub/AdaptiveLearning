import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Clock, ListChecks, Play } from "lucide-react";
import useFetch from "../hooks/useFetch";
import { listAssessments, startInterview } from "../services/candidateService";
import { errMsg } from "../services/api";
import { useToast } from "../context/ToastContext";
import { Badge, Empty, ErrorMessage, PageLoader } from "../components/ui";

export default function AssessmentList() {
  const { data, loading, error, reload } = useFetch(
    () => listAssessments(),
    [],
  );
  const [startingId, setStartingId] = useState(null);
  const navigate = useNavigate();
  const toast = useToast();

  const start = async (id) => {
    if (startingId) return;
    setStartingId(id);
    try {
      const res = await startInterview(id);
      navigate(`/interview/${res.sessionId}`);
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
          <p className="muted">Difficulty adapts to your answers as you go.</p>
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
          {data.map((a) => (
            <div key={a.id} className="card hover">
              <h2>{a.title}</h2>
              <p className="muted" style={{ minHeight: 44 }}>
                {a.description}
              </p>
              <div className="chips" style={{ margin: "14px 0 18px" }}>
                <Badge>
                  <ListChecks size={12} style={{ verticalAlign: -2 }} />{" "}
                  {a.totalQuestions} questions
                </Badge>
                <Badge>
                  <Clock size={12} style={{ verticalAlign: -2 }} />{" "}
                  {a.durationMinutes} min
                </Badge>
              </div>
              <button
                className="btn btn-primary"
                disabled={startingId !== null}
                onClick={() => start(a.id)}
              >
                <Play size={16} />
                {startingId === a.id ? "Starting..." : "Start interview"}
              </button>
            </div>
          ))}
        </div>
      )}
    </>
  );
}
