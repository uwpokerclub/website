import { Router } from "express";
import { Query, where } from "postgres-driver-service";
import RouteHandler from "../lib/route_handler/RouteHandler";
import { CODES } from "../models/constants";
import { Membership, Semester } from "../types";

type CreateMembershipParams = {
  userId: string;
  semesterId: string;
  paid: boolean;
  discounted: boolean;
};

function validateCreateReq(body: CreateMembershipParams): void {
  const { semesterId, userId, paid, discounted } = body;

  const nonNullValues = [semesterId, userId, paid, discounted];

  if (nonNullValues.some((v) => v === undefined || v === null)) {
    throw new Error(
      "Missing required fields: semesterId, userId, paid, discounted"
    );
  }

  if (
    typeof semesterId !== "string" ||
    typeof userId !== "string" ||
    typeof paid !== "boolean" ||
    typeof discounted !== "boolean"
  ) {
    throw new Error(
      "Required fields have incorrect type: semesterId, userId, paid, discounted"
    );
  }
}

export default class MembershipsRouteHandler extends RouteHandler {
  handler(): Router {
    this.router.get("/", async (req, res, next) => {
      const { semesterId, userId } = req.query;

      // Check if neither semesterId or userId is in query
      if (
        (semesterId === undefined && userId === undefined) ||
        (semesterId === "" && userId === "")
      ) {
        return res.status(CODES.INVALID_REQUEST).json({
          error: "INVALID_REQUEST",
          message: "Request must include either semesterId or userId"
        });
      }

      // Check if both semesterId and userId are in query
      if (semesterId !== undefined && userId !== undefined) {
        return res.status(CODES.INVALID_REQUEST).json({
          error: "INVALID_REQUEST",
          message: "Request must include either semesterId or userId"
        });
      }

      const client = await this.db.getConnection();
      const query = new Query("memberships", client);
      let memberships: Membership[] = [];

      try {
        if (semesterId && typeof semesterId === "string") {
          memberships = await query.query(
            `SELECT memberships.id, users.id AS user_id, users.first_name, users.last_name, memberships.paid, memberships.discounted FROM users
          INNER JOIN memberships ON users.id = memberships.user_id
          WHERE memberships.semester_id = $1;`,
            [semesterId]
          );
        } else if (userId && typeof userId === "string") {
          memberships = await query.query(
            `SELECT memberships.id, semesters.id AS semester_id, semesters.name, memberships.paid, memberships.discounted FROM semesters
          INNER JOIN memberships ON semesters.id = memberships.semester_id
          WHERE memberships.user_id = $1;`,
            [userId]
          );
        }

        return res.status(CODES.OK).json({
          memberships
        });
      } catch (err) {
        next(err);

        return res.status(CODES.INTERNAL_SERVER_ERROR).json({
          error: "INTERNAL_ERROR",
          message: "A lookup error occurred"
        });
      } finally {
        client.release();
      }
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

      const { semesterId, userId, paid, discounted } = req.body;

      const client = await this.db.getConnection();
      const query = new Query("memberships", client);

      try {
        await query.insert<Membership>({
          semester_id: semesterId,
          user_id: userId,
          paid,
          discounted
        });

        // Pull semester and update the current budget if membership is paid.
        if (paid) {
          const semesterQuery = new Query("semesters", client);
          const semester = await semesterQuery.find<Semester>("id", semesterId);

          await semesterQuery.update([where("id = ?", [semesterId])], {
            current_budget: discounted
              ? Number(semester.current_budget) +
                semester.membership_discount_fee
              : Number(semester.current_budget) + semester.membership_fee
          });
        }

        return res.status(CODES.CREATED).json({
          membership: req.body
        });
      } catch (err) {
        next(err);

        return res.status(CODES.INTERNAL_SERVER_ERROR).json({
          error: "INTERNAL_ERROR",
          message: "An insertion error occurred"
        });
      } finally {
        client.release();
      }
    });

    this.router.get("/:id", async (req, res, next) => {
      const { id } = req.params;

      const client = await this.db.getConnection();
      const query = new Query("memberships", client);

      try {
        const membership = await query.find<Membership>("id", id);

        if (membership === undefined) {
          return res.status(CODES.NOT_FOUND).json({
            error: "NOT_FOUND",
            messsage: "Could not find membership"
          });
        }

        return res.status(CODES.OK).json({ membership });
      } catch (err) {
        next(err);

        return res.status(CODES.INTERNAL_SERVER_ERROR).json({
          error: "INTERNAL_ERROR",
          message: "A lookup error occurred"
        });
      } finally {
        client.release();
      }
    });

    this.router.patch("/:id", async (req, res, next) => {
      const { id } = req.params;

      // Check if body is present
      const { paid, discounted } = req.body;
      if (
        (paid === undefined || paid === null || typeof paid !== "boolean") &&
        (discounted === undefined ||
          discounted === null ||
          typeof discounted !== "boolean")
      ) {
        return res.status(CODES.INVALID_REQUEST).json({
          error: "INVALID_REQUEST",
          message: "Missing required field: paid, discounted"
        });
      }

      const client = await this.db.getConnection();
      const query = new Query("memberships", client);

      try {
        const membership = await query.find<Membership>("id", id);

        await query.update<Membership>([where("id = ?", [id])], {
          paid,
          discounted
        });

        // Update current semester budget if membership is being paid for
        if (paid && !membership.paid) {
          const semesterQuery = new Query("semesters", client);
          const semester = await semesterQuery.find<Semester>(
            "id",
            membership.semester_id
          );

          await semesterQuery.update(
            [where("id = ?", [membership.semester_id])],
            {
              current_budget: discounted
                ? Number(semester.current_budget) +
                  semester.membership_discount_fee
                : Number(semester.current_budget) + semester.membership_fee
            }
          );
        }

        return res.status(CODES.OK).end();
      } catch (err) {
        next(err);

        return res.status(CODES.INTERNAL_SERVER_ERROR).json({
          error: "INTERNAL_ERROR",
          message: "An update error occurred"
        });
      } finally {
        client.release();
      }
    });

    return this.router;
  }
}
