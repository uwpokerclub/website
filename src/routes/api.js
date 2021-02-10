const RouteHandler = require("../lib/route_handler/RouteHandler");

class APIRouteHandler extends RouteHandler {
  handler() {
    return this.router;
  }
}

module.exports = APIRouteHandler
