import { useState } from "react";
import { Link, Navigate, useNavigate } from "react-router-dom";
import { Sparkles } from "lucide-react";
import { useAuth } from "../hooks/useAuth";
import useFetch from "../hooks/useFetch";
import { authConfig, login as loginRequest } from "../services/authService";
import { errMsg } from "../services/api";
import { APP_NAME, homeFor } from "../utils/constants";
import DemoButton from "../components/DemoButton";

export default function Auth({ mode }) {
  const isAdmin = mode === "admin";
  const { user, saveSession } = useAuth();
  const navigate = useNavigate();
  const { data: cfg } = useFetch(() => authConfig(), []);
  const [f, setF] = useState({ email: "", password: "" });
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  if (user) return <Navigate to={homeFor(user.role)} replace />;
  const set = (k) => (e) => setF({ ...f, [k]: e.target.value });

  const submit = async (e) => {
    e.preventDefault();
    setError("");
    if (!/^\S+@\S+\.\S+$/.test(f.email))
      return setError("Enter a valid email address");
    if (!f.password) return setError("Enter your password");
    setBusy(true);
    try {
      const res = await loginRequest(f.email.trim(), f.password);
      if (isAdmin && res.user.role !== "ADMIN") {
        setError("This account does not have admin access");
        return;
      }
      saveSession(res);
      navigate(homeFor(res.user.role), { replace: true });
    } catch (err) {
      setError(errMsg(err));
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="auth">
      <form className="card" onSubmit={submit}>
        <div className="brand" style={{ paddingLeft: 0 }}>
          <Sparkles size={24} color="var(--accent)" />
          <span className="grad">{APP_NAME}</span>
        </div>
        <h1 style={{ marginBottom: 4 }}>
          {isAdmin ? "Admin sign in" : "Welcome back"}
        </h1>
        <p className="muted" style={{ marginBottom: 24 }}>
          Adaptive technical assessments that match your level.
        </p>

        <label className="field">
          <span>Email</span>
          <input
            className="input"
            type="email"
            value={f.email}
            onChange={set("email")}
            autoComplete="email"
          />
        </label>
        <label className="field">
          <span>Password</span>
          <input
            className="input"
            type="password"
            value={f.password}
            onChange={set("password")}
            autoComplete="current-password"
          />
        </label>

        {error && <div className="form-error">{error}</div>}
        <button
          className="btn btn-primary"
          style={{ width: "100%" }}
          disabled={busy}
        >
          {busy ? "Please wait..." : "Sign in"}
        </button>

        {!isAdmin && (
          <>
            <p
              className="muted small"
              style={{ marginTop: 18, textAlign: "center" }}
            >
              New here?{" "}
              <Link to="/register" className="grad">
                Create an account
              </Link>
            </p>
            {cfg?.demoLogin && <DemoButton onError={setError} />}
          </>
        )}
      </form>
    </div>
  );
}
