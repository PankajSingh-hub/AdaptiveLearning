import axios from "axios";

export const TOKEN_KEY = "aap_token";
export const USER_KEY = "aap_user";

// The ONLY place the API URL is configured.
const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || "http://localhost:8080/api",
});

api.interceptors.request.use((cfg) => {
  const token = localStorage.getItem(TOKEN_KEY);
  if (token) cfg.headers.Authorization = `Bearer ${token}`;
  return cfg;
});

api.interceptors.response.use(
  (res) => res,
  (err) => {
    const isLogin = err.config?.url?.includes("/auth/login");
    if (err.response?.status === 401 && !isLogin) {
      localStorage.removeItem(TOKEN_KEY);
      localStorage.removeItem(USER_KEY);
      window.location.href = "/login";
    }
    return Promise.reject(err);
  },
);

// Backend wraps everything as { success, message, data }.
export const unwrap = (promise) => promise.then((r) => r.data.data);

export const errMsg = (e) =>
  e.response?.data?.message ||
  (e.request
    ? "Cannot reach the server. Is the backend running?"
    : "Something went wrong");

export default api;
