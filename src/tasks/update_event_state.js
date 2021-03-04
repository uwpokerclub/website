/* eslint-disable no-console */
import { Pool } from "pg";

if (process.env.DATABASE_URL === undefined) {
  console.error(
    "Environment variable DATABASE_URL is not set. Please set it and run this task again."
  );
  process.exit(1);
}

const pool = new Pool({ connectionString: process.env.DATABASE_URL });

const updateAllEventStatesQuery = "UPDATE events SET state = 1;";

async function run() {
  console.log("Starting task...");
  const start = new Date();

  console.log("Updating all events...");
  await pool.query(updateAllEventStatesQuery, []);

  const end = new Date() - start;
  console.log(`Task Successfully completed in ${end}ms`);
}

run().finally(() => pool.end());
