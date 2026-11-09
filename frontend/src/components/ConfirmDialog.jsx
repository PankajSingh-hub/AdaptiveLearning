import { Modal } from "./ui";

export default function ConfirmDialog({
  title,
  message,
  confirmText = "Confirm",
  busy,
  onConfirm,
  onCancel,
}) {
  return (
    <Modal title={title} onClose={onCancel}>
      <p className="muted" style={{ marginBottom: 20 }}>
        {message}
      </p>
      <div style={{ display: "flex", gap: 10, justifyContent: "flex-end" }}>
        <button className="btn" onClick={onCancel} disabled={busy}>
          Cancel
        </button>
        <button className="btn btn-primary" onClick={onConfirm} disabled={busy}>
          {busy ? "Working..." : confirmText}
        </button>
      </div>
    </Modal>
  );
}
