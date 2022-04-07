/* eslint-disable camelcase */
import { Router } from "express";
import { Query, where } from "postgres-driver-service";
import RouteHandler from "../lib/route_handler/RouteHandler";
import PointsService from "../lib/services/points_service/PointsService";
import { CODES, EVENT_STATE } from "../models/constants";
import { Entry, Event, EventState, Ranking } from "../types";

type EventParams = {
  name: string;
  format: string;
  notes: string;
  startDate: string;
  semesterId: string;
};

function validateCreateReq(body: EventParams): void {
  const { name, startDate, format, notes, semesterId } = body;

  const nonNullValues = [name, startDate, format, semesterId];

  if (nonNullValues.some((v) => v === undefined)) {
    throw new Error(
      "Missing required fields: name, startDate, format, semesterId"
    );
  }

  if (
    typeof name !== "string" ||
    typeof format !== "string" ||
    typeof semesterId !== "string"
  ) {
    throw new Error(
      "Required fields have incorrect type: name, format, semesterId"
    );
  }

  if (notes !== undefined && typeof notes !== "string") {
    throw new Error("Optional fields have incorrect type: notes");
  }

  if (Number.isNaN(Date.parse(startDate))) {
    throw new Error("Required date fields cannot be parsed: startDate");
  }

  if (name === "" || format === "" || semesterId === "") {
    throw new Error(
      "Required fields cannot be empty: name, format, semesterId"
    );
  }
}

async function updateRankings(
  pQuery: Query,
  rQuery: Query,
  entry: Entry,
  points: number,
  placement: number
): Promise<void> {
  await pQuery.update<Entry>(
    [
      where("membership_id = ? AND event_id = ?", [
        entry.membership_id,
        entry.event_id
      ])
    ],
    {
      placement
    }
  );

  const [ranking] = await rQuery.all<Ranking>([
    where("membership_id = ?", [entry.membership_id])
  ]);

  if (ranking === undefined) {
    await rQuery.insert<Ranking>({
      membership_id: entry.membership_id,
      points
    });
  } else {
    await rQuery.update([where("membership_id = ?", [entry.membership_id])], {
      points: ranking.points + points
    });
  }
}

export default class EventsRouteHandler extends RouteHandler {
  handler(): Router {
    this.router.get("/", async (req, res, next) => {
      const { semesterId } = req.query;

      const client = await this.db.getConnection();
      const query = new Query("events", client);
      let events: Event[];

      try {
        if (semesterId !== undefined) {
          events = await query.query<Event>(
            `SELECT * FROM events LEFT JOIN
          (SELECT event_id, COUNT(*) FROM participants GROUP BY event_id)
          AS entries ON events.id = entries.event_id WHERE events.semester_id = $1
          ORDER BY events.start_date DESC;`,
            [semesterId]
          );
        } else {
          events = await query.query<Event>(
            `SELECT * FROM events LEFT JOIN
          (SELECT event_id, COUNT(*) FROM participants GROUP BY event_id)
          AS entries ON events.id = entries.event_id ORDER BY events.start_date DESC;`,
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

      return res.status(CODES.OK).json({ events });
    });

    this.router.get("/:id", async (req, res, next) => {
      const { id } = req.params;

      const client = await this.db.getConnection();
      const query = new Query("events", client);

      try {
        const event = await query.find<Event>("id", id);

        if (event === undefined) {
          return res.status(CODES.NOT_FOUND).json({
            error: "NOT_FOUND",
            message: "That event could not be found"
          });
        }

        return res.status(CODES.OK).json({ event });
      } catch (err) {
        next(err);

        return res.status(CODES.INTERNAL_SERVER_ERROR).json({
          error: "DATABASE_ERROR",
          message: "A lookup error occurred"
        });
      } finally {
        client.release();
      }
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
        name,
        startDate,
        format,
        notes,
        semesterId
      }: EventParams = req.body;

      const client = await this.db.getConnection();
      const query = new Query("events", client);

      try {
        await query.insert<Event>({
          name,
          start_date: new Date(startDate),
          format,
          notes,
          semester_id: semesterId
        });
      } catch (err) {
        next(err);

        return res.status(CODES.INTERNAL_SERVER_ERROR).json({
          error: "DATABASE_ERROR",
          message: "An insertion error occurred"
        });
      }

      client.release();

      return res.status(CODES.CREATED).json({
        event: req.body
      });
    });

    this.router.post("/:id/end", async (req, res, next) => {
      const { id } = req.params;

      const client = await this.db.getConnection();
      let query = new Query("events", client);

      const event = await query.find<Event>("id", id).catch((err) => next(err));

      if (event === undefined) {
        return res.status(CODES.FORBIDDEN).json({
          error: "FORBIDDEN",
          message: "You cannot perform this action"
        });
      }

      if (event.state === EVENT_STATE.ENDED) {
        return res.status(CODES.FORBIDDEN).json({
          error: "FORBIDDEN",
          message: "You cannot perform this action. Event has already ended."
        });
      }

      query = new Query("participants", client);

      const entries = await query
        .query<Entry>(
          "SELECT * FROM participants WHERE event_id = $1 ORDER BY signed_out_at DESC;",
          [event.id]
        )
        .catch((err) => next(err));

      if (entries === undefined) {
        return res.status(CODES.INTERNAL_SERVER_ERROR).json({
          error: "INTERNAL_ERROR",
          message: "A lookup error occurred"
        });
      }

      // Reject request if all users are not signed out.
      const unsignedOutEntries = entries.filter(
        (e) => e.signed_out_at === null
      );
      if (unsignedOutEntries.length !== 0) {
        return res.status(CODES.FORBIDDEN).json({
          error: "FORBIDDEN",
          message:
            "You cannot perform this action. There are still users that haven't been signed out."
        });
      }

      query = new Query("events", client);
      try {
        await query.update<Event>([where("id = ?", [event.id])], {
          state: EVENT_STATE.ENDED as EventState
        });
      } catch (err) {
        next(err);

        return res.status(CODES.INTERNAL_SERVER_ERROR).json({
          error: "DATABASE_ERROR",
          message: "An update error occurred"
        });
      }

      const pQuery = new Query("participants", client);
      const rQuery = new Query("rankings", client);
      const ps = new PointsService(entries.length);
      for (const [i, entry] of entries.entries()) {
        const points = ps.calculatePoints(i + 1);

        try {
          await updateRankings(pQuery, rQuery, entry, points, i + 1);
        } catch (err) {
          next(err);

          return res.status(CODES.INTERNAL_SERVER_ERROR).json({
            error: "DATABASE_ERROR",
            message: "An update error occurred"
          });
        }
      }

      client.release();

      return res.status(CODES.OK).end();
    });

    return this.router;
  }
}
