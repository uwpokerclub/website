/* eslint-disable camelcase */

exports.shorthands = undefined;

exports.up = pgm => {
  pgm.addColumns("events", {
    state: {
      type: "smallint",
      default: 0
    },
  });
};

exports.down = pgm => {
  pgm.dropColumns("events", ["state"], { cascade: true });
};
