import { useEffect, useState } from "react";
import { NavLink, Outlet, useLocation, useNavigate } from "react-router-dom";
import {
  ClipboardList,
  LayoutDashboard,
  Library,
  LogOut,
  Menu,
  Sparkles,
  TrendingUp,
  Users,
} from "lucide-react";
import { useAuth } from "../hooks/useAuth";
import { APP_NAME } from "../utils/constants";
import { Badge } from "./ui";

const CANDIDATE_NAV = [
  { to: "/dashboard", label: "Dashboard", icon: LayoutDashboard },
  { to: "/assessments", label: "Assessments", icon: ClipboardList },
  { to: "/performance", label: "Performance", icon: TrendingUp },
];
const ADMIN_NAV = [
  { to: "/admin", label: "Overview", icon: LayoutDashboard },
  { to: "/admin/questions", label: "Questions", icon: Library },
  { to: "/admin/assessments", label: "Assessments", icon: ClipboardList },
  { to: "/admin/candidates", label: "Candidates", icon: Users },
];

export default function Layout() {
  const { user, logout, isAdmin } = useAuth();
  const [open, setOpen] = useState(false);
  const navigate = useNavigate();
  const { pathname } = useLocation();
  useEffect(() => setOpen(false), [pathname]);

  const items = isAdmin ? ADMIN_NAV : CANDIDATE_NAV;
  const onLogout = () => {
    logout();
    navigate("/login", { replace: true });
  };

  return (
    <div className="shell">
      <aside className={`sidebar ${open ? "open" : ""}`}>
        <div className="brand">
          <Sparkles size={22} color="#e8b86a" />
          <span className="grad">{APP_NAME}</span>
        </div>
        <nav className="nav">
          {items.map(({ to, label, icon: Icon }) => (
            <NavLink
              key={to}
              to={to}
              end={to === "/admin" || to === "/dashboard"}
            >
              <Icon size={18} />
              {label}
            </NavLink>
          ))}
        </nav>
        <div className="bottom nav">
          <a
            href="#logout"
            onClick={(e) => {
              e.preventDefault();
              onLogout();
            }}
          >
            <LogOut size={18} />
            Log out
          </a>
        </div>
      </aside>
      {open && <div className="scrim" onClick={() => setOpen(false)} />}

      <div className="main">
        <header className="topbar">
          <button
            className="icon-btn menu-btn"
            onClick={() => setOpen(true)}
            aria-label="Menu"
          >
            <Menu size={20} />
          </button>
          <span className="spacer" />
          <div className="user-chip">
            <Badge tone="accent">{user.role}</Badge>
            <span className="who">{user.name}</span>
            <span className="avatar">{user.name?.[0]?.toUpperCase()}</span>
          </div>
        </header>
        <main className="content" key={pathname}>
          <Outlet />
        </main>
      </div>
    </div>
  );
}
