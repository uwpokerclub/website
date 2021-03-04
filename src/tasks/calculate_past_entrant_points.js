/* eslint-disable max-len */
/* eslint-disable camelcase */
/* eslint-disable no-console */
import { Pool } from "pg";

import PointsService from "../lib/services/points_service/PointsService";

if (process.env.DATABASE_URL === undefined) {
  console.error(
    "Environment variable DATABASE_URL is not set. Please set it and run this task again."
  );
  process.exit(1);
}

const pool = new Pool({ connectionString: process.env.DATABASE_URL });

const getAllEventsQuery = "SELECT * FROM events;";
const getAllEntriesForEventsQuery =
  "SELECT * FROM participants WHERE event_id = $1;";
const updateUserPlacementQuery =
  "UPDATE participants SET placement = $1 WHERE user_id = $2 AND event_id = $3;";
const findUserRankingQuery =
  "SELECT * FROM rankings WHERE user_id = $1 AND semester_id = $2;";
const createUserRankingQuery =
  "INSERT INTO rankings (user_id, semester_id, points) VALUES ($1, $2, $3);";
const updateUserRankingQuery =
  "UPDATE rankings SET points = $1 WHERE user_id = $2 AND semester_id = $3";

async function run() {
  console.log("Starting task...");
  const start = new Date();

  const client = await pool.connect();

  await client.query("BEGIN").catch((err) => {
    console.error(`Failed to begin transaction: ${err}`);
    process.exit(1);
  });

  try {
    console.log("Pulling all events...");
    const { rows: events } = await client.query(getAllEventsQuery, []);

    for (const event of events) {
      console.log(`Pulling all entries for event ${event.id}...`);
      const { rows: entries } = await client.query(
        getAllEntriesForEventsQuery,
        [event.id]
      );

      const pointsService = new PointsService(entries.length);

      const unsignedOutUsersPlacements = entries
        .filter((e) => e.placement === null)
        .map((_, i) => i + 1);

      for (const [i, entry] of entries.entries()) {
        if (entry.placement === null) {
          console.log(`Updating un-signed out user ${entry.user_id}...`);
          await client.query(updateUserPlacementQuery, [
            unsignedOutUsersPlacements[i],
            entry.user_id,
            entry.event_id
          ]);

          const points = pointsService.calculatePoints(
            unsignedOutUsersPlacements[i]
          );

          console.log(`Calculating points for user ${entry.user_id}...`);
          const { rows: rankings } = await client.query(findUserRankingQuery, [
            entry.user_id,
            event.semester_id
          ]);
          if (rankings.length === 0) {
            // Create new ranking
            console.log(
              `User ${entry.user_id} doesn't have any rankings, creating one...`
            );
            await client.query(createUserRankingQuery, [
              entry.user_id,
              event.semester_id,
              points
            ]);
          } else {
            // Update ranking
            console.log(`Updating rankings for user ${entry.user_id}...`);
            const [rank] = rankings;
            await client.query(updateUserRankingQuery, [
              rank.points + points,
              entry.user_id,
              event.semester_id
            ]);
          }
        } else {
          const points = pointsService.calculatePoints(entry.placement);

          console.log(`Calculating points for user ${entry.user_id}...`);
          const { rows: rankings } = await client.query(findUserRankingQuery, [
            entry.user_id,
            event.semester_id
          ]);
          if (rankings.length === 0) {
            // Create new ranking
            console.log(
              `User ${entry.user_id} doesn't have any rankings, creating one...`
            );
            await client.query(createUserRankingQuery, [
              entry.user_id,
              event.semester_id,
              points
            ]);
          } else {
            // Update ranking
            console.log(`Updating rankings for user ${entry.user_id}...`);
            const [rank] = rankings;
            await client.query(updateUserRankingQuery, [
              rank.points + points,
              entry.user_id,
              event.semester_id
            ]);
          }
        }
      }
    }

    await client.query("COMMIT");
  } catch (err) {
    await client
      .query("ROLLBACK")
      .then(() => console.error("Rollback was successful."))
      .catch((rerr) => {
        console.error(`Failed to rollback query: ${rerr}`);
        process.exit(1);
      });
  } finally {
    client.release();
  }

  const end = new Date() - start;
  console.log(`Task Successfully completed in ${end}ms`);
}

run().then(() => {
  pool.end();
});
