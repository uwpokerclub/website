/* eslint-disable no-console */
import { Pool } from "pg";
import { DriverService } from "postgres-driver-service";

import EnvironmentChecker from "./lib/environment/EnvironmentChecker";
import Server from "./lib/server/Server";

import environmentConfig from "../config/environment.json";

// Initialize an environment checker and verify all variables are present
const ec = new EnvironmentChecker(environmentConfig);
try {
  ec.verify();
} catch (err) {
  console.error(`Error verifying environment variables: ${err.message}`);
  process.exit(1);
}

// Initalize postgres pool and the postgres driver service
const pool = new Pool({ connectionString: process.env.DATABASE_URL });
const dbs = new DriverService(pool);

// Initalize server
const server = new Server(dbs);
server.run();
