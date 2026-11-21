import api, { unwrap } from "./api";

const tz = () => {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone;
  } catch {
    return "UTC";
  }
};

export const getProfile = () =>
  unwrap(api.get("/me/profile", { params: { tz: tz() } }));
export const getStreak = () =>
  unwrap(api.get("/me/streak", { params: { tz: tz() } }));
export const updateProfile = (name) => unwrap(api.put("/me/profile", { name }));
export const changePassword = (currentPassword, newPassword, confirmPassword) =>
  unwrap(
    api.post("/me/password", { currentPassword, newPassword, confirmPassword }),
  );
