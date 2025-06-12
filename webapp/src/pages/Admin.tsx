import { Route, Routes } from "react-router-dom";
import { Home } from "./Home";
import { Users } from "./Users";
import { Semesters } from "./Semesters";
import { Rankings } from "./Rankings";
import { Events } from "./Events";
import AdminLayout from "@/layouts/AdminLayout";

export function Admin() {
  return (
    <AdminLayout>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/users/*" element={<Users />} />
        <Route path="/events/*" element={<Events />} />
        <Route path="/semesters/*" element={<Semesters />} />
        <Route path="/rankings/*" element={<Rankings />} />
      </Routes>
    </AdminLayout>
  );
}
