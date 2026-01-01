import { useState, useCallback } from "react";
import { Link } from "react-router-dom";
import { useAuth, useFetch } from "../../../hooks";
import { Semester } from "../../../types";
import { CreateSemesterModal } from "./CreateSemesterModal";

export function SemestersList() {
  const { data: semesters, setData: setSemesters } = useFetch<Semester[]>("semesters");
  const { hasPermission } = useAuth();
  const [isModalOpen, setIsModalOpen] = useState(false);

  const handleCreateClick = () => {
    setIsModalOpen(true);
  };

  const handleModalClose = () => {
    setIsModalOpen(false);
  };

  const handleCreateSuccess = useCallback(
    (newSemester: Semester) => {
      setIsModalOpen(false);
      // Add new semester to the list (at the beginning since it's newest)
      setSemesters((prev) => (prev ? [newSemester, ...prev] : [newSemester]));
    },
    [setSemesters],
  );

  // TODO: Need design for case where there are no semesters / error returned by API
  if (!semesters) {
    return <></>;
  }

  return (
    <div>
      <h1 data-qa="semesters-header">Semesters</h1>
      {hasPermission("create", "semester") && (
        <button
          data-qa="create-semester-btn"
          onClick={handleCreateClick}
          className="btn btn-primary btn-responsive"
          type="button"
        >
          Create a Semester
        </button>
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

      {hasPermission("create", "semester") && (
        <CreateSemesterModal isOpen={isModalOpen} onClose={handleModalClose} onSuccess={handleCreateSuccess} />
      )}
    </div>
  );
}
