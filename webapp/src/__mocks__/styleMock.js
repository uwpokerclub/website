// CSS-module stand-in: every class name resolves to itself, so tests can assert on
// them and components receive real strings.
//
// __esModule must answer false. Returning the key for it (as this mock used to)
// makes Babel's interop treat the proxy as an ES module and take `.default`, so
// `styles` became the string "default" and every `styles.x` was undefined — or
// worse, a String.prototype method: `styles.sub` returned a function, which React
// rejects as a className.
module.exports = new Proxy(
  {},
  {
    get: (_target, key) => (key === "__esModule" ? false : key),
  },
);
