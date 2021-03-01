/* eslint-disable camelcase */
exports.shorthands = undefined;

exports.up = pgm => {
  pgm.createTable("users", {
    id: { type: "bigint", primaryKey: true },
    firstname: { type: "varchar" },
    lastname: { type: "varchar" },
    email: { type: "varchar" },
    faculty: { type: "varchar" },
    questid: { type: "varchar" },
    paid: { type: "boolean" },
    last_semester_registered: { type: "integer" },
    created: {
      type: "date",
      default: pgm.func("current_date")
    }
  }, {
    ifNotExists: true
  });

  pgm.createTable("semesters", {
    id: "id",
    name: { type: "varchar" },
    meta: { type: "varchar" },
    startdate: {
      type: "date"
    },
    enddate: {
      type: "date"
    }
  }, {
    ifNotExists: true
  });

  pgm.createTable("events", {
    id: "id",
    name: { type: "varchar" },
    format: { type: "varchar" },
    notes: { type: "varchar" },
    semester: {
      type: "serial",
      references: "semesters"
    },
    startdate: {
      type: "timestamptz"
    }
  }, {
    ifNotExists: true
  });

  pgm.createTable("participants", {
    num: { type: "serial" },
    id: {
      type: "bigint",
      references: "users",
      primaryKey: true
    },
    eventid: {
      type: "serial",
      references: "events",
      primaryKey: true
    },
    placement: { type: "integer" }
  }, {
    ifNotExists: true
  });

  pgm.createTable("logins", {
    loginid: { type: "varchar", primaryKey: true },
    pass: { type: "varchar" }
  }, {
    ifNotExists: true
  });
};
