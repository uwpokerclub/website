/* eslint-disable camelcase */

exports.shorthands = undefined;

exports.up = pgm => {
  pgm.addColumns("semesters", {
    uuid: {
      type: "uuid",
      unique: true,
      notNull: true,
      default: pgm.func("uuid_generate_v4()")
    }
  }, {
    ifNotExists: true
  });

  pgm.addColumns("events", {
    semester_uuid: {
      type: "uuid",
    }
  }, {
    ifNotExists: true
  });

  pgm.addColumns("users", {
    semester_id: {
      type: "uuid",
    }
  }, {
    ifNotExists: true
  });

  // Update events table
  pgm.sql("UPDATE events SET semester_uuid = semesters.uuid FROM semesters WHERE events.semester = semesters.id;");

  pgm.alterColumn("events", "semester_uuid", {
    notNull: true
  });

  pgm.dropColumn("events", "semester", { ifExists: true, cascade: true });

  pgm.renameColumn("events", "semester_uuid", "semester_id");

  pgm.createIndex("events", "semester_id", { ifNotExists: true });

  pgm.addConstraint("events", "events_semester_id_fkey", {
    foreignKeys: {
      columns: "semester_id",
      references: "semesters(uuid)"
    }
  });

  // Update user table
  pgm.sql("UPDATE users SET semester_id = semesters.uuid FROM semesters WHERE users.last_semester_registered = semesters.id");

  pgm.alterColumn("users", "semester_id", {
    notNull: true
  });

  pgm.dropColumn("users", "last_semester_registered", { ifExists: true, cascade: true });

  pgm.createIndex("users", "semester_id", { ifNotExists: true });

  pgm.addConstraint("users", "users_semester_id_fkey", {
    foreignKeys: {
      columns: "semester_id",
      references: "semesters(uuid)"
    }
  });

  // Clean up semester table
  pgm.dropColumn("semesters", "id", { ifExists: true, cascade: true });

  pgm.renameColumn("semesters", "uuid", "id");

  pgm.addConstraint("semesters", "semesters_pkey", {
    primaryKey: "id"
  });
};

exports.down = pgm => {
  throw "Irreverisble Migration";
};
