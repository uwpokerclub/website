import { MouseEvent, useRef, useState } from "react";
import { ROLES } from "../../../data";

type CreateLoginFormProps = {
  onSubmit: (username: string, password: string, role: string) => Promise<void>;
};

/**
 * CreateLoginForm - Legacy form for admin user creation
 * Used by CreateLogin component (out of scope for redesign)
 */
export function CreateLoginForm({ onSubmit }: CreateLoginFormProps) {
  const usernameRef = useRef<HTMLInputElement | null>(null);
  const passwordRef = useRef<HTMLInputElement | null>(null);
  const roleRef = useRef<HTMLSelectElement | null>(null);
  const formRef = useRef<HTMLFormElement | null>(null);

  const [submitDisabled, setSubmitDisabled] = useState(false);

  const handleClick = async (e: MouseEvent<HTMLButtonElement>) => {
    e.preventDefault();
    setSubmitDisabled(true);
    await onSubmit(usernameRef.current!.value, passwordRef.current!.value, roleRef.current?.value || "");
    setSubmitDisabled(false);
    formRef.current!.reset();
  };

  return (
    <form ref={formRef} className="content-wrap">
      <div className="form-group">
        <label htmlFor="username">Username:</label>
        <input ref={usernameRef} type="text" name="username" className="form-control" />
      </div>

      <div className="form-group">
        <label htmlFor="password">Password:</label>
        <input ref={passwordRef} type="password" name="password" className="form-control" />
      </div>

      <div className="form-group">
        <label htmlFor="role">Role:</label>
        <select ref={roleRef} name="role" className="form-control">
          {Object.entries(ROLES).map(([key, value]) => (
            <option key={key} value={key}>
              {value}
            </option>
          ))}
        </select>
      </div>

      <button type="submit" className="btn btn-success" disabled={submitDisabled} onClick={(e) => handleClick(e)}>
        Create
      </button>
    </form>
  );
}
