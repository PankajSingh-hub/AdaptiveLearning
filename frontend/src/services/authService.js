import api, { unwrap } from "./api";

export const login = (email, password) =>
  unwrap(api.post("/auth/login", { email, password }));
export const register = (name, email, password) =>
  unwrap(api.post("/auth/register", { name, email, password }));
export const me = () => unwrap(api.get("/auth/me"));
