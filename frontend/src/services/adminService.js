import api, { unwrap } from "./api";

const clean = (params) =>
  Object.fromEntries(
    Object.entries(params).filter(([, v]) => v !== "" && v != null),
  );

export const adminAnalytics = () => unwrap(api.get("/analytics/admin"));

export const listQuestions = (params) =>
  unwrap(api.get("/questions", { params: clean(params) }));
export const createQuestion = (body) => unwrap(api.post("/questions", body));
export const updateQuestion = (id, body) =>
  unwrap(api.put(`/questions/${id}`, body));
// Returns the full response so the UI can show "deactivated instead of deleted".
export const deleteQuestion = (id) =>
  api.delete(`/questions/${id}`).then((r) => r.data);

export const listAllAssessments = () => unwrap(api.get("/assessments"));
export const createAssessment = (body) =>
  unwrap(api.post("/assessments", body));
export const updateAssessment = (id, body) =>
  unwrap(api.put(`/assessments/${id}`, body));
export const deleteAssessment = (id) =>
  api.delete(`/assessments/${id}`).then((r) => r.data);

export const listCandidates = (params) =>
  unwrap(api.get("/candidates", { params: clean(params) }));
export const getCandidate = (id) => unwrap(api.get(`/candidates/${id}`));
