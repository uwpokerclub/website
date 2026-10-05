const { TextEncoder, TextDecoder } = require("util");

if (typeof global.TextEncoder === "undefined") {
  global.TextEncoder = TextEncoder;
}
if (typeof global.TextDecoder === "undefined") {
  global.TextDecoder = TextDecoder;
}

// jsdom implements neither ResizeObserver nor element layout. This setup stubs only
// ResizeObserver; element sizes remain unmocked. Chart geometry tests must size
// ResponsiveContainer themselves, as chartGeometry.test.tsx does.
if (typeof global.ResizeObserver === "undefined") {
  global.ResizeObserver = class {
    observe() {}
    unobserve() {}
    disconnect() {}
  };
}
