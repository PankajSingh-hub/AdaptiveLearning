import { Navigate, useLocation } from "react-router-dom";
import { useAuth } from "../hooks/useAuth";
import { homeFor } from "../utils/constants";

// UX guard only. The real enforcement is the backend (JWT + role middleware).
export default function ProtectedRoute({ role, children }) {
  const { user } = useAuth();
  const location = useLocation();
  if (!user) return <Navigate to="/login" state={{ from: location }} replace />;
  if (role && user.role !== role)
    return <Navigate to={homeFor(user.role)} replace />;
  return children;
}
