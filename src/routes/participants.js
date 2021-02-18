/* eslint-disable camelcase */
const RouteHandler = require("../lib/route_handler/RouteHandler");
const { CODES, EVENT_STATE } = require("../models/constants");

const NOT_FOUND = -1;

function validateCreateReq(body) {
  const { eventId, participants } = body;

  if (eventId === undefined || eventId === "") {
    throw Error("Required fields cannot be empty: eventId");
  }

  if (participants.some((p) => p === undefined || p === "")) {
    throw Error("Required fields cannot be empty: participants must contain valid users");
  }
}

class ParticipantsRouteHandler extends RouteHandler {
  handler() {
    this.router.get("/", async (req, res, next) => {
      const { eventId } = req.query;

      let participantsQuery = this.db.table("participants").select();
      if (eventId !== undefined && eventId !== "") {
        participantsQuery = participantsQuery.where("event_id = ?", eventId);
      }

      const participants = await participantsQuery.execute().catch((err) => {
        res.status(CODES.INTERNAL_SERVER_ERROR).json({
          error: "DATABASE_ERROR",
          message: "A lookup error occurred"
        });

        next(err);
      });

      return res.status(CODES.OK).json({
        participants
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

      const { eventId, participants } = req.body;

      const [ event ] = await this.db
        .table("events")
        .select()
        .where("id = ?", eventId)
        .execute()
        .catch((err) => {
          res.status(CODES.INTERNAL_SERVER_ERROR).json({
            error: "DATABASE_ERROR",
            message: "A lookup error occurred"
          });

          next(err);
        });

      if (event === undefined || event.state === EVENT_STATE.ENDED) {
        return res.status(CODES.FORBIDDEN).json({
          error: "FORBIDDEN",
          message: "You cannot perform this action"
        });
      }

      const signedInUsers = await this.db
        .table("participants")
        .select()
        .where("event_id = ?", eventId)
        .execute()
        .catch((err) => {
          res.status(CODES.INTERNAL_SERVER_ERROR).json({
            error: "DATABASE_ERROR",
            message: "A lookup error occurred"
          });

          next(err);
        });

      const signedInUsersIds = signedInUsers.map((p) => p.user_id);
      const errors = [];

      for (const p of participants) {
        if (signedInUsersIds.indexOf(p) !== NOT_FOUND) {
          errors.push({
            userId: p,
            message: "This user is already registered for this event"
          });

          participants.shift();
        } else {
          await this.db
            .table("participants")
            .create({
              user_id: p,
              event_id: eventId
            })
            .execute()
            .catch((err) => {
              res.status(CODES.INTERNAL_SERVER_ERROR).json({
                error: "DATABASE_ERROR",
                message: "A creation error occurred"
              });

              next(err);
            });
        }
      }

      return res.status(CODES.OK).json({
        signedIn: participants,
        errors
      });
    });

    this.router.post("/sign-out", async (req, res, next) => {
      const { userId, eventId } = req.body;

      if (userId === undefined || userId === "") {
        return res.status(CODES.INVALID_REQUEST).json({
          error: "INVALID_REQUEST",
          message: "Required fields are missing: userId"
        });
      } else if (eventId === undefined || eventId === "") {
        return res.status(CODES.INVALID_REQUEST).json({
          error: "INVALID_REQUEST",
          message: "Required fields are missing: userId"
        });
      }

      const now = new Date();

      const [ event ] = await this.db
        .table("events")
        .select()
        .where("id = ?", eventId)
        .execute()
        .catch((err) => {
          res.status(CODES.INTERNAL_SERVER_ERROR).json({
            error: "DATABASE_ERROR",
            message: "A lookup error occurred"
          });

          next(err);
        });

      if (event === undefined || event.state === EVENT_STATE.ENDED) {
        return res.status(CODES.FORBIDDEN).json({
          error: "FORBIDDEN",
          message: "You cannot perform this action"
        });
      }

      await this.db
        .table("participants")
        .update({ signed_out_at: now })
        .where("user_id = ? AND event_id = ?", userId, eventId)
        .execute()
        .catch((err) => {
          res.status(CODES.INTERNAL_SERVER_ERROR).json({
            error: "DATABASE_ERROR",
            message: "An update error occurred"
          });

          next(err);
        });

      return res.status(CODES.OK).json({
        userId,
        eventId,
        signedOutAt: now
      });
    });

    this.router.post("/sign-in", async (req, res, next) => {
      const { userId, eventId } = req.body;

      if (userId === undefined || userId === "") {
        return res.status(CODES.INVALID_REQUEST).json({
          error: "INVALID_REQUEST",
          message: "Required fields are missing: userId"
        });
      } else if (eventId === undefined || eventId === "") {
        return res.status(CODES.INVALID_REQUEST).json({
          error: "INVALID_REQUEST",
          message: "Required fields are missing: userId"
        });
      }

      const [ event ] = await this.db
        .table("events")
        .select()
        .where("id = ?", eventId)
        .execute()
        .catch((err) => {
          res.status(CODES.INTERNAL_SERVER_ERROR).json({
            error: "DATABASE_ERROR",
            message: "A lookup error occurred"
          });

          next(err);
        });

      if (event === undefined || event.state === EVENT_STATE.ENDED) {
        return res.status(CODES.FORBIDDEN).json({
          error: "FORBIDDEN",
          message: "You cannot perform this action"
        });
      }

      await this.db
        .table("participants")
        .update({ signed_out_at: null })
        .where("user_id = ? AND event_id = ?", userId, eventId)
        .execute()
        .catch((err) => {
          res.status(CODES.INTERNAL_SERVER_ERROR).json({
            error: "DATABASE_ERROR",
            message: "An update error occurred"
          });

          next(err);
        });

      return res.status(CODES.OK).json({
        userId,
        eventId,
        signedOutAt: null
      });
    });

    this.router.delete("/", async (req, res, next) => {
      const { userId, eventId } = req.body;

      if (userId === undefined || userId === "") {
        return res.status(CODES.INVALID_REQUEST).json({
          error: "INVALID_REQUEST",
          message: "Required fields are missing: userId"
        });
      } else if (eventId === undefined || eventId === "") {
        return res.status(CODES.INVALID_REQUEST).json({
          error: "INVALID_REQUEST",
          message: "Required fields are missing: userId"
        });
      }

      const [ event ] = await this.db
        .table("events")
        .select()
        .where("id = ?", eventId)
        .execute()
        .catch((err) => {
          res.status(CODES.INTERNAL_SERVER_ERROR).json({
            error: "DATABASE_ERROR",
            message: "A lookup error occurred"
          });

          next(err);
        });

      if (event === undefined || event.state === EVENT_STATE.ENDED) {
        return res.status(CODES.FORBIDDEN).json({
          error: "FORBIDDEN",
          message: "You cannot perform this action"
        });
      }

      await this.db
        .table("participants")
        .delete()
        .where("user_id = ? and event_id = ?", userId, eventId)
        .execute()
        .catch((err) => {
          res.status(CODES.INTERNAL_SERVER_ERROR).json({
            error: "DATABASE_ERROR",
            message: "A deletion error occurred"
          });

          next(err);
        });

      return res.status(CODES.OK).end();
    });

    return this.router;
  }
}

module.exports = ParticipantsRouteHandler;
