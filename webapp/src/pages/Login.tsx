import { Route, Routes } from "react-router-dom";
import { LoginPage } from "../features/auth";

export function Login() {
  return (
    <Routes>
      <Route path="/" element={<LoginPage />} />
    </Routes>
  );
}
