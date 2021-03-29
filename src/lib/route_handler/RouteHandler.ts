import { ConnectionPool } from "postgres-driver-service";
import { Router as ExpressRouter } from "express";
import Router from "express-promise-router";

export default class RouteHandler {
  /**
   * db is used to run queries against the database.
   */
  public db: ConnectionPool;

  /**
   * router is the internal Express router.
   */
  public router: ExpressRouter;

  /**
   * path is the url path for this route.
   */
  public path: string;

  public constructor(path: string, db: ConnectionPool) {
    this.path = path;
    this.db = db;
    this.router = Router();
  }

  public getPath(): string {
    return this.path;
  }

  public handler(): ExpressRouter {
    throw new Error("ERR_NOT_IMPLEMENTED");
  }
}

module.exports = RouteHandler;
