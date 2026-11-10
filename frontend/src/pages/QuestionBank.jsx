import { useEffect, useState } from "react";
import { Pencil, Plus, Search, Trash2 } from "lucide-react";
import useFetch from "../hooks/useFetch";
import {
  adminAnalytics,
  deleteQuestion,
  listQuestions,
} from "../services/adminService";
import { errMsg } from "../services/api";
import { useToast } from "../context/ToastContext";
import {
  Badge,
  DifficultyBadge,
  Empty,
  ErrorMessage,
  PageLoader,
  Pager,
} from "../components/ui";
import ConfirmDialog from "../components/ConfirmDialog";
import QuestionForm from "../components/QuestionForm";
import { DIFFICULTIES } from "../utils/constants";

export default function QuestionBank() {
  const toast = useToast();
  const [text, setText] = useState("");
  const [search, setSearch] = useState("");
  const [topic, setTopic] = useState("");
  const [difficulty, setDifficulty] = useState("");
  const [status, setStatus] = useState("");
  const [page, setPage] = useState(0);
  const [editing, setEditing] = useState(null); // null | 'new' | question
  const [deleting, setDeleting] = useState(null);
  const [busy, setBusy] = useState(false);

  // Debounce typing so we don't hit the API on every keystroke.
  useEffect(() => {
    const t = setTimeout(() => {
      setSearch(text);
      setPage(0);
    }, 350);
    return () => clearTimeout(t);
  }, [text]);

  const { data, loading, error, reload } = useFetch(
    () => listQuestions({ search, topic, difficulty, status, page, size: 10 }),
    [search, topic, difficulty, status, page],
  );
  const stats = useFetch(() => adminAnalytics(), []);
  const topics = (stats.data?.topicDistribution ?? []).map((t) => t.label);

  const filter = (setter) => (e) => {
    setter(e.target.value);
    setPage(0);
  };

  const confirmDelete = async () => {
    setBusy(true);
    try {
      const res = await deleteQuestion(deleting.id);
      toast.success(res.message);
      setDeleting(null);
      reload();
    } catch (e) {
      toast.error(errMsg(e));
    } finally {
      setBusy(false);
    }
  };

  return (
    <>
      <div className="page-head">
        <div>
          <h1>Question bank</h1>
          <p className="muted">{data ? `${data.totalItems} questions` : " "}</p>
        </div>
        <button className="btn btn-primary" onClick={() => setEditing("new")}>
          <Plus size={16} />
          Add question
        </button>
      </div>

      <div className="toolbar">
        <div style={{ position: "relative", flex: 2, minWidth: 180 }}>
          <Search
            size={16}
            style={{
              position: "absolute",
              left: 12,
              top: 14,
              color: "#8e9a90",
            }}
          />
          <input
            className="input"
            style={{ paddingLeft: 36 }}
            placeholder="Search questions..."
            value={text}
            onChange={(e) => setText(e.target.value)}
          />
        </div>
        <select className="input" value={topic} onChange={filter(setTopic)}>
          <option value="">All topics</option>
          {topics.map((t) => (
            <option key={t}>{t}</option>
          ))}
        </select>
        <select
          className="input"
          value={difficulty}
          onChange={filter(setDifficulty)}
        >
          <option value="">All difficulties</option>
          {DIFFICULTIES.map((d) => (
            <option key={d}>{d}</option>
          ))}
        </select>
        <select className="input" value={status} onChange={filter(setStatus)}>
          <option value="">Any status</option>
          <option value="active">Active</option>
          <option value="inactive">Inactive</option>
        </select>
      </div>

      <div className="card">
        {loading ? (
          <PageLoader text="Loading questions..." />
        ) : error ? (
          <ErrorMessage message={error} onRetry={reload} />
        ) : data.items.length === 0 ? (
          <Empty title="No questions match" hint="Try changing the filters." />
        ) : (
          <>
            <div className="table-wrap">
              <table>
                <thead>
                  <tr>
                    <th>Question</th>
                    <th>Topic</th>
                    <th>Level</th>
                    <th>Status</th>
                    <th />
                  </tr>
                </thead>
                <tbody>
                  {data.items.map((q) => (
                    <tr key={q.id}>
                      <td>
                        <div className="truncate">{q.text}</div>
                      </td>
                      <td>{q.topic}</td>
                      <td>
                        <DifficultyBadge level={q.difficulty} />
                      </td>
                      <td>
                        <Badge tone={q.active ? "good" : ""}>
                          {q.active ? "Active" : "Inactive"}
                        </Badge>
                      </td>
                      <td style={{ whiteSpace: "nowrap" }}>
                        <button
                          className="icon-btn"
                          onClick={() => setEditing(q)}
                          aria-label="Edit"
                        >
                          <Pencil size={16} />
                        </button>{" "}
                        <button
                          className="icon-btn btn-danger"
                          onClick={() => setDeleting(q)}
                          aria-label="Delete"
                        >
                          <Trash2 size={16} />
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            <Pager
              page={page}
              totalPages={data.totalPages}
              onChange={setPage}
            />
          </>
        )}
      </div>

      {editing && (
        <QuestionForm
          initial={editing === "new" ? null : editing}
          topics={topics}
          onClose={() => setEditing(null)}
          onSaved={(msg) => {
            toast.success(msg);
            setEditing(null);
            reload();
            stats.reload();
          }}
        />
      )}
      {deleting && (
        <ConfirmDialog
          title="Delete question?"
          confirmText="Delete"
          busy={busy}
          message="If this question was already used in interviews it will be deactivated instead, so old reports stay intact."
          onCancel={() => setDeleting(null)}
          onConfirm={confirmDelete}
        />
      )}
    </>
  );
}
