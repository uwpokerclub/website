/* eslint-disable camelcase */

exports.shorthands = undefined;

exports.up = pgm => {
  pgm.createTable("rankings", {
    user_id: {
      type: "bigint",
      references: "users",
      primaryKey: true
    },
    semester_id: {
      type: "uuid",
      references: "semesters",
      primaryKey: true
    },
    points: {
      type: "integer"
    }
  });
};

exports.down = pgm => {
  pgm.dropTable("rankings", {
    cascade: true
  });
};
