import { useState } from "react";
import { Pencil, Plus, Trash2 } from "lucide-react";
import useFetch from "../hooks/useFetch";
import {
  createAssessment,
  deleteAssessment,
  listAllAssessments,
  updateAssessment,
} from "../services/adminService";
import { errMsg } from "../services/api";
import { useToast } from "../context/toastCtx";
import {
  Badge,
  Empty,
  ErrorMessage,
  Modal,
  PageLoader,
} from "../components/ui";
import ConfirmDialog from "../components/ConfirmDialog";

const blank = {
  title: "",
  description: "",
  totalQuestions: 10,
  durationMinutes: 30,
  showExplanations: false,
  active: true,
};

function AssessmentForm({ initial, onClose, onSaved }) {
  const [f, setF] = useState(initial ?? blank);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const set = (k) => (e) => setF({ ...f, [k]: e.target.value });
  const check = (k) => (e) => setF({ ...f, [k]: e.target.checked });

  const submit = async (e) => {
    e.preventDefault();
    const total = Number(f.totalQuestions),
      mins = Number(f.durationMinutes);
    if (f.title.trim().length < 3)
      return setError("Title must be at least 3 characters");
    if (!(total >= 1 && total <= 50))
      return setError("Questions must be between 1 and 50");
    if (!(mins >= 1 && mins <= 240))
      return setError("Duration must be between 1 and 240 minutes");
    setBusy(true);
    setError("");
    const body = {
      ...f,
      title: f.title.trim(),
      totalQuestions: total,
      durationMinutes: mins,
    };
    try {
      initial
        ? await updateAssessment(initial.id, body)
        : await createAssessment(body);
      onSaved(initial ? "Assessment updated" : "Assessment created");
    } catch (err) {
      setError(errMsg(err));
      setBusy(false);
    }
  };

  return (
    <Modal
      title={initial ? "Edit assessment" : "New assessment"}
      onClose={onClose}
    >
      <form onSubmit={submit}>
        <label className="field">
          <span>Title</span>
          <input className="input" value={f.title} onChange={set("title")} />
        </label>
        <label className="field">
          <span>Description</span>
          <textarea
            className="input"
            rows={2}
            value={f.description}
            onChange={set("description")}
          />
        </label>
        <div className="form-row">
          <label className="field">
            <span>Questions</span>
            <input
              className="input"
              type="number"
              value={f.totalQuestions}
              onChange={set("totalQuestions")}
            />
          </label>
          <label className="field">
            <span>Duration (min)</span>
            <input
              className="input"
              type="number"
              value={f.durationMinutes}
              onChange={set("durationMinutes")}
            />
          </label>
        </div>
        <label
          className="field"
          style={{ display: "flex", gap: 10, alignItems: "center" }}
        >
          <input
            type="checkbox"
            checked={f.showExplanations}
            onChange={check("showExplanations")}
          />{" "}
          Show answer + explanation after each question
        </label>
        <label
          className="field"
          style={{ display: "flex", gap: 10, alignItems: "center" }}
        >
          <input
            type="checkbox"
            checked={f.active}
            onChange={check("active")}
          />{" "}
          Active (visible to candidates)
        </label>
        {error && <div className="form-error">{error}</div>}
        <div style={{ display: "flex", gap: 10, justifyContent: "flex-end" }}>
          <button type="button" className="btn" onClick={onClose}>
            Cancel
          </button>
          <button className="btn btn-primary" disabled={busy}>
            {busy ? "Saving..." : "Save"}
          </button>
        </div>
      </form>
    </Modal>
  );
}

export default function AssessmentManagement() {
  const toast = useToast();
  const { data, loading, error, reload } = useFetch(
    () => listAllAssessments(),
    [],
  );
  const [editing, setEditing] = useState(null);
  const [deleting, setDeleting] = useState(null);
  const [busy, setBusy] = useState(false);

  const confirmDelete = async () => {
    setBusy(true);
    try {
      const res = await deleteAssessment(deleting.id);
      toast.success(res.message);
      setDeleting(null);
      reload();
    } catch (e) {
      toast.error(errMsg(e));
    } finally {
      setBusy(false);
    }
  };

  if (loading) return <PageLoader text="Loading assessments..." />;
  if (error) return <ErrorMessage message={error} onRetry={reload} />;

  return (
    <>
      <div className="page-head">
        <div>
          <h1>Assessments</h1>
          <p className="muted">Configure what candidates can take.</p>
        </div>
        <button className="btn btn-primary" onClick={() => setEditing("new")}>
          <Plus size={16} />
          New assessment
        </button>
      </div>
      {data.length === 0 ? (
        <div className="card">
          <Empty title="No assessments yet" />
        </div>
      ) : (
        <div className="grid two">
          {data.map((a) => (
            <div key={a.id} className="card hover">
              <div
                style={{
                  display: "flex",
                  justifyContent: "space-between",
                  gap: 10,
                }}
              >
                <h2>{a.title}</h2>
                <Badge tone={a.active ? "good" : ""}>
                  {a.active ? "Active" : "Inactive"}
                </Badge>
              </div>
              <p className="muted small" style={{ minHeight: 40 }}>
                {a.description}
              </p>
              <div className="chips" style={{ margin: "12px 0 16px" }}>
                <Badge>{a.totalQuestions} questions</Badge>
                <Badge>{a.durationMinutes} min</Badge>
                {a.showExplanations && (
                  <Badge tone="accent">Shows explanations</Badge>
                )}
              </div>
              <button className="btn" onClick={() => setEditing(a)}>
                <Pencil size={15} />
                Edit
              </button>{" "}
              <button className="btn btn-danger" onClick={() => setDeleting(a)}>
                <Trash2 size={15} />
                Delete
              </button>
            </div>
          ))}
        </div>
      )}
      {editing && (
        <AssessmentForm
          initial={editing === "new" ? null : editing}
          onClose={() => setEditing(null)}
          onSaved={(m) => {
            toast.success(m);
            setEditing(null);
            reload();
          }}
        />
      )}
      {deleting && (
        <ConfirmDialog
          title="Delete assessment?"
          confirmText="Delete"
          busy={busy}
          message="If candidates already took it, it will be deactivated instead so their reports stay intact."
          onCancel={() => setDeleting(null)}
          onConfirm={confirmDelete}
        />
      )}
    </>
  );
}
