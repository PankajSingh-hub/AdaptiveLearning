import { useContext } from "react";
import { ThemeContext } from "../context/themeCtx";

export const useTheme = () => useContext(ThemeContext);
