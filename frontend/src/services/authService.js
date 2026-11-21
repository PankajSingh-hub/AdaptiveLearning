import api, { unwrap } from "./api";

export const login = (email, password) =>
  unwrap(api.post("/auth/login", { email, password }));
export const me = () => unwrap(api.get("/auth/me"));
export const authConfig = () => unwrap(api.get("/auth/config"));
export const demoLogin = () => unwrap(api.post("/auth/demo"));
export const signupStart = (email) =>
  unwrap(api.post("/auth/signup/start", { email }));
export const signupVerify = (email, code) =>
  unwrap(api.post("/auth/signup/verify", { email, code }));
export const signupComplete = (signupToken, name, password, confirmPassword) =>
  unwrap(
    api.post("/auth/signup/complete", {
      signupToken,
      name,
      password,
      confirmPassword,
    }),
  );
