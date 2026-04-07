import { Navigate, Route, Routes } from "react-router-dom";
import { Dashboard } from "./Dashboard";
import { Rankings } from "./Rankings";
import { Events } from "./Events";
import { Members } from "./Members";
import { Inventory } from "./Inventory";
import { Finances } from "./Finances";
import { Executive } from "./Executive";
import { Logins } from "./Logins";
import AdminLayout from "@/layouts/AdminLayout";

export function Admin() {
  return (
    <AdminLayout>
      <Routes>
        <Route path="/" element={<Navigate to="/admin/dashboard" replace />} />
        <Route path="/dashboard" element={<Dashboard />} />
        <Route path="/members/*" element={<Members />} />
        <Route path="/events/*" element={<Events />} />
        <Route path="/rankings/*" element={<Rankings />} />
        <Route path="/inventory" element={<Inventory />} />
        <Route path="/finances" element={<Finances />} />
        <Route path="/executive" element={<Executive />} />
        <Route path="/logins/*" element={<Logins />} />
      </Routes>
    </AdminLayout>
  );
}
