import { useRef } from "react";

// A plain textarea with a line-number gutter and Tab support (no external editor library).
export default function CodeEditor({ value, onChange, disabled }) {
  const gutter = useRef(null);
  const lines = value.split("\n").length;

  const onKeyDown = (e) => {
    if (e.key !== "Tab") return;
    e.preventDefault();
    const t = e.target;
    const start = t.selectionStart;
    onChange(value.slice(0, start) + "    " + value.slice(t.selectionEnd));
    requestAnimationFrame(() => {
      t.selectionStart = t.selectionEnd = start + 4;
    });
  };

  return (
    <div className="editor">
      <pre className="gutter" ref={gutter}>
        {Array.from({ length: lines }, (_, i) => i + 1).join("\n")}
      </pre>
      <textarea
        value={value}
        disabled={disabled}
        spellCheck={false}
        autoCapitalize="off"
        autoCorrect="off"
        wrap="off"
        aria-label="Code editor"
        onChange={(e) => onChange(e.target.value)}
        onKeyDown={onKeyDown}
        onScroll={(e) => {
          if (gutter.current) gutter.current.scrollTop = e.target.scrollTop;
        }}
      />
    </div>
  );
}
