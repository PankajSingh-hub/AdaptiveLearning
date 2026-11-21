import { useState } from "react";
import {
  Check,
  Circle,
  KeyRound,
  Moon,
  Palette,
  Sun,
  User,
} from "lucide-react";
import useFetch from "../hooks/useFetch";
import { useAuth } from "../hooks/useAuth";
import { useTheme } from "../hooks/useTheme";
import { useToast } from "../context/toastCtx";
import {
  changePassword,
  getProfile,
  updateProfile,
} from "../services/profileService";
import { errMsg } from "../services/api";
import { Badge, ErrorMessage, PageLoader } from "../components/ui";
import { formatDate } from "../utils/format";

function Row({ label, children }) {
  return (
    <div className="kvrow">
      <span>{label}</span>
      <div>{children}</div>
    </div>
  );
}

function ProfileTab({ profile, reload }) {
  const { updateUser, isAdmin } = useAuth();
  const toast = useToast();
  const [name, setName] = useState(profile.name);
  const [busy, setBusy] = useState(false);
  const dirty = name.trim() !== profile.name && name.trim().length >= 2;

  const save = async (e) => {
    e.preventDefault();
    if (busy || !dirty) return;
    setBusy(true);
    try {
      const u = await updateProfile(name.trim());
      updateUser({ name: u.name });
      toast.success("Name updated");
      reload();
    } catch (err) {
      toast.error(errMsg(err));
    } finally {
      setBusy(false);
    }
  };

  return (
    <>
      <div className="card">
        <h2>Profile</h2>
        <Row label="User ID">
          <span className="mono">#{profile.id}</span>
        </Row>
        <Row label="Name">
          <form
            onSubmit={save}
            style={{ display: "flex", gap: 10, flexWrap: "wrap" }}
          >
            <input
              className="input"
              style={{ flex: 1, minWidth: 180 }}
              value={name}
              disabled={profile.isDemo}
              onChange={(e) => setName(e.target.value)}
              maxLength={100}
            />
            <button
              className="btn btn-primary"
              disabled={!dirty || busy || profile.isDemo}
            >
              {busy ? "Saving..." : "Save"}
            </button>
          </form>
        </Row>
        <Row label="Email">
          <span style={{ overflowWrap: "anywhere" }}>{profile.email}</span>{" "}
          <Badge tone={profile.emailVerified ? "good" : ""}>
            {profile.emailVerified ? "Verified" : "Not verified"}
          </Badge>
          <div className="muted small" style={{ marginTop: 4 }}>
            Email cannot be changed here.
          </div>
        </Row>
        <Row label="Role">
          <Badge tone="accent">{profile.role}</Badge>
        </Row>
        <Row label="Member since">{formatDate(profile.joinedAt)}</Row>
        {profile.isDemo && (
          <p className="muted small" style={{ marginTop: 10 }}>
            This is the shared demo account, so its profile is read-only.
          </p>
        )}
      </div>

      {!isAdmin && (
        <div className="card">
          <h2>Your activity</h2>
          <div className="mini-stats">
            <div className="mini">
              <b>{profile.interviewsCompleted}</b>
              <span>Quizzes completed</span>
            </div>
            <div className="mini">
              <b>{profile.questionsAnswered}</b>
              <span>Questions answered</span>
            </div>
            <div className="mini">
              <b>{profile.codingSolved}</b>
              <span>Coding problems solved</span>
            </div>
            <div className="mini">
              <b>⚡ {profile.currentStreak}</b>
              <span>Day streak</span>
            </div>
          </div>
        </div>
      )}
    </>
  );
}

