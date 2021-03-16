/* eslint-disable no-console */
const { Pool } = require("pg");
const { DriverService } = require("postgres-driver-service");

const EnvironmentChecker = require("./lib/environment/EnvironmentChecker");
const Server = require("./lib/server/Server");

const environmentConfig = require("../config/environment.json");

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
