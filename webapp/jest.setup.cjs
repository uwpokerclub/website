const { TextEncoder, TextDecoder } = require("util");

if (typeof global.TextEncoder === "undefined") {
  global.TextEncoder = TextEncoder;
}
if (typeof global.TextDecoder === "undefined") {
  global.TextDecoder = TextDecoder;
}

// jsdom implements neither ResizeObserver nor element layout, both of which
// Recharts' ResponsiveContainer requires. Chart tests assert the card's text and
// states, not SVG geometry, so a no-op observer and a fixed box are enough.
if (typeof global.ResizeObserver === "undefined") {
  global.ResizeObserver = class {
    observe() {}
    unobserve() {}
    disconnect() {}
  };
}
