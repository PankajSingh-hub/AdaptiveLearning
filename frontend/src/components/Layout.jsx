import { useState } from "react";
import { NavLink, Outlet, useLocation, useNavigate } from "react-router-dom";
import {
  ClipboardList,
  LayoutDashboard,
  Library,
  Menu,
  Settings,
  Sparkles,
  TrendingUp,
  Users,
} from "lucide-react";
import { useAuth } from "../hooks/useAuth";
import { APP_NAME, homeFor } from "../utils/constants";
import ProfileMenu from "./ProfileMenu";

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
  const { user, isAdmin, logout } = useAuth();
  const navigate = useNavigate();
  const [open, setOpen] = useState(false);
  const { pathname } = useLocation();
  const items = isAdmin ? ADMIN_NAV : CANDIDATE_NAV;
  const close = () => setOpen(false);

  return (
    <div className="shell">
      <aside className={`sidebar ${open ? "open" : ""}`}>
        {/* A plain link on purpose: clicking the logo does a full page reload. */}
        <a
          className="brand"
          href={homeFor(user.role)}
          aria-label={`${APP_NAME} home`}
        >
          <Sparkles size={22} color="var(--accent)" />
          <span>{APP_NAME}</span>
        </a>
        <nav className="nav">
          {items.map(({ to, label, icon: Icon }) => (
            <NavLink
              key={to}
              to={to}
              end={to === "/admin" || to === "/dashboard"}
              onClick={close}
            >
              <Icon size={18} />
              {label}
            </NavLink>
          ))}
        </nav>
        <div className="bottom nav">
          <NavLink to="/settings" onClick={close}>
            <Settings size={18} />
            Settings
          </NavLink>
        </div>
      </aside>
      {open && <div className="scrim" onClick={close} />}

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
          <ProfileMenu />
        </header>
        <main className="content" key={pathname}>
          {user.isDemo && (
            <div className="demo-strip">
              Temporary demo account: nothing is saved after 24 hours.{" "}
              <button
                className="link-btn"
                onClick={() => {
                  logout();
                  navigate("/register");
                }}
              >
                Create a free account to keep your results
              </button>
            </div>
          )}
          <Outlet />
        </main>
      </div>
    </div>
  );
}
