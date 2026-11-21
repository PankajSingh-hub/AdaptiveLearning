import { useRef } from "react";

const LENGTH = 6;

// Six boxes that behave like one field: auto-advance, backspace, arrow keys and paste.
export default function OtpInput({ value, onChange, onComplete, disabled }) {
  const refs = useRef([]);
  const digits = Array.from({ length: LENGTH }, (_, i) => value[i] || "");
  const focus = (i) => refs.current[i]?.focus();

  const update = (next) => {
    onChange(next);
    if (next.length === LENGTH) onComplete?.(next);
  };

  const onInput = (i, e) => {
    const typed = e.target.value.replace(/\D/g, "");
    if (!typed) return;
    const next = (value.slice(0, i) + typed).slice(0, LENGTH); // also handles autofill of all 6 digits
    update(next);
    focus(Math.min(next.length, LENGTH - 1));
  };

  const onKeyDown = (i, e) => {
    if (e.key === "Backspace") {
      e.preventDefault();
      if (digits[i]) update(value.slice(0, i) + value.slice(i + 1));
      else if (i > 0) {
        update(value.slice(0, i - 1) + value.slice(i));
        focus(i - 1);
      }
    } else if (e.key === "ArrowLeft" && i > 0) focus(i - 1);
    else if (e.key === "ArrowRight" && i < LENGTH - 1) focus(i + 1);
  };

  const onPaste = (e) => {
    const pasted = e.clipboardData
      .getData("text")
      .replace(/\D/g, "")
      .slice(0, LENGTH);
    if (!pasted) return;
    e.preventDefault();
    update(pasted);
    focus(Math.min(pasted.length, LENGTH - 1));
  };

  return (
    <div className="otp" onPaste={onPaste}>
      {digits.map((d, i) => (
        <input
          key={i}
          ref={(el) => {
            refs.current[i] = el;
          }}
          className={d ? "filled" : ""}
          value={d}
          disabled={disabled}
          inputMode="numeric"
          autoComplete={i === 0 ? "one-time-code" : "off"}
          aria-label={`Digit ${i + 1}`}
          autoFocus={i === 0}
          onFocus={(e) => e.target.select()}
          onChange={(e) => onInput(i, e)}
          onKeyDown={(e) => onKeyDown(i, e)}
        />
      ))}
    </div>
  );
}
