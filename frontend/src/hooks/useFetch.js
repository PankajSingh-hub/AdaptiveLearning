import { useCallback, useEffect, useRef, useState } from "react";
import { errMsg } from "../services/api";

// Loads data from an API call and tracks loading/error. A response only counts if it is for the
// current dependency key, so stale responses (a slow search request, say) are ignored.
export default function useFetch(fn, deps = []) {
  const [state, setState] = useState({ key: null, data: null, error: "" });
  const [version, setVersion] = useState(0);
  const fnRef = useRef(fn);
  useEffect(() => {
    fnRef.current = fn;
  });

  const key = `${JSON.stringify(deps)}:${version}`;

  useEffect(() => {
    let cancelled = false;
    fnRef
      .current()
      .then((data) => {
        if (!cancelled) setState({ key, data, error: "" });
      })
      .catch((e) => {
        if (!cancelled)
          setState((s) => ({ key, data: s.data, error: errMsg(e) }));
      });
    return () => {
      cancelled = true;
    };
  }, [key]);

  const reload = useCallback(() => setVersion((v) => v + 1), []);
  const current = state.key === key;
  return {
    data: state.data,
    loading: !current,
    error: current ? state.error : "",
    reload,
  };
}
