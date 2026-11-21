import { useCallback, useMemo, useState } from "react";
import * as authService from "../services/authService";
import { TOKEN_KEY, USER_KEY } from "../services/api";
import { AuthContext } from "./authCtx";

const readUser = () => {
  try {
    return JSON.parse(localStorage.getItem(USER_KEY));
  } catch {
    return null;
  }
};

export function AuthProvider({ children }) {
  const [user, setUser] = useState(readUser);

  const saveSession = useCallback((res) => {
    localStorage.setItem(TOKEN_KEY, res.token);
    localStorage.setItem(USER_KEY, JSON.stringify(res.user));
    setUser(res.user);
    return res.user;
  }, []);

  const login = useCallback(
    async (email, password) =>
      saveSession(await authService.login(email, password)),
    [saveSession],
  );

  const logout = useCallback(() => {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
    setUser(null);
  }, []);

  // Merge changed fields (for example a new name) into the stored user.
  const updateUser = useCallback((patch) => {
    const next = { ...readUser(), ...patch };
    localStorage.setItem(USER_KEY, JSON.stringify(next));
    setUser(next);
  }, []);

  const value = useMemo(
    () => ({
      user,
      login,
      logout,
      saveSession,
      updateUser,
      isAdmin: user?.role === "ADMIN",
    }),
    [user, login, logout, saveSession, updateUser],
  );
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
