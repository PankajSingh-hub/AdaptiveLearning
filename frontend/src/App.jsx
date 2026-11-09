import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import { AuthProvider } from "./context/AuthContext";
import { ToastProvider } from "./context/ToastContext";
import { useAuth } from "./hooks/useAuth";
import { homeFor } from "./utils/constants";
import ProtectedRoute from "./components/ProtectedRoute";
import Layout from "./components/Layout";
import Auth from "./pages/Auth";
import CandidateDashboard from "./pages/CandidateDashboard";
import AssessmentList from "./pages/AssessmentList";
import InterviewSession from "./pages/InterviewSession";
import InterviewReport from "./pages/InterviewReport";
import Performance from "./pages/Performance";
import AdminDashboard from "./pages/AdminDashboard";
import QuestionBank from "./pages/QuestionBank";
import AssessmentManagement from "./pages/AssessmentManagement";
import CandidateList from "./pages/CandidateList";
import CodingSession from "./pages/CodingSession";
import CodingResult from "./pages/CodingResult";
function Home() {
  const { user } = useAuth();
  return <Navigate to={homeFor(user.role)} replace />;
}

const only = (role, el) => <ProtectedRoute role={role}>{el}</ProtectedRoute>;

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <ToastProvider>
          <div className="aurora">
            <i />
            <i />
            <i />
          </div>
          <Routes>
            <Route path="/login" element={<Auth mode="login" />} />
            <Route path="/register" element={<Auth mode="register" />} />
            <Route
              path="/interview/:id"
              element={only("CANDIDATE", <InterviewSession />)}
            />
            <Route
              path="/coding/:id"
              element={only("CANDIDATE", <CodingSession />)}
            />

            <Route
              element={
                <ProtectedRoute>
                  <Layout />
                </ProtectedRoute>
              }
            >
              <Route path="/" element={<Home />} />
              <Route path="/report/:id" element={<InterviewReport />} />
              <Route
                path="/coding/:id/result"
                element={only("CANDIDATE", <CodingResult />)}
              />
              <Route
                path="/dashboard"
                element={only("CANDIDATE", <CandidateDashboard />)}
              />
              <Route
                path="/assessments"
                element={only("CANDIDATE", <AssessmentList />)}
              />
              <Route
                path="/performance"
                element={only("CANDIDATE", <Performance />)}
              />
              <Route
                path="/admin"
                element={only("ADMIN", <AdminDashboard />)}
              />
              <Route
                path="/admin/questions"
                element={only("ADMIN", <QuestionBank />)}
              />
              <Route
                path="/admin/assessments"
                element={only("ADMIN", <AssessmentManagement />)}
              />
              <Route
                path="/admin/candidates"
                element={only("ADMIN", <CandidateList />)}
              />
            </Route>
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </ToastProvider>
      </AuthProvider>
    </BrowserRouter>
  );
}
