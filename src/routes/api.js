const RouteHandler = require("../lib/route_handler/RouteHandler");

const LoginRouteHandler = require("./login");

class APIRouteHandler extends RouteHandler {
  handler() {
    const loginRoute = new LoginRouteHandler("/login", this.db);

    this.router.use(loginRoute.path, loginRoute.handler());

    return this.router;
  }
}

module.exports = APIRouteHandler;
