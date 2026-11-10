import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { ArrowDownUp, Search } from "lucide-react";
import useFetch from "../hooks/useFetch";
import { getCandidate, listCandidates } from "../services/adminService";
import {
  Badge,
  Empty,
  ErrorMessage,
  Modal,
  PageLoader,
  Pager,
} from "../components/ui";
import { formatDate, formatLevel } from "../utils/format";
import { LEVEL_TONE } from "../utils/constants";

function CandidateDetail({ id, onClose }) {
  const { data, loading, error } = useFetch(() => getCandidate(id), [id]);
  return (
    <Modal title={data ? data.name : "Candidate"} onClose={onClose} wide>
      {loading ? (
        <PageLoader />
      ) : error ? (
        <ErrorMessage message={error} />
      ) : (
        <>
          <p className="muted small" style={{ marginBottom: 12 }}>
            {data.email} · joined {formatDate(data.joinedAt)}
          </p>
          {data.reports.length === 0 ? (
            <Empty title="No completed interviews" />
          ) : (
            data.reports.map((r) => (
              <Link key={r.id} to={`/report/${r.id}`} className="list-row">
                <div>
                  <div>{r.assessmentTitle}</div>
                  <div className="muted small">{formatDate(r.createdAt)}</div>
                </div>
                <div style={{ textAlign: "right" }}>
                  <strong>{r.accuracy}%</strong>{" "}
                  <Badge tone={LEVEL_TONE[r.level]}>
                    {formatLevel(r.level)}
                  </Badge>
                </div>
              </Link>
            ))
          )}
        </>
      )}
    </Modal>
  );
}

export default function CandidateList() {
  const [text, setText] = useState("");
  const [search, setSearch] = useState("");
  const [sort, setSort] = useState("joined");
  const [dir, setDir] = useState("desc");
  const [page, setPage] = useState(0);
  const [viewing, setViewing] = useState(null);

  useEffect(() => {
    const t = setTimeout(() => {
      setSearch(text);
      setPage(0);
    }, 350);
    return () => clearTimeout(t);
  }, [text]);

  const { data, loading, error, reload } = useFetch(
    () => listCandidates({ search, sort, dir, page, size: 10 }),
    [search, sort, dir, page],
  );

  return (
    <>
      <div className="page-head">
        <div>
          <h1>Candidates</h1>
          <p className="muted">
            {data ? `${data.totalItems} registered` : " "}
          </p>
        </div>
      </div>
      <div className="toolbar">
        <div style={{ position: "relative", flex: 2, minWidth: 180 }}>
          <Search
            size={16}
            style={{
              position: "absolute",
              left: 12,
              top: 14,
              color: "#8e9a90",
            }}
          />
          <input
            className="input"
            style={{ paddingLeft: 36 }}
            placeholder="Search by name or email..."
            value={text}
            onChange={(e) => setText(e.target.value)}
          />
        </div>
        <select
          className="input"
          value={sort}
          onChange={(e) => {
            setSort(e.target.value);
            setPage(0);
          }}
        >
          <option value="joined">Sort: Joined</option>
          <option value="name">Sort: Name</option>
          <option value="accuracy">Sort: Accuracy</option>
          <option value="interviews">Sort: Interviews</option>
        </select>
        <button
          className="btn"
          onClick={() => {
            setDir(dir === "asc" ? "desc" : "asc");
            setPage(0);
          }}
        >
          <ArrowDownUp size={16} />
          {dir === "asc" ? "Ascending" : "Descending"}
        </button>
      </div>

      <div className="card">
        {loading ? (
          <PageLoader text="Loading candidates..." />
        ) : error ? (
          <ErrorMessage message={error} onRetry={reload} />
        ) : data.items.length === 0 ? (
          <Empty title="No candidates found" />
        ) : (
          <>
            <div className="table-wrap">
              <table>
                <thead>
                  <tr>
                    <th>Name</th>
                    <th>Email</th>
                    <th>Interviews</th>
                    <th>Avg accuracy</th>
                    <th>Last interview</th>
                    <th />
                  </tr>
                </thead>
                <tbody>
                  {data.items.map((c) => (
                    <tr key={c.id}>
                      <td>{c.name}</td>
                      <td className="muted">{c.email}</td>
                      <td>{c.interviews}</td>
                      <td>{c.interviews ? `${c.avgAccuracy}%` : "—"}</td>
                      <td>{formatDate(c.lastInterview)}</td>
                      <td>
                        <button
                          className="btn"
                          onClick={() => setViewing(c.id)}
                        >
                          View
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            <Pager
              page={page}
              totalPages={data.totalPages}
              onChange={setPage}
            />
          </>
        )}
      </div>
      {viewing && (
        <CandidateDetail id={viewing} onClose={() => setViewing(null)} />
      )}
    </>
  );
}
