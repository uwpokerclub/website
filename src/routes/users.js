/* eslint-disable camelcase */
const fs = require("fs");
const fsPromises = fs.promises;

const RouteHandler = require("../lib/route_handler/RouteHandler");
const { CODES } = require("../models/constants");

function validateCreateReq(body) {
  const { id, firstName, lastName, email, faculty, questId, paid, semesterId } = body;

  const stringValues = [id, firstName, lastName, email, faculty, questId, semesterId];

  if (stringValues.some((v) => v === undefined) || paid === undefined) {
    throw new Error("Missing required fields: id, firstName, lastName, email, faculty, questId, paid, semesterId");
  }

  if (stringValues.some((v) => typeof(v) !== "string") || typeof(paid) !== "boolean") {
    throw new Error("Required fields have incorrect type: id, firstName, lastName, email, faculty, questId, paid, semesterId");
  }

  if (stringValues.some((v) => v === "")) {
    throw new Error("Required fields cannot be empty: id, firstName, lastName, email, faculty, questId, semesterId");
  }
}

function validateUpdateReq(body) {
  const { firstName, lastName, faculty, paid, semesterId } = body;

  const stringValues = [firstName, lastName, faculty, semesterId];

  if (stringValues.some((v) => v === undefined) || paid === undefined) {
    throw new Error("Missing required fields: firstName, lastName, faculty, paid, semesterId");
  }

  if (stringValues.some((v) => typeof(v) !== "string") || typeof(paid) !== "boolean") {
    throw new Error("Required fields have incorrect type: firstName, lastName, faculty, paid, semesterId");
  }

  if (stringValues.some((v) => v === "")) {
    throw new Error("Required fields cannot be empty: firstName, lastName, faculty, semesterId");
  }
}

function convertJSONToCSV(data) {
  if (data.length === 0) {
    return "";
  }
  let csvString = "";
  csvString += Object.keys(data[0]) + "\n";
  csvString += data.map(Object.values).join("\n");
  return csvString;
}

class UsersRouteHandler extends RouteHandler {
  handler() {
    this.router.get("/", async (req, res, next) => {
      // only support semesterId filter for now
      const { semesterId } = req.query;

      let membersQuery = this.db.table("users").select();

      if (semesterId !== undefined && semesterId !== "") {
        membersQuery = membersQuery.where("semester_id = ?", semesterId);
      }

      const users = await membersQuery
        .orderBy("DESC", "created_at")
        .execute()
        .catch((err) => {
          res.status(CODES.INTERNAL_SERVER_ERROR).json({
            error: "INTERNAL_ERROR",
            message: "An unknown error occured"
          });

          next(err);
        });

      res.status(CODES.OK).json({
        users: users
      });
    });

    this.router.get("/:id", async (req, res, next) => {
      const { id } = req.params;

      const [ user ] = await this.db
        .table("users")
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

      if (user === undefined) {
        return res.status(CODES.NOT_FOUND).json({
          error: "NOT_FOUND",
          message: "That user could not be found"
        });
      }

      return res.status(CODES.OK).json({
        user: user
      });
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

      const { id, firstName, lastName, email, faculty, questId, paid, semesterId } = req.body;

      await this.db
        .table("users")
        .create({
          id,
          first_name: firstName,
          last_name: lastName,
          email,
          faculty,
          quest_id: questId,
          paid,
          semester_id: semesterId
        })
        .execute()
        .catch((err) => {
          res.status(CODES.INTERNAL_SERVER_ERROR).json({
            error: "DATABASE_ERROR",
            message: "An insertion error occurred"
          });

          next(err);
        });

      return res.status(CODES.CREATED).json({
        user: req.body
      });
    });

    this.router.patch("/:id", async (req, res, next) => {
      try {
        validateUpdateReq(req.body);
      } catch (err) {
        return res.status(CODES.INVALID_REQUEST).json({
          error: "INVALID_REQUEST",
          message: err.message
        });
      }

      const { firstName, lastName, faculty, paid, semesterId } = req.body;
      const { id } = req.params;

      await this.db
        .table("users")
        .update({
          first_name: firstName,
          last_name: lastName,
          faculty,
          paid,
          semester_id: semesterId
        })
        .where("id = ?", id)
        .execute()
        .catch((err) => {
          res.status(CODES.INTERNAL_SERVER_ERROR).json({
            error: "DATABASE_ERROR",
            message: "An update error occurred"
          });

          next(err);
        });

      return res.status(CODES.OK).json({
        user: {
          id,
          firstName,
          lastName,
          faculty,
          paid,
          semesterId
        }
      });
    });

    this.router.delete("/:id", async (req, res, next) => {
      const { id } = req.params;

      await this.db
        .table("users")
        .delete()
        .where("id = ?", id)
        .execute()
        .catch((err) => {
          res.status(CODES.INTERNAL_SERVER_ERROR).json({
            error: "DATABASE_ERROR",
            message: "An error occurred during deletion"
          });

          next(err);
        });

      return res.status(CODES.OK).end();
    });

    this.router.get("/export", async (req, res, next) => {
      const { semesterId } = req.query;

      let filename = "";
      let users = null;

      if (semesterId === undefined) {
        users = await this.db
          .table("users")
          .select()
          .execute()
          .catch((err) => {
            res.status(CODES.INTERNAL_SERVER_ERROR).json({
              error: "DATABASE_ERROR",
              message: "A lookup error occurred"
            });

            next(err);
          });

        filename = "users.csv";
      } else {
        users = await this.db
          .table("users")
          .select()
          .where("semester_id = ?", semesterId)
          .execute()
          .catch((err) => {
            res.status(CODES.INTERNAL_SERVER_ERROR).json({
              error: "DATABASE_ERROR",
              message: "A lookup error occurred"
            });

            next(err);
          });

        const [ semester ] = await this.db
          .table("semesters")
          .select()
          .where("id = ?", semesterId)
          .execute()
          .catch((err) => {
            res.status(CODES.INTERNAL_SERVER_ERROR).json({
              error: "DATABASE_ERROR",
              message: "A lookup error occurred"
            });

            next(err);
          });

        filename = `users_${semester.id}.csv`;
      }

      const data = convertJSONToCSV(users);

      await fsPromises.writeFile(`tmp/${filename}`, data, "utf8");

      return res.status(CODES.OK).download(`tmp/${filename}`, filename, (err) => {
        if (err) {
          res.status(CODES.INTERNAL_SERVER_ERROR).json({
            error: "INTERNAL_ERROR",
            message: "An error occurred during download"
          });

          next(err);
        } else {
          fs.rm(`tmp/${filename}`);
        }
      });
    });

    return this.router;
  }
}

module.exports = UsersRouteHandler;
