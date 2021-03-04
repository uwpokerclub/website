/* eslint-disable camelcase */
import RouteHandler from "../lib/route_handler/RouteHandler";
import { CODES } from "../models/constants";

function validateCreateReq(body) {
  const { name, startDate, endDate, meta } = body;

  const nonNullValues = [name, startDate, endDate];

  if (nonNullValues.some((v) => v === undefined)) {
    throw new Error("Missing required fields: name, startDate, endDate");
  }

  if (typeof name !== "string") {
    throw new Error("Required fields have incorrect type: name");
  }

  if (meta !== undefined && typeof meta !== "string") {
    throw new Error("Optional fields have incorrect type: meta");
  }

  if (Number.isNaN(Date.parse(startDate))) {
    throw new Error("Required date fields cannot be parsed: startDate");
  }

  if (Number.isNaN(Date.parse(endDate))) {
    throw new Error("Required date fields cannot be parsed: endDate");
  }

  if (name === "") {
    throw new Error("Required fields cannot be empty: name");
  }
}

class SemestersRouteHandler extends RouteHandler {
  handler() {
    this.router.get("/", async (req, res, next) => {
      const semesters = await this.db
        .table("semesters")
        .select()
        .orderBy("DESC", "start_date")
        .execute()
        .catch((err) => {
          res.status(CODES.INTERNAL_SERVER_ERROR).json({
            error: "DATABASE_ERROR",
            message: "A lookup error occurred"
          });

          next(err);
        });

      return res.status(CODES.OK).json({ semesters });
    });

    this.router.get("/:id", async (req, res, next) => {
      const { id } = req.params;

      const [semester] = await this.db
        .table("semesters")
        .select()
        .where("id = ?", id)
        .execute()
        .catch((err) => {
          res.status(CODES.INTERNAL_SERVER_ERROR).json({
            error: "DATABASE_ERROR",
            message: "A lookup error occurred"
          });

          next(err);
        });

      if (semester === undefined) {
        return res.status(CODES.NOT_FOUND).json({
          error: "NOT_FOUND",
          message: "That semester could not be found"
        });
      }

      return res.status(CODES.OK).json({ semester });
    });

    this.router.post("/", async (req, res, next) => {
      try {
        validateCreateReq(req.body);
      } catch (err) {
        return res.status(CODES.INVALID_REQUEST).json({
          error: "INVALID_REQUEST",
          message: err.message
        });
      }

      const { name, startDate, endDate, meta } = req.body;

      await this.db
        .table("semesters")
        .create({
          name,
          start_date: startDate,
          end_date: endDate,
          meta: meta
        })
        .execute()
        .catch((err) => {
          res.status(CODES.INTERNAL_SERVER_ERROR).json({
            error: "DATABASE_ERROR",
            message: "An insertion error occurred"
          });

          next(err);
        });

      return res.status(CODES.CREATED).json({ semester: req.body });
    });

    this.router.get("/:id/rankings", async (req, res, next) => {
      const { id } = req.params;

      const rankings = await this.db
        .query(
          `SELECT users.id, users.first_name, users.last_name, rankings.points
                FROM rankings LEFT JOIN users ON users.id = rankings.user_id
                WHERE rankings.semester_id = $1 ORDER BY rankings.points DESC;`,
          id
        )
        .catch((err) => {
          res.status(CODES.INTERNAL_SERVER_ERROR).json({
            error: "DATABASE_ERROR",
            message: "A lookup error occurred"
          });

          next(err);
        });

      if (rankings === undefined) {
        return res.status(CODES.NOT_FOUND).json({
          error: "NOT_FOUND",
          message: "That semester could not be found"
        });
      }

      return res.status(CODES.OK).json({ rankings });
    });

    return this.router;
  }
}

export default SemestersRouteHandler;
