import { useNavigate, useLocation } from "react-router-dom";
import { useAuth } from "@/hooks";
import { LoginForm } from "./LoginForm";

export function NewSession() {
  const navigate = useNavigate();
  const location = useLocation();
  const { login, error } = useAuth();

  const from = (location.state as { from: { pathname: string } })?.from?.pathname || "/admin";

  const handleLogin = async (username: string, password: string) => {
    await login(username, password, () => navigate(from, { replace: true }));
  };

  return (
    <div className="center">
      {error && (
        <div data-qa="login-error-banner" role="alert" className="alert alert-danger">
          <span>{error}</span>
        </div>
      )}

      <div className="row">
        <div className="col-md-4 col-lg-4 col-sm-3" />

        <div className="col-md-4 col-lg-4 col-sm-6">
          <h1>Login</h1>
          <LoginForm create={false} onSubmit={handleLogin} />
        </div>

        <div className="col-md-4 col-lg-4 col-sm-3" />
      </div>
    </div>
  );
}
