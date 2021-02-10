const jwt = require("jsonwebtoken");

const { CODES } = require("../models/constants");

const SECRET = process.env.JWT_SECRET;

exports.requireAuthentication = (req, res, next) => {
  const token = req.body.token || req.query.token || req.headers["x-access-token"] || req.cookies.pctoken;
  if (token) {
    jwt.verify(token, SECRET, (err) => {
      if (err) {
        res.status(CODES.FORBIDDEN)
          .json({
            status: "Failed",
            message: "Bad token."
          });
      } else {
        next();
      }
    });
  } else {
    return res.status(CODES.FORBIDDEN)
      .json({
        status: "Failed",
        message: "Authentication required."
      });
  }
};
