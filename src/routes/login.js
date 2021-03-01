const bcrypt = require("bcrypt");
const jwt = require("jsonwebtoken");

const RouteHandler = require("../lib/route_handler/RouteHandler");
const { CODES } = require("../models/constants");

const SALT_ROUNDS = 10;
const DAY_IN_SECONDS = 86400;

function validateCreateRequest(username, password) {
  if (username === undefined || password === undefined) {
    throw new Error("Missing required fields: username, password");
  }

  if (typeof(username) !== "string" || typeof(password) !== "string") {
    throw new Error("Required fields have incorrect type: username, password");
  }

  if (username === "" || password === "") {
    throw new Error("Required fields cannot be empty: username, password");
  }
}

function validateSessionRequest(username, password) {
  if (username === undefined || password === undefined) {
    throw new Error("Missing required fields: username, password");
  }

  if (typeof(username) !== "string" || typeof(password) !== "string") {
    throw new Error("Required fields have incorrect type: username, password");
  }
}


class LoginRouteHandler extends RouteHandler {
  handler() {
    this.router.post("/", async (req, res, next) => {
      const { username, password } = req.body;

      try {
        validateCreateRequest(username, password);
      } catch (err) {
        return res.status(CODES.INVALID_REQUEST).json({
          error: "INVALID_REQUEST",
          message: err.message
        });
      }

      const loginCount = await this.db.table("logins").count();
      if (loginCount > 0) {
        return res.status(CODES.FORBIDDEN).json({
          error: "FORBIDDEN",
          message: "You cannot perform this action"
        });
      }

      const hash = await bcrypt.hash(password, SALT_ROUNDS).catch((err) => {
        res.status(CODES.INTERNAL_SERVER_ERROR).json({
          error: "INTERNAL_ERROR",
          message: "Error occurred creating the login"
        });

        next(err);
      });

      await this.db
        .table("logins")
        .create({
          username: req.body.username,
          password: hash
        })
        .execute()
        .catch((err) => {
          res.status(CODES.INTERNAL_SERVER_ERROR).json({
            error: "INTERNAL_ERROR",
            message: "Error occurred creating the login"
          });

          next(err);
        });

      return res.status(CODES.CREATED).end();
    });


    this.router.post("/session", async (req, res, next) => {
      const { username, password } = req.body;

      try {
        validateSessionRequest(username, password);
      } catch (err) {
        return res.status(CODES.INVALID_REQUEST).json({
          error: "INVALID_REQUEST",
          message: err.message
        });
      }

      const [ login ] = await this.db
        .table("logins")
        .select()
        .where("username = ?", username)
        .execute()
        .catch((err) => {
          res.status(CODES.INTERNAL_SERVER_ERROR).json({
            error: "INTERNAL_ERROR",
            message: "An database error occurred"
          });

          next(err);
        });

      if (login === undefined) {
        return res.status(CODES.UNAUTHORIZED).json({
          error: "UNAUTHORIZED",
          message: "Invalid username or password"
        });
      }

      const match = await bcrypt.compare(password, login.password).catch((err) => {
        res.status(CODES.INTERNAL_SERVER_ERROR).json({
          error: "INTERNAL_ERROR",
          message: "An unknown error occurred"
        });

        next(err);
      });

      if (!match) {
        return res.status(CODES.UNAUTHORIZED).json({
          error: "UNAUTHORIZED",
          message: "Invalid username or password"
        });
      }

      const token = jwt.sign(login, process.env.JWT_SECRET, { expiresIn: DAY_IN_SECONDS });

      return res
        .cookie("pctoken", token)
        .status(CODES.CREATED)
        .end();
    });

    return this.router;
  }
}


module.exports = LoginRouteHandler;
