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
          const eventQuery = new Query("events", client);

          const event = await eventQuery.find<Event>("id", eventId);

          if (event.state === EVENT_STATE.ENDED) {
            participants = await query.query(
              `SELECT users.first_name, users.last_name, users.id, entries.signed_out_at, entries.placement, entries.rebuys, entries.id AS membership_id
                FROM (SELECT memberships.id, memberships.user_id, participants.signed_out_at, participants.placement, participants.rebuys
                FROM participants INNER JOIN memberships ON memberships.id = participants.membership_id
                WHERE participants.event_id = $1) AS entries INNER JOIN users ON users.id = entries.user_id
                ORDER BY entries.placement ASC;`,
              [eventId]
            );
          } else {
            participants = await query.query(
              `SELECT users.first_name, users.last_name, users.id, entries.signed_out_at, entries.placement, entries.rebuys, entries.id AS membership_id
                FROM (SELECT memberships.id, memberships.user_id, participants.signed_out_at, participants.placement, participants.rebuys
                FROM participants INNER JOIN memberships ON memberships.id = participants.membership_id
                WHERE participants.event_id = $1) AS entries INNER JOIN users ON users.id = entries.user_id
                ORDER BY entries.signed_out_at DESC;`,
              [eventId]
            );
          }
        } else {
          participants = await query.query(
            `SELECT users.first_name, users.last_name, users.id, entries.signed_out_at, entries.placement, entries.rebuys, entries.id AS membership_id
            FROM (SELECT memberships.id, memberships.user_id, participants.signed_out_at, participants.placement, participants.rebuys
            FROM participants INNER JOIN memberships ON memberships.id = participants.membership_id
            WHERE participants.event_id = $1) AS entries INNER JOIN users ON users.id = entries.user_id;`,
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

      const signedInUsersIds = signedInUsers.map((p) => p.membership_id);
      const errors: { membershipId: string; message: string }[] = [];

      for (const p of participants) {
        if (signedInUsersIds.indexOf(p) !== NOT_FOUND) {
          errors.push({
            membershipId: p,
            message: "This user is already registered for this event"
          });

          participants.shift();
        } else {
          try {
            await query.insert<Entry>({
              membership_id: p,
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
      const { membershipId, eventId } = req.body;

      if (membershipId === undefined || membershipId === "") {
        return res.status(CODES.INVALID_REQUEST).json({
          error: "INVALID_REQUEST",
          message: "Required fields are missing: membershipId"
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
          [
            where("membership_id = ? AND event_id = ?", [membershipId, eventId])
          ],
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
        membershipId,
        eventId,
        signedOutAt: now
      });
    });

    this.router.post("/sign-in", async (req, res, next) => {
      const { membershipId, eventId } = req.body;

      if (membershipId === undefined || membershipId === "") {
        return res.status(CODES.INVALID_REQUEST).json({
          error: "INVALID_REQUEST",
          message: "Required fields are missing: membershipId"
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
          [
            where("membership_id = ? AND event_id = ?", [membershipId, eventId])
          ],
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
        membershipId,
        eventId,
        signedOutAt: null
      });
    });

    this.router.post("/rebuy", async (req, res, next) => {
      const { membershipId, eventId } = req.body;

      if (membershipId === undefined || membershipId === "") {
        return res.status(CODES.INVALID_REQUEST).json({
          error: "INVALID_REQUEST",
          message: "Required fields are missing: membershipId"
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

      const participants = await query
        .all<Entry>([
          where("membership_id = ? AND event_id = ?", [membershipId, eventId])
        ])
        .catch((err) => next(err));

      if (!participants) {
        client.release();

        return res.status(CODES.NOT_FOUND).json({
          err: "NOT_FOUND",
          message: "Participant not found"
        });
      }

      try {
        await query.update<Entry>(
          [
            where("membership_id = ? AND event_id = ?", [membershipId, eventId])
          ],
          { rebuys: participants[0].rebuys + 1 }
        );
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
        participant: {
          ...participants[0],
          rebuys: participants[0].rebuys + 1
        }
      });
    });

    this.router.delete("/", async (req, res, next) => {
      const { membershipId, eventId } = req.body;

      if (membershipId === undefined || membershipId === "") {
        return res.status(CODES.INVALID_REQUEST).json({
          error: "INVALID_REQUEST",
          message: "Required fields are missing: membershipId"
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
          where("membership_id = ? AND event_id = ?", [membershipId, eventId])
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
