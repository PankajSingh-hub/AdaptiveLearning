import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Zap } from "lucide-react";
import { useAuth } from "../hooks/useAuth";
import { demoLogin } from "../services/authService";
import { errMsg } from "../services/api";
import { homeFor } from "../utils/constants";

export default function DemoButton({ onError }) {
  const { saveSession } = useAuth();
  const navigate = useNavigate();
  const [busy, setBusy] = useState(false);

  const go = async () => {
    if (busy) return;
    setBusy(true);
    try {
      const res = await demoLogin();
      saveSession(res);
      navigate(homeFor(res.user.role), { replace: true });
    } catch (e) {
      onError?.(errMsg(e));
      setBusy(false);
    }
  };

  return (
    <>
      <div className="divider">or</div>
      <button
        type="button"
        className="btn"
        style={{ width: "100%" }}
        disabled={busy}
        onClick={go}
      >
        <Zap size={16} />
        {busy ? "Opening demo..." : "Continue as demo candidate"}
      </button>
      <p className="muted small" style={{ marginTop: 8, textAlign: "center" }}>
        No email needed. The demo account is shared, so its results are visible
        to other demo visitors.
      </p>
    </>
  );
}
