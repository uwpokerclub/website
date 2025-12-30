import { Route, Routes } from "react-router-dom";
import { Home } from "./Home";
import { Users } from "./Users";
import { Semesters } from "./Semesters";
import { Rankings } from "./Rankings";
import { Events } from "./Events";
import { Members } from "./Members";
import { Inventory } from "./Inventory";
import { Finances } from "./Finances";
import { Executive } from "./Executive";
import AdminLayout from "@/layouts/AdminLayout";

export function Admin() {
  return (
    <AdminLayout>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/users/*" element={<Users />} />
        <Route path="/members/*" element={<Members />} />
        <Route path="/events/*" element={<Events />} />
        <Route path="/semesters/*" element={<Semesters />} />
        <Route path="/rankings/*" element={<Rankings />} />
        <Route path="/inventory" element={<Inventory />} />
        <Route path="/finances" element={<Finances />} />
        <Route path="/executive" element={<Executive />} />
      </Routes>
    </AdminLayout>
  );
}