function SecurityTab({ profile }) {
  const toast = useToast();
  const [f, setF] = useState({ current: "", next: "", confirm: "" });
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const set = (k) => (e) => setF({ ...f, [k]: e.target.value });

  const rules = {
    length: f.next.length >= 8,
    mix: /[A-Za-z]/.test(f.next) && /\d/.test(f.next),
    match: f.next.length > 0 && f.next === f.confirm,
  };

  const submit = async (e) => {
    e.preventDefault();
    if (busy) return;
    setError("");
    if (!f.current) return setError("Enter your current password");
    if (!rules.length || !rules.mix)
      return setError(
        "New password must be 8+ characters with a letter and a number",
      );
    if (!rules.match) return setError("New passwords do not match");
    setBusy(true);
    try {
      await changePassword(f.current, f.next, f.confirm);
      toast.success("Password updated");
      setF({ current: "", next: "", confirm: "" });
    } catch (err) {
      setError(errMsg(err));
    } finally {
      setBusy(false);
    }
  };

  if (profile.isDemo) {
    return (
      <div className="card">
        <h2>Password</h2>
        <p className="muted">
          The shared demo account signs in without a password, so there is
          nothing to change.
        </p>
      </div>
    );
  }

  return (
    <div className="card">
      <h2>Change password</h2>
      <form onSubmit={submit} style={{ maxWidth: 420 }}>
        <label className="field">
          <span>Current password</span>
          <input
            className="input"
            type="password"
            value={f.current}
            onChange={set("current")}
            autoComplete="current-password"
          />
        </label>
        <label className="field">
          <span>New password</span>
          <input
            className="input"
            type="password"
            value={f.next}
            onChange={set("next")}
            autoComplete="new-password"
          />
        </label>
        <label className="field">
          <span>Confirm new password</span>
          <input
            className="input"
            type="password"
            value={f.confirm}
            onChange={set("confirm")}
            autoComplete="new-password"
          />
        </label>
        <ul className="rules">
          {[
            ["length", "At least 8 characters"],
            ["mix", "A letter and a number"],
            ["match", "Passwords match"],
          ].map(([k, label]) => (
            <li key={k} className={rules[k] ? "ok" : ""}>
              {rules[k] ? <Check size={14} /> : <Circle size={14} />}
              {label}
            </li>
          ))}
        </ul>
        {error && <div className="form-error">{error}</div>}
        <button className="btn btn-primary" disabled={busy}>
          {busy ? "Updating..." : "Update password"}
        </button>
      </form>
    </div>
  );
}

function AppearanceTab() {
  const { theme, setTheme } = useTheme();
  return (
    <div className="card">
      <h2>Appearance</h2>
      <p className="muted small" style={{ marginBottom: 16 }}>
        Saved on this device.
      </p>
      <div className="theme-grid">
        {[
          ["dark", "Black", Moon],
          ["light", "White", Sun],
        ].map(([id, label, Icon]) => (
          <button
            key={id}
            className={`theme-card ${theme === id ? "on" : ""}`}
            onClick={() => setTheme(id)}
          >
            <div className={`preview ${id}`}>
              <i />
              <i />
              <i />
            </div>
            <span
              style={{
                display: "flex",
                gap: 8,
                alignItems: "center",
                fontWeight: 600,
              }}
            >
              <Icon size={16} />
              {label}
              {theme === id && (
                <Check
                  size={16}
                  style={{ marginLeft: "auto", color: "var(--accent)" }}
                />
              )}
            </span>
          </button>
        ))}
      </div>
    </div>
  );
}

const TABS = [
  ["profile", "Profile", User],
  ["security", "Security", KeyRound],
  ["appearance", "Appearance", Palette],
];

export default function Settings() {
  const [tab, setTab] = useState("profile");
  const { data, loading, error, reload } = useFetch(() => getProfile(), []);

  if (loading && !data) return <PageLoader text="Loading settings..." />;
  if (error && !data) return <ErrorMessage message={error} onRetry={reload} />;

  return (
    <>
      <div className="page-head">
        <div>
          <h1>Settings</h1>
          <p className="muted">Manage your account and preferences.</p>
        </div>
      </div>
      <div className="settings">
        <nav className="snav">
          {TABS.map(([id, label, Icon]) => (
            <button
              key={id}
              className={tab === id ? "on" : ""}
              onClick={() => setTab(id)}
            >
              <Icon size={17} />
              {label}
            </button>
          ))}
        </nav>
        <div>
          {tab === "profile" && <ProfileTab profile={data} reload={reload} />}
          {tab === "security" && <SecurityTab profile={data} />}
          {tab === "appearance" && <AppearanceTab />}
        </div>
      </div>
    </>
  );
}
