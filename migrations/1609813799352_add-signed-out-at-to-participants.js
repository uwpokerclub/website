/* eslint-disable camelcase */

exports.shorthands = undefined;

exports.up = pgm => {
  pgm.addColumns("participants", {
    signed_out_at: {
      type: "timestamp"
    }
  });
};

exports.down = pgm => {
  pgm.dropColumns("participants", ["signed_out_at"], { cascade: true });
};
