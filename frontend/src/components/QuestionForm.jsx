import { useState } from "react";
import { Plus, Trash2 } from "lucide-react";
import { Modal } from "./ui";
import { CATEGORIES, DIFFICULTIES } from "../utils/constants";
import { createQuestion, updateQuestion } from "../services/adminService";
import { errMsg } from "../services/api";

const blank = {
  text: "",
  topic: "",
  category: "DSA",
  difficulty: "MEDIUM",
  explanation: "",
  timeLimitSec: 60,
  active: true,
  options: [
    { text: "", correct: true },
    { text: "", correct: false },
    { text: "", correct: false },
    { text: "", correct: false },
  ],
};

export default function QuestionForm({ initial, topics, onClose, onSaved }) {
  const [f, setF] = useState(() =>
    initial
      ? {
          ...initial,
          options: initial.options.map((o) => ({
            text: o.text,
            correct: o.correct,
          })),
        }
      : blank,
  );
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  const set = (k) => (e) => setF({ ...f, [k]: e.target.value });
  const setOpt = (i, patch) =>
    setF({
      ...f,
      options: f.options.map((o, j) => (j === i ? { ...o, ...patch } : o)),
    });
  const markCorrect = (i) =>
    setF({
      ...f,
      options: f.options.map((o, j) => ({ ...o, correct: j === i })),
    });
  const addOpt = () =>
    setF({ ...f, options: [...f.options, { text: "", correct: false }] });
  const removeOpt = (i) => {
    let opts = f.options.filter((_, j) => j !== i);
    if (!opts.some((o) => o.correct))
      opts = opts.map((o, j) => ({ ...o, correct: j === 0 }));
    setF({ ...f, options: opts });
  };

  const validate = () => {
    if (f.text.trim().length < 10)
      return "Question text must be at least 10 characters";
    if (!f.topic.trim()) return "Topic is required";
    if (f.options.some((o) => !o.text.trim()))
      return "Fill in every option (or remove the empty ones)";
    return "";
  };

  const submit = async (e) => {
    e.preventDefault();
    const problem = validate();
    if (problem) return setError(problem);
    setBusy(true);
    setError("");
    const body = { ...f, timeLimitSec: Number(f.timeLimitSec) || 60 };
    try {
      initial
        ? await updateQuestion(initial.id, body)
        : await createQuestion(body);
      onSaved(initial ? "Question updated" : "Question created");
    } catch (err) {
      setError(errMsg(err));
      setBusy(false);
    }
  };

  return (
    <Modal
      title={initial ? "Edit question" : "New question"}
      onClose={onClose}
      wide
    >
      <form onSubmit={submit}>
        <label className="field">
          <span>Question</span>
          <textarea
            className="input"
            rows={3}
            value={f.text}
            onChange={set("text")}
          />
        </label>
        <div className="form-row">
          <label className="field">
            <span>Topic</span>
            <input
              className="input"
              list="topic-list"
              value={f.topic}
              onChange={set("topic")}
            />
            <datalist id="topic-list">
              {topics.map((t) => (
                <option key={t} value={t} />
              ))}
            </datalist>
          </label>
          <label className="field">
            <span>Category</span>
            <select
              className="input"
              value={f.category}
              onChange={set("category")}
            >
              {CATEGORIES.map((c) => (
                <option key={c}>{c}</option>
              ))}
            </select>
          </label>
        </div>
        <div className="form-row">
          <label className="field">
            <span>Difficulty</span>
            <select
              className="input"
              value={f.difficulty}
              onChange={set("difficulty")}
            >
              {DIFFICULTIES.map((d) => (
                <option key={d}>{d}</option>
              ))}
            </select>
          </label>
          <label className="field">
            <span>Time limit (sec)</span>
            <input
              className="input"
              type="number"
              min="10"
              max="600"
              value={f.timeLimitSec}
              onChange={set("timeLimitSec")}
            />
          </label>
        </div>

        <div className="field">
          <span className="muted small">Options — select the correct one</span>
          {f.options.map((o, i) => (
            <div className="opt-row" key={i}>
              <input
                type="radio"
                name="correct"
                checked={o.correct}
                onChange={() => markCorrect(i)}
              />
              <input
                className="input"
                placeholder={`Option ${i + 1}`}
                value={o.text}
                onChange={(e) => setOpt(i, { text: e.target.value })}
              />
              <button
                type="button"
                className="icon-btn"
                disabled={f.options.length <= 2}
                onClick={() => removeOpt(i)}
              >
                <Trash2 size={16} />
              </button>
            </div>
          ))}
          {f.options.length < 6 && (
            <button type="button" className="btn" onClick={addOpt}>
              <Plus size={16} />
              Add option
            </button>
          )}
        </div>

        <label className="field">
          <span>Explanation (shown only if the assessment allows it)</span>
          <textarea
            className="input"
            rows={2}
            value={f.explanation}
            onChange={set("explanation")}
          />
        </label>
        <label
          className="field"
          style={{ display: "flex", gap: 10, alignItems: "center" }}
        >
          <input
            type="checkbox"
            checked={f.active}
            onChange={(e) => setF({ ...f, active: e.target.checked })}
          />{" "}
          Active
        </label>

        {error && <div className="form-error">{error}</div>}
        <div style={{ display: "flex", gap: 10, justifyContent: "flex-end" }}>
          <button type="button" className="btn" onClick={onClose}>
            Cancel
          </button>
          <button className="btn btn-primary" disabled={busy}>
            {busy ? "Saving..." : "Save question"}
          </button>
        </div>
      </form>
    </Modal>
  );
}
