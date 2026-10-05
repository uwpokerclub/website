import { lazy, Suspense } from "react";
import { BrowserRouter as Router, Route, Routes } from "react-router-dom";
import { ToastProvider } from "@uwpokerclub/components";
import { Activate } from "./pages/Activate";
import { Index } from "./pages/Index";
import { Login } from "./pages/Login";
import { LoadingScreen } from "@/components/LoadingScreen";
import { AuthProvider, RequireAuth, SemesterProvider } from "@/components";
import { ErrorBoundary } from "./components/ErrorBoundary";

const Admin = lazy(() => import("./pages/Admin").then(({ Admin: page }) => ({ default: page })));

function App() {
  return (
    <AuthProvider>
      <ToastProvider>
        <Router>
          <Routes>
            <Route path="/*" element={<Index />} />
            <Route path="/activate" element={<Activate />} />
            <Route path="/admin/login/*" element={<Login />} />
            <Route
              path="/admin/*"
              element={
                <RequireAuth>
                  <SemesterProvider>
                    <Suspense fallback={<LoadingScreen />}>
                      <ErrorBoundary recoveryAction={{ label: "Reload app", onClick: () => window.location.reload() }}>
                        <Admin />
                      </ErrorBoundary>
                    </Suspense>
                  </SemesterProvider>
                </RequireAuth>
              }
            />
          </Routes>
        </Router>
      </ToastProvider>
    </AuthProvider>
  );
}

export default App;
