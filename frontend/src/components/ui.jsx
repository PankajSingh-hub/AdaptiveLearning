import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { ChevronLeft, ChevronRight, X } from "lucide-react";
import { DIFF_TONE } from "../utils/constants";

export const Spinner = () => <div className="spinner" />;

export const PageLoader = ({ text = "Loading..." }) => (
  <div className="center">
    <Spinner />
    <span className="muted">{text}</span>
  </div>
);

export const ErrorMessage = ({ message, onRetry }) => (
  <div className="card error-box center">
    <strong>Something went wrong</strong>
    <span className="muted">{message}</span>
    {onRetry && (
      <button className="btn" onClick={onRetry}>
        Try again
      </button>
    )}
  </div>
);

export const Empty = ({ title, hint, to, action }) => (
  <div className="center">
    <strong>{title}</strong>
    {hint && <span className="muted">{hint}</span>}
    {to && (
      <Link to={to} className="btn btn-primary">
        {action}
      </Link>
    )}
  </div>
);

export const Badge = ({ tone = "", children }) => (
  <span className={`badge ${tone}`}>{children}</span>
);
export const DifficultyBadge = ({ level }) => (
  <Badge tone={DIFF_TONE[level]}>{level}</Badge>
);
export const Progress = ({ value }) => (
  <div className="bar">
    <i style={{ width: `${Math.min(100, value)}%` }} />
  </div>
);

function useCountUp(target, ms = 700) {
  const [v, setV] = useState(0);
  useEffect(() => {
    const end = Number(target) || 0;
    let raf, t0;
    const step = (t) => {
      t0 = t0 ?? t;
      const p = Math.min((t - t0) / ms, 1);
      setV(end * (1 - Math.pow(1 - p, 3)));
      if (p < 1) raf = requestAnimationFrame(step);
    };
    raf = requestAnimationFrame(step);
    return () => cancelAnimationFrame(raf);
  }, [target, ms]);
  return v;
}

export function Stat({ label, value, suffix = "", decimals = 0, icon: Icon }) {
  const v = useCountUp(value);
  return (
    <div className="card stat">
      <div className="stat-top">
        <span>{label}</span>
        {Icon && <Icon size={18} />}
      </div>
      <div className="stat-value">
        {v.toFixed(decimals)}
        {suffix}
      </div>
    </div>
  );
}

export function Pager({ page, totalPages, onChange }) {
  if (totalPages <= 1) return null;
  return (
    <div className="pager">
      <button
        className="icon-btn"
        disabled={page === 0}
        onClick={() => onChange(page - 1)}
      >
        <ChevronLeft size={18} />
      </button>
      <span>
        Page {page + 1} of {totalPages}
      </span>
      <button
        className="icon-btn"
        disabled={page + 1 >= totalPages}
        onClick={() => onChange(page + 1)}
      >
        <ChevronRight size={18} />
      </button>
    </div>
  );
}

export function Modal({ title, onClose, children, wide }) {
  useEffect(() => {
    const onKey = (e) => e.key === "Escape" && onClose();
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onClose]);
  return (
    <div
      className="modal-back"
      onMouseDown={(e) => e.target === e.currentTarget && onClose()}
    >
      <div className={`modal card ${wide ? "wide" : ""}`}>
        <div className="modal-head">
          <h3>{title}</h3>
          <button className="icon-btn" onClick={onClose} aria-label="Close">
            <X size={18} />
          </button>
        </div>
        {children}
      </div>
    </div>
  );
}
