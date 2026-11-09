import api, { unwrap } from "./api";

export const listAssessments = () => unwrap(api.get("/assessments"));
export const startInterview = (assessmentId) =>
  unwrap(api.post("/interviews/start", { assessmentId }));
export const getInterview = (id) => unwrap(api.get(`/interviews/${id}`));
export const submitAnswer = (id, sessionQuestionId, optionId) =>
  unwrap(api.post(`/interviews/${id}/answer`, { sessionQuestionId, optionId }));
export const finishInterview = (id) =>
  unwrap(api.post(`/interviews/${id}/finish`));
export const listReports = (page = 0, size = 8) =>
  unwrap(api.get("/reports", { params: { page, size } }));
export const getReport = (id) => unwrap(api.get(`/reports/${id}`));
export const candidateAnalytics = (id) =>
  unwrap(api.get(`/analytics/candidate/${id}`));
