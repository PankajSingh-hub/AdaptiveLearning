import { useEffect, useRef, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import {
  ChevronRight,
  LogOut,
  Moon,
  Palette,
  Settings,
  Sun,
} from "lucide-react";
import { useAuth } from "../hooks/useAuth";
import { useTheme } from "../hooks/useTheme";
import useFetch from "../hooks/useFetch";
import { getStreak } from "../services/profileService";
import ConfirmDialog from "./ConfirmDialog";
import { Badge } from "./ui";

// Mounted only while the menu is open, so the streak is fetched fresh every time.
function StreakCard() {
  const { data, loading, error } = useFetch(() => getStreak(), []);
  if (loading)
    return <div className="streak muted small">Loading streak...</div>;
  if (error || !data)
    return (
      <div className="streak muted small">Streak unavailable right now</div>
    );
  return (
    <div className="streak">
      <div className="streak-top">
        <span>Current streak</span>
        <span className="streak-count">
          ⚡ {data.current} <small>{data.current === 1 ? "day" : "days"}</small>
        </span>
      </div>
      <div className="streak-days">
        {data.days.map((d) => (
          <div
            key={d.date}
            className={`sday ${d.active ? "on" : ""} ${d.today ? "today" : ""}`}
            title={d.date}
          >
            <span className="bolt">{d.active ? "⚡" : ""}</span>
            {d.label}
          </div>
        ))}
      </div>
      <p className="streak-note">
        A day counts when you answer a quiz question or submit code.
      </p>
    </div>
  );
}

export default function ProfileMenu() {
  const { user, logout, isAdmin } = useAuth();
  const { theme, setTheme } = useTheme();
  const navigate = useNavigate();
  const [open, setOpen] = useState(false);
  const [showTheme, setShowTheme] = useState(false);
  const [confirm, setConfirm] = useState(false);
  const box = useRef(null);

  useEffect(() => {
    if (!open) return;
    const onDown = (e) => {
      if (box.current && !box.current.contains(e.target)) setOpen(false);
    };
    const onKey = (e) => {
      if (e.key === "Escape") setOpen(false);
    };
    document.addEventListener("mousedown", onDown);
    document.addEventListener("keydown", onKey);
    return () => {
      document.removeEventListener("mousedown", onDown);
      document.removeEventListener("keydown", onKey);
    };
  }, [open]);

  const signOut = () => {
    setConfirm(false);
    setOpen(false);
    logout();
    navigate("/login", { replace: true });
  };

  return (
    <div className="pm" ref={box}>
      <button
        className="avatar-btn"
        onClick={() => setOpen((o) => !o)}
        aria-haspopup="menu"
        aria-expanded={open}
        aria-label="Account menu"
      >
        <span className="avatar">{user.name?.[0]?.toUpperCase()}</span>
      </button>

      {open && (
        <div className="pm-pop" role="menu">
          <div className="pm-head">
            <span className="avatar">{user.name?.[0]?.toUpperCase()}</span>
            <div style={{ minWidth: 0 }}>
              <div className="pm-name">{user.name}</div>
              <div className="pm-email">{user.email}</div>
              <Badge tone="accent">{user.role}</Badge>
            </div>
          </div>

          {!isAdmin && <StreakCard />}

          <Link
            to="/settings"
            className="pm-item"
            role="menuitem"
            onClick={() => setOpen(false)}
          >
            <Settings size={18} />
            Settings
          </Link>

          <button
            className={`pm-item ${showTheme ? "open" : ""}`}
            onClick={() => setShowTheme((s) => !s)}
            aria-expanded={showTheme}
          >
            <Palette size={18} />
            Appearance
            <ChevronRight size={16} className="chev" />
          </button>
          {showTheme && (
            <div className="seg">
              <button
                className={theme === "dark" ? "on" : ""}
                onClick={() => setTheme("dark")}
              >
                <Moon size={15} />
                Black
              </button>
              <button
                className={theme === "light" ? "on" : ""}
                onClick={() => setTheme("light")}
              >
                <Sun size={15} />
                White
              </button>
            </div>
          )}

          <div className="pm-sep" />
          <button className="pm-item" onClick={() => setConfirm(true)}>
            <LogOut size={18} />
            Sign out
          </button>
        </div>
      )}

      {confirm && (
        <ConfirmDialog
          title="Sign out?"
          confirmText="Sign out"
          message="You will need to sign in again to continue. Your results and progress stay saved."
          onCancel={() => setConfirm(false)}
          onConfirm={signOut}
        />
      )}
    </div>
  );
}
