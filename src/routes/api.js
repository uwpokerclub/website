const RouteHandler = require("../lib/route_handler/RouteHandler");

const LoginRouteHandler = require("./login");
const UsersRouteHandler = require("./users");
const EventsRouteHandler = require("./events");

const { requireAuthentication } = require("../middleware/authenticate");

class APIRouteHandler extends RouteHandler {
  handler() {
    const loginRoute = new LoginRouteHandler("/login", this.db);
    const usersRoute = new UsersRouteHandler("/users", this.db);
    const eventsRoute = new EventsRouteHandler("/events", this.db);

    this.router.use(loginRoute.path, loginRoute.handler());
    this.router.use(usersRoute.path, requireAuthentication, usersRoute.handler());
    this.router.use(eventsRoute.path, requireAuthentication, eventsRoute.handler());

    return this.router;
  }
}

module.exports = APIRouteHandler;
