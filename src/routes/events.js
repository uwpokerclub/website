/* eslint-disable camelcase */
import RouteHandler from "../lib/route_handler/RouteHandler";
import PointsService from "../lib/services/points_service/PointsService";
import { CODES, EVENT_STATE } from "../models/constants";

function validateCreateReq(body) {
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

async function updateRankings(db, entry, points, placement, semesterId) {
  await db
    .table("participants")
    .update({
      placement
    })
    .where("user_id = ? AND event_id = ?", entry.user_id, entry.event_id)
    .execute();

  const [ranking] = await db
    .table("rankings")
    .select()
    .where("user_id = ? AND semester_id = ?", entry.user_id, semesterId)
    .execute();

  if (ranking === undefined) {
    await db
      .table("rankings")
      .create({
        user_id: entry.user_id,
        semester_id: semesterId,
        points
      })
      .execute();
  } else {
    await db
      .table("rankings")
      .update({
        points: ranking.points + points
      })
      .where("user_id = ? AND semester_id = ?", entry.user_id, semesterId)
      .execute();
  }
}

class EventsRouteHandler extends RouteHandler {
  handler() {
    this.router.get("/", async (req, res, next) => {
      const { semesterId } = req.query;

      let events;
      if (semesterId !== undefined) {
        events = await this.db
          .query(
            `SELECT * FROM events LEFT JOIN
        (SELECT event_id, COUNT(*) FROM participants GROUP BY event_id)
        AS entries ON events.id = entries.event_id WHERE events.semester_id = $1
        ORDER BY events.start_date DESC;`,
            semesterId
          )
          .catch((err) => {
            res.status(CODES.INTERNAL_SERVER_ERROR).json({
              error: "DATABASE_ERROR",
              message: "A lookup error occurred"
            });

            next(err);
          });
      } else {
        events = await this.db
          .query(
            `SELECT * FROM events LEFT JOIN
        (SELECT event_id, COUNT(*) FROM participants GROUP BY event_id)
        AS entries ON events.id = entries.event_id ORDER BY events.start_date DESC;`
          )
          .catch((err) => {
            res.status(CODES.INTERNAL_SERVER_ERROR).json({
              error: "DATABASE_ERROR",
              message: "A lookup error occurred"
            });

            next(err);
          });
      }

      return res.status(CODES.OK).json({ events });
    });

    this.router.get("/:id", async (req, res, next) => {
      const { id } = req.params;

      const [event] = await this.db
        .table("events")
        .select()
        .where("id = ?", id)
        .execute()
        .catch((err) => {
          res.status(CODES.INTERNAL_SERVER_ERROR).json({
            error: "DATABASE_ERROR",
            message: "A lookup error occurred"
          });

          next(err);
        });

      if (event === undefined) {
        return res.status(CODES.NOT_FOUND).json({
          error: "NOT_FOUND",
          message: "That event could not be found"
        });
      }

      return res.status(CODES.OK).json({ event });
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

      const { name, startDate, format, notes, semesterId } = req.body;

      await this.db
        .table("events")
        .create({
          name,
          start_date: startDate,
          format,
          notes,
          semester_id: semesterId
        })
        .execute()
        .catch((err) => {
          res.status(CODES.INTERNAL_SERVER_ERROR).json({
            error: "DATABASE_ERROR",
            message: "An insertion error occurred"
          });

          next(err);
        });

      return res.status(CODES.CREATED).json({
        user: req.body
      });
    });

    this.router.post("/:id/end", async (req, res, next) => {
      const { id } = req.params;

      const [event] = await this.db
        .table("events")
        .select()
        .where("id = ?", id)
        .execute()
        .catch((err) => {
          res.status(CODES.INTERNAL_SERVER_ERROR).json({
            error: "INTERNAL_ERROR",
            message: "A lookup error occurred"
          });

          next(err);
        });

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

      const entries = await this.db
        .query(
          `SELECT * FROM participants
      WHERE event_id = $1 ORDER BY signed_out_at DESC;`,
          event.id
        )
        .catch((err) => {
          res.status(CODES.INTERNAL_SERVER_ERROR).json({
            error: "INTERNAL_ERROR",
            message: "A lookup error occurred"
          });

          next(err);
        });

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

      await this.db
        .table("events")
        .update({
          state: EVENT_STATE.ENDED
        })
        .where("id = ?", id)
        .execute()
        .catch((err) => {
          res.status(CODES.INTERNAL_SERVER_ERROR).json({
            error: "DATABASE_ERROR",
            message: "An update error occurred"
          });

          next(err);
        });

      const ps = new PointsService(entries.length);

      for (const [i, entry] of entries.entries()) {
        const points = ps.calculatePoints(i + 1);

        await updateRankings(
          this.db,
          entry,
          points,
          i + 1,
          event.semester_id
        ).catch((err) => {
          res.status(CODES.INTERNAL_SERVER_ERROR).json({
            error: "DATABASE_ERROR",
            message: "An update error occurred"
          });

          next(err);
        });
      }

      return res.status(CODES.OK).end();
    });

    return this.router;
  }
}

export default EventsRouteHandler;
