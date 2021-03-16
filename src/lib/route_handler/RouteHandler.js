const Router = require("express-promise-router");

class RouteHandler {
  constructor(path, db) {
    this.path = path;
    this.db = db;
    this.router = new Router();
  }

  getPath() {
    return this.path;
  }

  handler() {
    throw new Error("ERR_NOT_IMPLEMENTED");
  }
}

module.exports = RouteHandler;
