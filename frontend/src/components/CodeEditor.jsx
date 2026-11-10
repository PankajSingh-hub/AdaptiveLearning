import { useLayoutEffect, useRef } from "react";

const UNIT = "    ";
const PAIRS = { "(": ")", "[": "]", "{": "}", '"': '"', "'": "'" };
const CLOSERS = new Set([")", "]", "}", '"', "'"]);
const OPENERS = new Set(["{", "(", "["]);

// A plain textarea with a line-number gutter and basic editor behaviour:
// auto-indent on Enter, Tab / Shift+Tab, auto-closing brackets and quotes.
export default function CodeEditor({ value, onChange, disabled, language }) {
  const gutter = useRef(null);
  const area = useRef(null);
  const caret = useRef(null);
  const lines = value.split("\n").length;

  // Restore the caret after React re-renders the controlled textarea.
  useLayoutEffect(() => {
    if (caret.current && area.current) {
      area.current.setSelectionRange(caret.current[0], caret.current[1]);
      caret.current = null;
    }
  }, [value]);

  const apply = (next, start, end = start) => {
    if (next === value) {
      area.current.setSelectionRange(start, end);
      return;
    }
    caret.current = [start, end];
    onChange(next);
  };

  const handleEnter = (s, en) => {
    const lineStart = value.lastIndexOf("\n", s - 1) + 1;
    const line = value.slice(lineStart, s);
    const indent = line.match(/^[ \t]*/)[0];
    const trimmed = line.trimEnd();
    const last = trimmed[trimmed.length - 1];
    const opens =
      language === "python"
        ? last === ":" || OPENERS.has(last)
        : OPENERS.has(last);
    const extra = opens ? UNIT : "";
    const after = value.slice(en);
    const between = OPENERS.has(last) && after[0] === PAIRS[last];
    const head = value.slice(0, s);

    if (between) {
      const mid = `\n${indent}${extra}`;
      apply(head + mid + `\n${indent}` + after, s + mid.length);
    } else {
      const ins = `\n${indent}${extra}`;
      apply(head + ins + after, s + ins.length);
    }
  };

  const handleTab = (s, en, shift) => {
    const multi = value.slice(s, en).includes("\n");
    if (!shift && !multi) {
      apply(value.slice(0, s) + UNIT + value.slice(en), s + UNIT.length);
      return;
    }
    const endRef = en > s && value[en - 1] === "\n" ? en - 1 : en;
    const blockStart = value.lastIndexOf("\n", s - 1) + 1;
    let blockEnd = value.indexOf("\n", endRef);
    if (blockEnd < 0) blockEnd = value.length;
    const rows = value.slice(blockStart, blockEnd).split("\n");

    if (!shift) {
      const next =
        value.slice(0, blockStart) +
        rows.map((r) => UNIT + r).join("\n") +
        value.slice(blockEnd);
      apply(next, s + UNIT.length, en + UNIT.length * rows.length);
      return;
    }
    let removedFirst = 0;
    let removedTotal = 0;
    const out = rows.map((r, i) => {
      const m = r.match(/^( {1,4}|\t)/);
      const n = m ? m[0].length : 0;
      if (i === 0) removedFirst = n;
      removedTotal += n;
      return r.slice(n);
    });
    const next =
      value.slice(0, blockStart) + out.join("\n") + value.slice(blockEnd);
    const newS = Math.max(blockStart, s - removedFirst);
    apply(next, newS, Math.max(newS, en - removedTotal));
  };

  const onKeyDown = (e) => {
    if (e.nativeEvent.isComposing) return;
    const t = e.target;
    const s = t.selectionStart;
    const en = t.selectionEnd;
    const k = e.key;
    const plain = !e.ctrlKey && !e.metaKey && !e.altKey;

    if (k === "Enter" && plain && !e.shiftKey) {
      e.preventDefault();
      return handleEnter(s, en);
    }
    if (k === "Tab" && plain) {
      e.preventDefault();
      return handleTab(s, en, e.shiftKey);
    }
    if (!plain) return;

    if (
      k === "Backspace" &&
      s === en &&
      s > 0 &&
      PAIRS[value[s - 1]] === value[s]
    ) {
      e.preventDefault();
      return apply(value.slice(0, s - 1) + value.slice(s + 1), s - 1);
    }
    if (CLOSERS.has(k) && s === en && value[s] === k) {
      // type over an existing closer
      e.preventDefault();
      return apply(value, s + 1);
    }
    if (k === "}" && s === en) {
      // outdent a closing brace on a whitespace-only line
      const lineStart = value.lastIndexOf("\n", s - 1) + 1;
      const before = value.slice(lineStart, s);
      if (before.length >= UNIT.length && /^ +$/.test(before)) {
        e.preventDefault();
        return apply(
          value.slice(0, s - UNIT.length) + "}" + value.slice(en),
          s - UNIT.length + 1,
        );
      }
    }
    if (PAIRS[k]) {
      const close = PAIRS[k];
      if (s !== en) {
        // wrap the selection
        e.preventDefault();
        return apply(
          value.slice(0, s) + k + value.slice(s, en) + close + value.slice(en),
          s + 1,
          en + 1,
        );
      }
      const prev = value[s - 1] || "";
      const next = value[s] || "";
      const isQuote = k === '"' || k === "'";
      const okNext = next === "" || /[\s)\]}]/.test(next);
      if (okNext && !(isQuote && /\w/.test(prev))) {
        e.preventDefault();
        apply(value.slice(0, s) + k + close + value.slice(en), s + 1);
      }
    }
  };

  return (
    <div className="editor">
      <pre className="gutter" ref={gutter}>
        {Array.from({ length: lines }, (_, i) => i + 1).join("\n")}
      </pre>
      <textarea
        ref={area}
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
