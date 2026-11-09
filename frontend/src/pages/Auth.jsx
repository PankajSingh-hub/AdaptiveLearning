import { useState } from "react";
import { Link, Navigate, useNavigate } from "react-router-dom";
import { Sparkles } from "lucide-react";
import { useAuth } from "../hooks/useAuth";
import { register } from "../services/authService";
import { errMsg } from "../services/api";
import { APP_NAME, homeFor } from "../utils/constants";

export default function Auth({ mode }) {
  const isRegister = mode === "register";
  const { user, login } = useAuth();
  const navigate = useNavigate();
  const [f, setF] = useState({ name: "", email: "", password: "" });
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  if (user) return <Navigate to={homeFor(user.role)} replace />;
  const set = (k) => (e) => setF({ ...f, [k]: e.target.value });

  const submit = async (e) => {
    e.preventDefault();
    setError("");
    if (isRegister && f.name.trim().length < 2)
      return setError("Please enter your name");
    if (!/^\S+@\S+\.\S+$/.test(f.email))
      return setError("Enter a valid email address");
    if (isRegister && f.password.length < 8)
      return setError("Password must be at least 8 characters");
    if (!f.password) return setError("Enter your password");
    setBusy(true);
    try {
      if (isRegister) await register(f.name.trim(), f.email.trim(), f.password);
      const u = await login(f.email.trim(), f.password);
      navigate(homeFor(u.role), { replace: true });
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
          <Sparkles size={24} color="#6d8dff" />
          <span className="grad">{APP_NAME}</span>
        </div>
        <h1 style={{ marginBottom: 4 }}>
          {isRegister ? "Create your account" : "Welcome back"}
        </h1>
        <p className="muted" style={{ marginBottom: 24 }}>
          Adaptive technical assessments that match your level.
        </p>

        {isRegister && (
          <label className="field">
            <span>Name</span>
            <input
              className="input"
              value={f.name}
              onChange={set("name")}
              autoComplete="name"
            />
          </label>
        )}
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
            autoComplete={isRegister ? "new-password" : "current-password"}
          />
        </label>

        {error && <div className="form-error">{error}</div>}
        <button
          className="btn btn-primary"
          style={{ width: "100%" }}
          disabled={busy}
        >
          {busy ? "Please wait..." : isRegister ? "Create account" : "Sign in"}
        </button>

        <p
          className="muted small"
          style={{ marginTop: 18, textAlign: "center" }}
        >
          {isRegister ? "Already have an account? " : "New here? "}
          <Link to={isRegister ? "/login" : "/register"} className="grad">
            {isRegister ? "Sign in" : "Create an account"}
          </Link>
        </p>

        {!isRegister && import.meta.env.DEV && (
          <div
            className="chips"
            style={{ marginTop: 16, justifyContent: "center" }}
          >
            <button
              type="button"
              className="btn small"
              onClick={() =>
                setF({
                  ...f,
                  email: "candidate@example.com",
                  password: "Candidate@123",
                })
              }
            >
              Fill demo candidate
            </button>
            <button
              type="button"
              className="btn small"
              onClick={() =>
                setF({
                  ...f,
                  email: "admin@example.com",
                  password: "Admin@12345",
                })
              }
            >
              Fill demo admin
            </button>
          </div>
        )}
      </form>
    </div>
  );
}
