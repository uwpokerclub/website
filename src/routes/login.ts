import { hash, compare } from "bcrypt";
import { Router } from "express";
import { sign } from "jsonwebtoken";
import { Query } from "postgres-driver-service";

import RouteHandler from "../lib/route_handler/RouteHandler";
import { CODES } from "../models/constants";
import { Login } from "../types";

const SALT_ROUNDS = 10;
const DAY_IN_SECONDS = 86400;

type LoginBody = {
  username: string;
  password: string;
};

function validateCreateRequest(username: string, password: string): void {
  if (username === undefined || password === undefined) {
    throw new Error("Missing required fields: username, password");
  }

  if (typeof username !== "string" || typeof password !== "string") {
    throw new Error("Required fields have incorrect type: username, password");
  }

  if (username === "" || password === "") {
    throw new Error("Required fields cannot be empty: username, password");
  }
}

function validateSessionRequest(username: string, password: string): void {
  if (username === undefined || password === undefined) {
    throw new Error("Missing required fields: username, password");
  }

  if (typeof username !== "string" || typeof password !== "string") {
    throw new Error("Required fields have incorrect type: username, password");
  }
}

export default class LoginRouteHandler extends RouteHandler {
  handler(): Router {
    this.router.post("/", async (req, res, next) => {
      const { username, password }: LoginBody = req.body;

      try {
        validateCreateRequest(username, password);
      } catch (err) {
        return res.status(CODES.INVALID_REQUEST).json({
          error: "INVALID_REQUEST",
          message: err.message
        });
      }

      const client = await this.db.getConnection();
      const query = new Query("logins", client);

      const loginCount = await query.count([]).catch((err) => next(err));
      if (loginCount === undefined) {
        client.release();

        return res.status(CODES.INTERNAL_SERVER_ERROR).json({
          error: "INTERAL_SERVER_ERROR",
          message: "A lookup error occurred"
        });
      }

      if (loginCount > 0) {
        client.release();

        return res.status(CODES.FORBIDDEN).json({
          error: "FORBIDDEN",
          message: "You cannot perform this action"
        });
      }

      const passwordHash = await hash(password, SALT_ROUNDS).catch((err) =>
        next(err)
      );
      if (passwordHash === undefined) {
        client.release();

        return res.status(CODES.INTERNAL_SERVER_ERROR).json({
          error: "INTERNAL_ERROR",
          message: "Error occurred creating the login"
        });
      }

      try {
        await query.insert({
          username: req.body.username,
          password: passwordHash
        });
      } catch (err) {
        next(err);

        client.release();

        return res.status(CODES.INTERNAL_SERVER_ERROR).json({
          error: "INTERNAL_ERROR",
          message: "Error occurred creating the login"
        });
      }

      client.release();

      return res.status(CODES.CREATED).end();
    });

    this.router.post("/session", async (req, res, next) => {
      const { username, password }: LoginBody = req.body;

      try {
        validateSessionRequest(username, password);
      } catch (err) {
        return res.status(CODES.INVALID_REQUEST).json({
          error: "INVALID_REQUEST",
          message: err.message
        });
      }

      const client = await this.db.getConnection();
      const query = new Query("logins", client);

      const login = await query
        .find<Login>("username", username)
        .catch((err) => next(err));

      if (login === undefined) {
        client.release();

        return res.status(CODES.UNAUTHORIZED).json({
          error: "UNAUTHORIZED",
          message: "Invalid username or password"
        });
      }

      const match = await compare(password, login.password).catch((err) =>
        next(err)
      );
      if (match === undefined) {
        client.release();

        return res.status(CODES.INTERNAL_SERVER_ERROR).json({
          error: "INTERNAL_ERROR",
          message: "An unknown error occurred"
        });
      }

      if (!match) {
        client.release();

        return res.status(CODES.UNAUTHORIZED).json({
          error: "UNAUTHORIZED",
          message: "Invalid username or password"
        });
      }

      const token = sign(login, process.env.JWT_SECRET as string, {
        expiresIn: DAY_IN_SECONDS
      });

      client.release();

      return res.cookie("pctoken", token).status(CODES.CREATED).end();
    });

    return this.router;
  }
}
