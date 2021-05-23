/* eslint-disable camelcase */

exports.shorthands = undefined;

exports.up = pgm => {
  // Add membership table
  pgm.createTable("memberships", {
    id: {
      type: "uuid",
      primaryKey: true,
      default: pgm.func("uuid_generate_v4()")
    },
    user_id: {
      type: "bigint",
      references: "users"
    },
    semester_id: {
      type: "uuid",
      references: "semesters"
    },
    paid: {
      type: "boolean",
      default: "FALSE",
      notNull: true
    }
  });

  // Create existing memberships
  pgm.sql("INSERT INTO memberships (user_id, semester_id, paid) SELECT users.id, users.semester_id, users.paid FROM users;")

  pgm.dropColumn("users", ["semester_id", "paid"], { cascade: true })

  // Add column to participants
  pgm.addColumns("participants", {
    membership_id: {
      type: "uuid",
      references: "memberships"
    }
  });

  // Update participant table with membership ids
  pgm.sql("UPDATE participants SET membership_id = memberships.id FROM memberships WHERE participants.user_id = memberships.user_id;")

  // Drop user_id column and create new primary key
  pgm.dropColumn("participants", "user_id", { cascade: true })
  pgm.addConstraint("participants", "participants_pkey", { primaryKey: ["membership_id", "event_id"]})

  // Add membership column to rankings
  pgm.addColumns("rankings", {
    membership_id: {
      type: "uuid",
      references: "memberships"
    }
  })

  // Update rankings table to set membership_id
  pgm.sql("UPDATE rankings SET membership_id = memberships.id FROM memberships WHERE rankings.user_id = memberships.user_id;")

  // Drop user
  pgm.dropColumn("rankings", ["user_id", "semester_id"], { cascade: true })
  pgm.addConstraint("rankings", "rankings_pkey", { primaryKey: "membership_id" })
};

exports.down = pgm => {
  // Add back user_id and semester_id columns
  pgm.addColumns("rankings", {
    user_id: {
      type: "bigint",
      references: "users"
    },
    semester_id: {
      type: "uuid",
      references: "semesters"
    }
  })

  pgm.sql("UPDATE rankings SET user_id = memberships.user_id, semester_id = memberships.semester_id FROM memberships WHERE rankings.membership_id = memberships.id;")

  pgm.dropColumn("rankings", "membership_id", { cascade: true })
  pgm.addConstraint("rankings", "rankings_pkey", { primaryKey: ["user_id", "semester_id"] })

  pgm.addColumns("participants", {
    user_id: {
      type: "bigint",
      references: "users"
    }
  })

  pgm.sql("UPDATE participants SET user_id = memberships.user_id FROM memberships WHERE participants.membership_id = memberships.id;")

  pgm.dropColumn("participants", "membership_id", { cascade: true })
  pgm.addConstraint("participants", "participants_pkey", { primaryKey: ["user_id", "event_id"] })

  pgm.addColumn("users", {
    paid: {
      type: "boolean",
      default: "false",
      notNull: true
    },
    semester_id: {
      type: "uuid",
      references: "semesters"
    }
  })

  pgm.sql("UPDATE users SET paid = memberships.paid, semester_id = memberships.semester_id FROM memberships WHERE users.id = memberships.user_id;")

  pgm.dropTable("memberships", { cascade: true })
};
