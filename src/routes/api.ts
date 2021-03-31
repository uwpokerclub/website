import { Router } from "express";
import RouteHandler from "../lib/route_handler/RouteHandler";

import LoginRouteHandler from "./login";
import UsersRouteHandler from "./users";
import EventsRouteHandler from "./events";
import SemestersRouteHandler from "./semesters";
import ParticipantsRouteHandler from "./participants";

import requireAuthentication from "../middleware/authenticate";

export default class APIRouteHandler extends RouteHandler {
  public handler(): Router {
    const loginRoute = new LoginRouteHandler("/login", this.db);
    const usersRoute = new UsersRouteHandler("/users", this.db);
    const eventsRoute = new EventsRouteHandler("/events", this.db);
    const semestersRoute = new SemestersRouteHandler("/semesters", this.db);
    const participantsRoute = new ParticipantsRouteHandler(
      "/participants",
      this.db
    );

    this.router.use(loginRoute.path, loginRoute.handler());
    this.router.use(
      usersRoute.path,
      requireAuthentication,
      usersRoute.handler()
    );
    this.router.use(
      eventsRoute.path,
      requireAuthentication,
      eventsRoute.handler()
    );
    this.router.use(
      semestersRoute.path,
      requireAuthentication,
      semestersRoute.handler()
    );
    this.router.use(
      participantsRoute.path,
      requireAuthentication,
      participantsRoute.handler()
    );

    return this.router;
  }
}
