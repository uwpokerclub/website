/* eslint-disable camelcase */
import { promises, rm } from "fs";
import fsPromises = promises;

import RouteHandler from "../lib/route_handler/RouteHandler";
import { CODES } from "../models/constants";
import { User } from "../types";
import { Router } from "express";
import { orderBy, Query, where } from "postgres-driver-service";

type CreateUserBody = {
  id: string;
  firstName: string;
  lastName: string;
  email: string;
  faculty: string;
  questId: string;
};

function validateCreateReq(body: CreateUserBody) {
  const { id, firstName, lastName, email, faculty, questId } = body;

  const stringValues = [id, firstName, lastName, email, faculty, questId];

  if (stringValues.some((v) => v === undefined)) {
    throw new Error(
      "Missing required fields: id, firstName, lastName, email, faculty, questId"
    );
  }

  if (stringValues.some((v) => typeof v !== "string")) {
    throw new Error(
      "Required fields have incorrect type: id, firstName, lastName, email, faculty, questId"
    );
  }

  if (stringValues.some((v) => v === "")) {
    throw new Error(
      "Required fields cannot be empty: id, firstName, lastName, email, faculty, questId"
    );
  }
}

type UpdateUserBody = {
  firstName: string;
  lastName: string;
  faculty: string;
};

function validateUpdateReq(body: UpdateUserBody) {
  const { firstName, lastName, faculty } = body;

  const stringValues = [firstName, lastName, faculty];

  if (stringValues.some((v) => v === undefined)) {
    throw new Error("Missing required fields: firstName, lastName, faculty");
  }

  if (stringValues.some((v) => typeof v !== "string")) {
    throw new Error(
      "Required fields have incorrect type: firstName, lastName, faculty"
    );
  }

  if (stringValues.some((v) => v === "")) {
    throw new Error(
      "Required fields cannot be empty: firstName, lastName, faculty"
    );
  }
}

function convertJSONToCSV(data: User[]) {
  if (data.length === 0) {
    return "";
  }
  let csvString = "";
  csvString += Object.keys(data[0]) + "\n";
  csvString += data.map(Object.values).join("\n");
  return csvString;
}

export default class UsersRouteHandler extends RouteHandler {
  handler(): Router {
    this.router.get("/", async (req, res, next) => {
      const client = await this.db.getConnection();
      const query = new Query("users", client);

      const mods = [orderBy("created_at", "DESC")];

      const users = await query.all<User>(mods).catch((err) => next(err));

      if (users === undefined) {
        client.release();

        return res.status(CODES.INTERNAL_SERVER_ERROR).json({
          error: "INTERNAL_ERROR",
          message: "An unknown error occured"
        });
      }

      client.release();

      return res.status(CODES.OK).json({
        users
      });
    });

    this.router.get("/:id", async (req, res, next) => {
      const { id } = req.params;

      const client = await this.db.getConnection();
      const query = new Query("users", client);

      const user = await query.find<User>("id", id).catch((err) => next(err));

      if (user === undefined) {
        client.release();

        return res.status(CODES.NOT_FOUND).json({
          error: "NOT_FOUND",
          message: "That user could not be found"
        });
      }

      client.release();

      return res.status(CODES.OK).json({
        user
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

      const {
        id,
        firstName,
        lastName,
        email,
        faculty,
        questId
      }: CreateUserBody = req.body;

      const client = await this.db.getConnection();
      const query = new Query("users", client);

      try {
        await query.insert<User>({
          id,
          first_name: firstName,
          last_name: lastName,
          email,
          faculty,
          quest_id: questId
        });
      } catch (err) {
        next(err);

        client.release();

        return res.status(CODES.INTERNAL_SERVER_ERROR).json({
          error: "DATABASE_ERROR",
          message: "An insertion error occurred"
        });
      }

      client.release();

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

      const client = await this.db.getConnection();
      const query = new Query("users", client);

      const { firstName, lastName, faculty }: UpdateUserBody = req.body;
      const { id } = req.params;

      try {
        await query.update<User>([where("id = ?", [id])], {
          first_name: firstName,
          last_name: lastName,
          faculty
        });
      } catch (err) {
        next(err);

        client.release();

        return res.status(CODES.INTERNAL_SERVER_ERROR).json({
          error: "DATABASE_ERROR",
          message: "An update error occurred"
        });
      }

      client.release();

      return res.status(CODES.OK).json({
        user: {
          id,
          firstName,
          lastName,
          faculty
        }
      });
    });

    this.router.delete("/:id", async (req, res, next) => {
      const { id } = req.params;

      const client = await this.db.getConnection();
      const query = new Query("users", client);

      try {
        await query.delete([where("id = ?", [id])]);
      } catch (err) {
        next(err);

        client.release();

        return res.status(CODES.INTERNAL_SERVER_ERROR).json({
          error: "DATABASE_ERROR",
          message: "An error occurred during deletion"
        });
      }

      client.release();

      return res.status(CODES.OK).end();
    });

    this.router.post("/export", async (req, res, next) => {
      const client = await this.db.getConnection();
      const query = new Query("users", client);

      const users = await query.all<User>([]).catch((err) => next(err));

      if (users === undefined) {
        client.release();

        return res.status(CODES.INTERNAL_SERVER_ERROR).json({
          error: "DATABASE_ERROR",
          message: "A lookup error occurred"
        });
      }

      client.release();

      const filename = "users.csv";

      const data = convertJSONToCSV(users);

      await fsPromises.writeFile(`tmp/${filename}`, data, "utf8");

      return res
        .status(CODES.OK)
        .download(`tmp/${filename}`, filename, (err) => {
          if (err) {
            res.status(CODES.INTERNAL_SERVER_ERROR).json({
              error: "INTERNAL_ERROR",
              message: "An error occurred during download"
            });

            next(err);
          } else {
            rm(`tmp/${filename}`, () => true);
          }
        });
    });

    return this.router;
  }
}
