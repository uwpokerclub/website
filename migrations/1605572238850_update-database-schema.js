/* eslint-disable camelcase */

exports.shorthands = undefined;

exports.up = pgm => {
  // users
  pgm.renameColumn("users", "firstname", "first_name");
  pgm.renameColumn("users", "lastname", "last_name");
  pgm.renameColumn("users", "questid", "quest_id");
  pgm.renameColumn("users", "created", "created_at");

  pgm.alterColumn("users", "created_at", {
    notNull: true
  });

  // semesters
  pgm.renameColumn("semesters", "startdate", "start_date");
  pgm.renameColumn("semesters", "enddate", "end_date");

  pgm.alterColumn("semesters", "start_date", {
    notNull: true,
    default: pgm.func("current_date")
  });
  pgm.alterColumn("semesters", "end_date", {
    notNull: true,
    default: pgm.func("current_date")
  });

  // events
  pgm.renameColumn("events", "startdate", "start_date");
  pgm.alterColumn("events", "start_date", {
    notNull: true,
    default: pgm.func("current_timestamp")
  });

  // participants
  pgm.renameColumn("participants", "id", "user_id");
  pgm.renameColumn("participants", "num", "id");
  pgm.renameColumn("participants", "eventid", "event_id");
  pgm.renameConstraint("participants", "participants_eventid_fkey", "participants_event_id_fkey");
  pgm.renameConstraint("participants", "participants_id_fkey", "participants_user_id_fkey");

  // logins
  pgm.renameColumn("logins", "loginid", "username");
  pgm.renameColumn("logins", "pass", "password");
};

exports.down = pgm => {
  // users
  pgm.renameColumn("users", "first_name", "firstname");
  pgm.renameColumn("users", "last_name", "lastname");
  pgm.renameColumn("users", "quest_id", "questid");
  pgm.renameColumn("users", "created_at", "created");

  pgm.alterColumn("users", "created", {
    notNull: false
  });

  // semesters
  pgm.renameColumn("semesters", "start_date", "startdate");
  pgm.renameColumn("semesters", "end_date", "enddate");

  pgm.alterColumn("semesters", "startdate", {
    notNull: false,
    default: null
  });
  pgm.alterColumn("semesters", "enddate", {
    notNull: false,
    default: null
  });

  // events
  pgm.renameColumn("events", "start_date", "startdate");
  pgm.alterColumn("events", "startdate", {
    notNull: false,
    default: null
  });

  // participants
  pgm.renameColumn("participants", "id", "num");
  pgm.renameColumn("participants", "user_id", "id");
  pgm.renameColumn("participants", "event_id", "eventid");
  pgm.renameConstraint("participants", "participants_event_id_fkey", "participants_eventid_fkey");
  pgm.renameConstraint("participants", "participants_user_id_fkey", "participants_id_fkey");

  // logins
  pgm.renameColumn("logins", "username", "loginid");
  pgm.renameColumn("logins", "password", "pass");
};
