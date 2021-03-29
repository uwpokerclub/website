import { verify } from "jsonwebtoken";
import { Request, Response, NextFunction } from "express";

import { CODES } from "../models/constants";

const SECRET = process.env.JWT_SECRET;

export default function requireAuthentication(
  req: Request,
  res: Response,
  next: NextFunction
): Response | undefined {
  const token =
    req.body.token ||
    req.query.token ||
    req.headers["x-access-token"] ||
    req.cookies.pctoken;
  if (token) {
    verify(token, SECRET as string, (err: unknown) => {
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
