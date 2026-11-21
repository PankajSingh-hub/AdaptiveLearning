import { useEffect, useState } from "react";
import { Link, Navigate, useNavigate } from "react-router-dom";
import { ArrowLeft, Check, Circle, Mail, Sparkles } from "lucide-react";
import { useAuth } from "../hooks/useAuth";
import useFetch from "../hooks/useFetch";
import { useToast } from "../context/toastCtx";
import {
  authConfig,
  signupComplete,
  signupStart,
  signupVerify,
} from "../services/authService";
import { errMsg } from "../services/api";
import { APP_NAME, homeFor } from "../utils/constants";
import OtpInput from "../components/OtpInput";
import DemoButton from "../components/DemoButton";

const STEPS = ["email", "code", "password"];
const EMAIL_RE = /^\S+@\S+\.\S+$/;

export default function Signup() {
  const { user, saveSession } = useAuth();
  const navigate = useNavigate();
  const toast = useToast();
  const { data: cfg } = useFetch(() => authConfig(), []);

  const [step, setStep] = useState("email");
  const [email, setEmail] = useState("");
  const [code, setCode] = useState("");
  const [token, setToken] = useState("");
  const [name, setName] = useState("");
  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [cooldown, setCooldown] = useState(0);

  useEffect(() => {
    if (cooldown <= 0) return;
    const t = setTimeout(() => setCooldown((c) => c - 1), 1000);
    return () => clearTimeout(t);
  }, [cooldown]);

  if (user) return <Navigate to={homeFor(user.role)} replace />;

  const requestCode = async () => {
    setError("");
    const addr = email.trim().toLowerCase();
    if (!EMAIL_RE.test(addr)) {
      setError("Enter a valid email address");
      return false;
    }
    setBusy(true);
    try {
      const r = await signupStart(addr);
      setEmail(addr);
      setCode("");
      setCooldown(r.resendAfterSec);
      return true;
    } catch (err) {
      setError(errMsg(err));
      return false;
    } finally {
      setBusy(false);
    }
  };

  const onEmailSubmit = async (e) => {
    e.preventDefault();
    if (busy) return;
    if (await requestCode()) setStep("code");
  };

  const resend = async () => {
    if (busy || cooldown > 0) return;
    if (await requestCode()) toast.success("A new code is on its way");
  };

  const verify = async (value) => {
    if (busy) return;
    setBusy(true);
    setError("");
    try {
      const r = await signupVerify(email, value);
      setToken(r.signupToken);
      setStep("password");
    } catch (err) {
      setError(errMsg(err));
      setCode("");
    } finally {
      setBusy(false);
    }
  };

  const rules = {
    length: password.length >= 8,
    mix: /[A-Za-z]/.test(password) && /\d/.test(password),
    match: password.length > 0 && password === confirm,
  };

  const finish = async (e) => {
    e.preventDefault();
    if (busy) return;
    setError("");
    if (name.trim().length < 2) return setError("Please enter your name");
    if (!rules.length || !rules.mix)
      return setError(
        "Password must be 8+ characters with a letter and a number",
      );
    if (!rules.match) return setError("Passwords do not match");
    setBusy(true);
    try {
      const res = await signupComplete(token, name.trim(), password, confirm);
      saveSession(res);
      toast.success("Account created. Welcome!");
      navigate(homeFor(res.user.role), { replace: true });
    } catch (err) {
      setError(errMsg(err));
    } finally {
      setBusy(false);
    }
  };

  const startOver = () => {
    setStep("email");
    setCode("");
    setToken("");
    setPassword("");
    setConfirm("");
    setError("");
  };
  const idx = STEPS.indexOf(step);

  return (
    <div className="auth">
      <div className="card">
        <div className="brand" style={{ paddingLeft: 0 }}>
          <Sparkles size={24} color="var(--accent)" />
          <span className="grad">{APP_NAME}</span>
        </div>
        <div className="steps">
          {STEPS.map((s, i) => (
            <i key={s} className={i === idx ? "on" : i < idx ? "done" : ""} />
          ))}
        </div>

        <div key={step} className="step-pane">
          {step === "email" && (
            <form onSubmit={onEmailSubmit}>
              <h1 style={{ marginBottom: 4 }}>Create your account</h1>
              <p className="muted" style={{ marginBottom: 22 }}>
                We will email you a 6-digit code to verify your address.
              </p>
              <label className="field">
                <span>Email</span>
                <input
                  className="input"
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  autoComplete="email"
                  autoFocus
                />
              </label>
              {error && <div className="form-error">{error}</div>}
              <button
                className="btn btn-primary"
                style={{ width: "100%" }}
                disabled={busy}
              >
                <Mail size={16} />
                {busy ? "Sending code..." : "Send verification code"}
              </button>
              <p
                className="muted small"
                style={{ marginTop: 18, textAlign: "center" }}
              >
                Already have an account?{" "}
                <Link to="/login" className="grad">
                  Sign in
                </Link>
              </p>
              {cfg?.demoLogin && <DemoButton onError={setError} />}
            </form>
          )}

          {step === "code" && (
            <div>
              <h1 style={{ marginBottom: 4 }}>Check your email</h1>
              <p className="muted" style={{ marginBottom: 18 }}>
                Enter the 6-digit code sent to{" "}
                <strong style={{ color: "var(--text)" }}>{email}</strong>. It
                expires in 10 minutes.
              </p>
              <OtpInput
                value={code}
                onChange={setCode}
                onComplete={verify}
                disabled={busy}
              />
              {cfg?.emailMode === "console" && (
                <p
                  className="small"
                  style={{
                    color: "var(--warn)",
                    marginBottom: 12,
                    textAlign: "center",
                  }}
                >
                  Email is not configured on this server, so no email was sent.
                  The code is printed in the backend terminal.
                </p>
              )}
              {error && (
                <div className="form-error" style={{ textAlign: "center" }}>
                  {error}
                </div>
              )}
              <button
                className="btn btn-primary"
                style={{ width: "100%" }}
                disabled={busy || code.length !== 6}
                onClick={() => verify(code)}
              >
                {busy ? "Verifying..." : "Verify email"}
              </button>
              <div
                className="small muted"
                style={{
                  marginTop: 16,
                  display: "flex",
                  justifyContent: "space-between",
                }}
              >
                <button type="button" className="link-btn" onClick={startOver}>
                  <ArrowLeft size={13} /> Change email
                </button>
                <button
                  type="button"
                  className="link-btn"
                  disabled={busy || cooldown > 0}
                  onClick={resend}
                >
                  {cooldown > 0 ? `Resend in ${cooldown}s` : "Resend code"}
                </button>
              </div>
            </div>
          )}

          {step === "password" && (
            <form onSubmit={finish}>
              <h1 style={{ marginBottom: 4 }}>Almost there</h1>
              <p className="muted" style={{ marginBottom: 20 }}>
                Your email is verified. Choose a password to finish.
              </p>
              <label className="field">
                <span>Name</span>
                <input
                  className="input"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  autoComplete="name"
                  autoFocus
                />
              </label>
              <label className="field">
                <span>Password</span>
                <input
                  className="input"
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  autoComplete="new-password"
                />
              </label>
              <label className="field">
                <span>Confirm password</span>
                <input
                  className="input"
                  type="password"
                  value={confirm}
                  onChange={(e) => setConfirm(e.target.value)}
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
              <button
                className="btn btn-primary"
                style={{ width: "100%" }}
                disabled={busy}
              >
                {busy ? "Creating account..." : "Create account"}
              </button>
              <p
                className="small muted"
                style={{ marginTop: 14, textAlign: "center" }}
              >
                <button type="button" className="link-btn" onClick={startOver}>
                  Start over
                </button>
              </p>
            </form>
          )}
        </div>
      </div>
    </div>
  );
}
