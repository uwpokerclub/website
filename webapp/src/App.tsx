import { BrowserRouter as Router, Route, Routes } from "react-router-dom";
import { ToastProvider } from "@uwpokerclub/components";
import { Activate, Admin, Index, Login } from "./pages";
import { AuthProvider, RequireAuth, SemesterProvider } from "@/components";

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
                    <Admin />
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
