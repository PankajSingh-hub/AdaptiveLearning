import { AuthContext } from "./authCtx";
import { useCallback, useMemo, useState } from "react";
import * as authService from "../services/authService";
import { TOKEN_KEY, USER_KEY } from "../services/api";

const readUser = () => {
  try {
    return JSON.parse(localStorage.getItem(USER_KEY));
  } catch {
    return null;
  }
};

export function AuthProvider({ children }) {
  const [user, setUser] = useState(readUser);

  const login = useCallback(async (email, password) => {
    const res = await authService.login(email, password);
    localStorage.setItem(TOKEN_KEY, res.token);
    localStorage.setItem(USER_KEY, JSON.stringify(res.user));
    setUser(res.user);
    return res.user;
  }, []);

  const logout = useCallback(() => {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
    setUser(null);
  }, []);

  const value = useMemo(
    () => ({ user, login, logout, isAdmin: user?.role === "ADMIN" }),
    [user, login, logout],
  );
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
