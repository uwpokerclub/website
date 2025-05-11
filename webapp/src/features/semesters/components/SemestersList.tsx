import { Link } from "react-router-dom";
import { useAuth, useFetch } from "../../../hooks";
import { Semester } from "../../../types";

export function SemestersList() {
  const { data: semesters } = useFetch<Semester[]>("semesters");
  const { hasPermission } = useAuth();

  // TODO: Need design for case where there are no semesters / error returned by API
  if (!semesters) {
    return <></>;
  }

  return (
    <div>
      <h1 data-qa="semesters-header">Semesters</h1>
      {hasPermission("create", "semester") && (
        <Link data-qa="create-semester-btn" to="new" className="btn btn-primary btn-responsive">
          Create a Semester
        </Link>
      )}
      <div className="table-responsive">
        <table className="table">
          <thead>
            <tr>
              <th>Name</th>
              <th>Start Date</th>
              <th>End Date</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {semesters.map((semester) => (
              <tr key={semester.id}>
                <td data-qa="semester-name">{semester.name}</td>

                <td>
                  {new Date(semester.startDate).toLocaleDateString("en-US", {
                    month: "long",
                    day: "numeric",
                    year: "numeric",
                  })}
                </td>

                <td>
                  {new Date(semester.endDate).toLocaleDateString("en-US", {
                    month: "long",
                    day: "numeric",
                    year: "numeric",
                  })}
                </td>

                <td>
                  {hasPermission("get", "semester") && (
                    <Link data-qa={`view-semester-${semester.id}`} to={`${semester.id}`} className="btn btn-primary">
                      View
                    </Link>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
