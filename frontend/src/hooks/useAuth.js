import { useContext } from "react";
import { AuthContext } from "../context/authCtx";
export const useAuth = () => useContext(AuthContext);
