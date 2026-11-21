import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  Line,
  LineChart,
  Pie,
  PieChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import { useTheme } from "../hooks/useTheme";

const PALETTES = {
  dark: {
    accent: "#f5a524",
    good: "#5fd08a",
    bad: "#f06c5b",
    grid: "rgba(236,236,236,.08)",
    text: "#8f8f97",
    tipBg: "#202023",
    tipBorder: "#2b2b2f",
    tipText: "#ececec",
    cursor: "rgba(236,236,236,.05)",
  },
  light: {
    accent: "#b45309",
    good: "#15803d",
    bad: "#c0392b",
    grid: "rgba(28,28,30,.1)",
    text: "#6b6b73",
    tipBg: "#ffffff",
    tipBorder: "#dededa",
    tipText: "#1c1c1e",
    cursor: "rgba(28,28,30,.05)",
  },
};

function useChartStyle() {
  const { theme } = useTheme();
  const c = PALETTES[theme] || PALETTES.dark;
  return {
    c,
    tip: {
      contentStyle: {
        background: c.tipBg,
        border: `1px solid ${c.tipBorder}`,
        borderRadius: 10,
      },
      itemStyle: { color: c.tipText },
      labelStyle: { color: c.text },
    },
    axis: {
      tick: { fill: c.text, fontSize: 12 },
      axisLine: false,
      tickLine: false,
    },
  };
}

export function Donut({ correct, incorrect }) {
  const { c, tip } = useChartStyle();
  const data = [
    { name: "Correct", v: correct },
    { name: "Incorrect", v: incorrect },
  ];
  return (
    <ResponsiveContainer width="100%" height={200}>
      <PieChart>
        <Pie
          data={data}
          dataKey="v"
          innerRadius={58}
          outerRadius={82}
          paddingAngle={4}
          stroke="none"
        >
          <Cell fill={c.good} />
          <Cell fill={c.bad} />
        </Pie>
        <Tooltip {...tip} />
      </PieChart>
    </ResponsiveContainer>
  );
}

// rag = colour bars by accuracy (green >= 75, accent >= 60, red below)
export function HBars({
  data,
  nameKey = "topic",
  dataKey = "accuracy",
  unit = "%",
  rag = false,
  max,
}) {
  const { c, tip, axis } = useChartStyle();
  const color = (v) =>
    !rag ? c.accent : v >= 75 ? c.good : v >= 60 ? c.accent : c.bad;
  return (
    <ResponsiveContainer
      width="100%"
      height={Math.max(160, data.length * 40 + 20)}
    >
      <BarChart data={data} layout="vertical" margin={{ left: 0, right: 16 }}>
        <CartesianGrid horizontal={false} stroke={c.grid} />
        <XAxis type="number" domain={[0, max ?? "auto"]} {...axis} />
        <YAxis type="category" dataKey={nameKey} width={120} {...axis} />
        <Tooltip
          {...tip}
          cursor={{ fill: c.cursor }}
          formatter={(v) => `${v}${unit}`}
        />
        <Bar dataKey={dataKey} radius={[0, 8, 8, 0]} barSize={16}>
          {data.map((d, i) => (
            <Cell key={i} fill={color(d[dataKey])} />
          ))}
        </Bar>
      </BarChart>
    </ResponsiveContainer>
  );
}

export function Trend({ data }) {
  const { c, tip, axis } = useChartStyle();
  return (
    <ResponsiveContainer width="100%" height={220}>
      <LineChart data={data} margin={{ left: -16, right: 8, top: 8 }}>
        <CartesianGrid vertical={false} stroke={c.grid} />
        <XAxis dataKey="label" {...axis} />
        <YAxis domain={[0, 100]} {...axis} />
        <Tooltip {...tip} formatter={(v) => `${v}%`} />
        <Line
          type="monotone"
          dataKey="accuracy"
          stroke={c.accent}
          strokeWidth={3}
          dot={{ r: 4, fill: c.accent, stroke: "none" }}
        />
      </LineChart>
    </ResponsiveContainer>
  );
}

// Difficulty progression across the questions of one interview.
export function Steps({ levels }) {
  const { c, tip, axis } = useChartStyle();
  const num = { EASY: 1, MEDIUM: 2, HARD: 3 };
  const name = { 1: "Easy", 2: "Medium", 3: "Hard" };
  const data = levels.map((l, i) => ({ q: i + 1, level: num[l] }));
  return (
    <ResponsiveContainer width="100%" height={200}>
      <LineChart data={data} margin={{ left: -8, right: 8, top: 8 }}>
        <CartesianGrid vertical={false} stroke={c.grid} />
        <XAxis dataKey="q" {...axis} />
        <YAxis
          domain={[0.5, 3.5]}
          ticks={[1, 2, 3]}
          tickFormatter={(v) => name[v]}
          width={60}
          {...axis}
        />
        <Tooltip
          {...tip}
          formatter={(v) => name[v]}
          labelFormatter={(q) => `Question ${q}`}
        />
        <Line
          type="stepAfter"
          dataKey="level"
          stroke={c.accent}
          strokeWidth={3}
          dot={{ r: 4, fill: c.good, stroke: "none" }}
        />
      </LineChart>
    </ResponsiveContainer>
  );
}
