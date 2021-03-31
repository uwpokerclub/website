/* eslint-disable camelcase */
import { Router } from "express";
import { Query, where } from "postgres-driver-service";
import RouteHandler from "../lib/route_handler/RouteHandler";
import { CODES, EVENT_STATE } from "../models/constants";
import { Entry, Event } from "../types";

const NOT_FOUND = -1;

type ParticipantsBody = {
  eventId: string;
  participants: string[];
};

function validateCreateReq(body: ParticipantsBody) {
  const { eventId, participants } = body;

  if (eventId === undefined || eventId === "") {
    throw Error("Required fields cannot be empty: eventId");
  }

  if (participants.some((p) => p === undefined || p === "")) {
    throw Error(
      "Required fields cannot be empty: participants must contain valid users"
    );
  }
}

export default class ParticipantsRouteHandler extends RouteHandler {
  handler(): Router {
    this.router.get("/", async (req, res, next) => {
      const { eventId } = req.query;

      const client = await this.db.getConnection();
      const query = new Query("participants", client);
      let participants = [];

      try {
        if (eventId !== undefined && eventId !== "") {
          participants = await query.query(
            `SELECT participants.user_id as id, users.first_name, users.last_name,
                    participants.signed_out_at, participants.placement
                    FROM participants LEFT JOIN users ON users.id = participants.user_id
                    WHERE participants.event_id = $1 ORDER BY participants.placement ASC;`,
            [eventId]
          );
        } else {
          participants = await query.query(
            `SELECT participants.user_id as id, users.first_name, users.last_name,
                    participants.signed_out_at, participants.placement
                    FROM participants LEFT JOIN users ON users.id = participants.user_id
                    ORDER BY participants.placement ASC;`,
            []
          );
        }
      } catch (err) {
        next(err);

        return res.status(CODES.INTERNAL_SERVER_ERROR).json({
          error: "DATABASE_ERROR",
          message: "A lookup error occurred"
        });
      }

      client.release();

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

      const { eventId, participants }: ParticipantsBody = req.body;

      const client = await this.db.getConnection();
      let query = new Query("events", client);

      const event = await query
        .find<Event>("id", eventId)
        .catch((err) => next(err));

      if (event === undefined || event.state === EVENT_STATE.ENDED) {
        return res.status(CODES.FORBIDDEN).json({
          error: "FORBIDDEN",
          message: "You cannot perform this action"
        });
      }

      query = new Query("participants", client);
      const signedInUsers = await query
        .all<Entry>([where("event_id = ?", [eventId])])
        .catch((err) => next(err));
      if (signedInUsers === undefined) {
        return res.status(CODES.INTERNAL_SERVER_ERROR).json({
          error: "DATABASE_ERROR",
          message: "A lookup error occurred"
        });
      }

      const signedInUsersIds = signedInUsers.map((p) => p.user_id);
      const errors: { userId: string; message: string }[] = [];

      for (const p of participants) {
        if (signedInUsersIds.indexOf(p) !== NOT_FOUND) {
          errors.push({
            userId: p,
            message: "This user is already registered for this event"
          });

          participants.shift();
        } else {
          try {
            await query.insert<Entry>({
              user_id: p,
              event_id: eventId
            });
          } catch (err) {
            next(err);

            return res.status(CODES.INTERNAL_SERVER_ERROR).json({
              error: "DATABASE_ERROR",
              message: "A creation error occurred"
            });
          }
        }
      }

      client.release();

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
          message: "Required fields are missing: eventId"
        });
      }

      const client = await this.db.getConnection();
      let query = new Query("events", client);

      const now = new Date();

      const event = await query
        .find<Event>("id", eventId)
        .catch((err) => next(err));

      if (event === undefined || event.state === EVENT_STATE.ENDED) {
        return res.status(CODES.FORBIDDEN).json({
          error: "FORBIDDEN",
          message: "You cannot perform this action"
        });
      }

      query = new Query("participants", client);
      try {
        await query.update<Entry>(
          [where("user_id = ? AND event_id = ?", [userId, eventId])],
          { signed_out_at: now }
        );
      } catch (err) {
        next(err);

        return res.status(CODES.INTERNAL_SERVER_ERROR).json({
          error: "DATABASE_ERROR",
          message: "An update error occurred"
        });
      }

      client.release();

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
          message: "Required fields are missing: eventId"
        });
      }

      const client = await this.db.getConnection();
      let query = new Query("events", client);

      const event = await query
        .find<Event>("id", eventId)
        .catch((err) => next(err));

      if (event === undefined || event.state === EVENT_STATE.ENDED) {
        return res.status(CODES.FORBIDDEN).json({
          error: "FORBIDDEN",
          message: "You cannot perform this action"
        });
      }

      query = new Query("participants", client);
      try {
        await query.update<Entry>(
          [where("user_id = ? AND event_id = ?", [userId, eventId])],
          { signed_out_at: null }
        );
      } catch (err) {
        next(err);

        return res.status(CODES.INTERNAL_SERVER_ERROR).json({
          error: "DATABASE_ERROR",
          message: "An update error occurred"
        });
      }

      client.release();

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
          message: "Required fields are missing: eventId"
        });
      }

      const client = await this.db.getConnection();
      let query = new Query("events", client);

      const event = await query
        .find<Event>("id", eventId)
        .catch((err) => next(err));

      if (event === undefined || event.state === EVENT_STATE.ENDED) {
        return res.status(CODES.FORBIDDEN).json({
          error: "FORBIDDEN",
          message: "You cannot perform this action"
        });
      }

      query = new Query("participants", client);
      try {
        await query.delete([
          where("user_id = ? AND event_id = ?", [userId, eventId])
        ]);
      } catch (err) {
        next(err);

        return res.status(CODES.INTERNAL_SERVER_ERROR).json({
          error: "DATABASE_ERROR",
          message: "A deletion error occurred"
        });
      }

      client.release();

      return res.status(CODES.OK).end();
    });

    return this.router;
  }
}
