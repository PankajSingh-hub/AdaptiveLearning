export const APP_NAME = "AdaptIQ";
export const CATEGORIES = [
  "DSA",
  "Java",
  "OOP",
  "DBMS",
  "OS",
  "Computer Networks",
  "System Design",
  "SQL",
  "Programming",
];
export const DIFFICULTIES = ["EASY", "MEDIUM", "HARD"];
export const DIFF_TONE = { EASY: "good", MEDIUM: "warn", HARD: "bad" };
export const LEVEL_TONE = {
  STRONG: "good",
  AVERAGE: "warn",
  NEEDS_IMPROVEMENT: "bad",
  INCOMPLETE: "",
};
export const homeFor = (role) => (role === "ADMIN" ? "/admin" : "/dashboard");
