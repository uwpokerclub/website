import { verify } from "jsonwebtoken";

import { CODES } from "../models/constants";

const SECRET = process.env.JWT_SECRET;

export function requireAuthentication(req, res, next) {
  const token =
    req.body.token ||
    req.query.token ||
    req.headers["x-access-token"] ||
    req.cookies.pctoken;
  if (token) {
    verify(token, SECRET, (err) => {
      if (err) {
        res.status(CODES.FORBIDDEN).json({
          error: "TOKEN_ERROR",
          message: "Bad token"
        });
      } else {
        next();
      }
    });
  } else {
    return res.status(CODES.UNAUTHORIZED).json({
      error: "UNAUTHORIZED",
      message: "Authentication required"
    });
  }
}
