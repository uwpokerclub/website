/* eslint-disable camelcase */
import { Router } from "express";
import { orderBy, Query } from "postgres-driver-service";
import RouteHandler from "../lib/route_handler/RouteHandler";
import { CODES } from "../models/constants";
import { Semester } from "../types";

type SemesterBody = {
  name: string;
  startDate: string;
  endDate: string;
  meta: string;
};

function validateCreateReq(body: SemesterBody) {
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

export default class SemestersRouteHandler extends RouteHandler {
  handler(): Router {
    this.router.get("/", async (req, res, next) => {
      const client = await this.db.getConnection();
      const query = new Query("semesters", client);

      const semesters = await query
        .all([orderBy("start_date", "DESC")])
        .catch((err) => next(err));

      if (semesters === undefined) {
        return res.status(CODES.INTERNAL_SERVER_ERROR).json({
          error: "DATABASE_ERROR",
          message: "A lookup error occurred"
        });
      }

      client.release();

      return res.status(CODES.OK).json({ semesters });
    });

    this.router.get("/:id", async (req, res, next) => {
      const { id } = req.params;

      const client = await this.db.getConnection();
      const query = new Query("semesters", client);

      const semester = await query.find("id", id).catch((err) => next(err));

      if (semester === undefined) {
        return res.status(CODES.NOT_FOUND).json({
          error: "NOT_FOUND",
          message: "That semester could not be found"
        });
      }

      client.release();

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

      const client = await this.db.getConnection();
      const query = new Query("semesters", client);

      const { name, startDate, endDate, meta }: SemesterBody = req.body;

      try {
        await query.insert<Semester>({
          name,
          start_date: new Date(startDate),
          end_date: new Date(endDate),
          meta
        });
      } catch (err) {
        next(err);

        return res.status(CODES.INTERNAL_SERVER_ERROR).json({
          error: "DATABASE_ERROR",
          message: "An insertion error occurred"
        });
      }

      client.release();

      return res.status(CODES.CREATED).json({ semester: req.body });
    });

    this.router.get("/:id/rankings", async (req, res, next) => {
      const { id } = req.params;

      const client = await this.db.getConnection();
      const query = new Query("rankings", client);

      const rankings = await query
        .query(
          `select users.id, users.first_name, users.last_name, rankings.points from memberships m 
        inner join users on m.user_id = users.id 
        inner join rankings on m.id = rankings.membership_id 
        where m.semester_id = $1 order by rankings.points desc;`,
          [id]
        )
        .catch((err) => next(err));

      if (rankings === undefined) {
        return res.status(CODES.NOT_FOUND).json({
          error: "NOT_FOUND",
          message: "That semester could not be found"
        });
      }

      client.release();

      return res.status(CODES.OK).json({ rankings });
    });

    return this.router;
  }
}
