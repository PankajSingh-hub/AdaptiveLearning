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

const C = {
  accent: "#e8b86a",
  teal: "#8fc9a4",
  good: "#86d1a0",
  bad: "#e5846a",
  grid: "rgba(236,231,220,.08)",
  text: "#8e9a90",
};
const tip = {
  contentStyle: {
    background: "#17211b",
    border: "1px solid rgba(236,231,220,.1)",
    borderRadius: 10,
  },
  itemStyle: { color: "#ece7dc" },
  labelStyle: { color: "#8e9a90" },
};
const axis = {
  tick: { fill: C.text, fontSize: 12 },
  axisLine: false,
  tickLine: false,
};

export function Donut({ correct, incorrect }) {
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
          <Cell fill={C.good} />
          <Cell fill={C.bad} />
        </Pie>
        <Tooltip {...tip} />
      </PieChart>
    </ResponsiveContainer>
  );
}

// rag = colour bars by accuracy (teal >=75, blue >=60, red below)
export function HBars({
  data,
  nameKey = "topic",
  dataKey = "accuracy",
  unit = "%",
  rag = false,
  max,
}) {
  const color = (v) =>
    !rag ? C.accent : v >= 75 ? C.teal : v >= 60 ? C.accent : C.bad;
  return (
    <ResponsiveContainer
      width="100%"
      height={Math.max(160, data.length * 40 + 20)}
    >
      <BarChart data={data} layout="vertical" margin={{ left: 0, right: 16 }}>
        <CartesianGrid horizontal={false} stroke={C.grid} />
        <XAxis type="number" domain={[0, max ?? "auto"]} {...axis} />
        <YAxis type="category" dataKey={nameKey} width={120} {...axis} />
        <Tooltip
          {...tip}
          cursor={{ fill: "rgba(236,231,220,.04)" }}
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
  return (
    <ResponsiveContainer width="100%" height={220}>
      <LineChart data={data} margin={{ left: -16, right: 8, top: 8 }}>
        <CartesianGrid vertical={false} stroke={C.grid} />
        <XAxis dataKey="label" {...axis} />
        <YAxis domain={[0, 100]} {...axis} />
        <Tooltip {...tip} formatter={(v) => `${v}%`} />
        <Line
          type="monotone"
          dataKey="accuracy"
          stroke={C.accent}
          strokeWidth={3}
          dot={{ r: 4, fill: C.teal, stroke: "none" }}
        />
      </LineChart>
    </ResponsiveContainer>
  );
}

// Difficulty progression across the questions of one interview.
export function Steps({ levels }) {
  const num = { EASY: 1, MEDIUM: 2, HARD: 3 };
  const name = { 1: "Easy", 2: "Medium", 3: "Hard" };
  const data = levels.map((l, i) => ({ q: i + 1, level: num[l] }));
  return (
    <ResponsiveContainer width="100%" height={200}>
      <LineChart data={data} margin={{ left: -8, right: 8, top: 8 }}>
        <CartesianGrid vertical={false} stroke={C.grid} />
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
          stroke={C.teal}
          strokeWidth={3}
          dot={{ r: 4, fill: C.accent, stroke: "none" }}
        />
      </LineChart>
    </ResponsiveContainer>
  );
}
