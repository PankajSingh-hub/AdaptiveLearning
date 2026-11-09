import { useCallback, useEffect, useRef, useState } from "react";
import { errMsg } from "../services/api";

// Loads data from an API call and tracks loading/error. Ignores stale responses
// (e.g. while the user is still typing in a search box).
export default function useFetch(fn, deps = []) {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const seq = useRef(0);

  // eslint-disable-next-line react-hooks/exhaustive-deps
  const load = useCallback(() => {
    const id = ++seq.current;
    setLoading(true);
    setError("");
    return fn()
      .then((d) => id === seq.current && setData(d))
      .catch((e) => id === seq.current && setError(errMsg(e)))
      .finally(() => id === seq.current && setLoading(false));
  }, deps);

  useEffect(() => {
    load();
  }, [load]);
  return { data, loading, error, reload: load };
}
