/* eslint-disable camelcase */

exports.shorthands = undefined;

exports.up = pgm => {
  pgm.addColumns("semesters", {
    starting_budget: {
      type: "decimal",
      notNull: true,
      default: "0"
    },
    current_budget: {
      type: "decimal",
      notNull: true,
      default: "0"
    },
    membership_fee: {
      type: "smallint",
      notNull: true,
      default: "0"
    },
    membership_discount_fee: {
      type: "smallint",
      notNull: true,
      default: "0"
    },
    rebuy_fee: {
      type: "smallint",
      notNull: true,
      default: "0"
    }
  });

  pgm.addColumns("memberships", {
    discounted: {
      type: "boolean",
      notNull: true,
      default: "FALSE"
    }
  });

  pgm.addColumns("participants", {
    rebuys: {
      type: "smallint",
      notNull: true,
      default: "0"
    }
  });

  pgm.createTable("transactions", {
    id: {
      type: "serial",
      notNull: true,
      primaryKey: true
    },
    semester_id: {
      type: "uuid",
      references: "semesters"
    },
    amount: {
      type: "decimal",
      notNull: true,
      default: "0"
    },
    description: {
      type: "text"
    }
  });
};

exports.down = pgm => {
  pgm.dropColumns("semesters", ["starting_budget", "current_budget", "membership_fee", "membership_discount_fee", "rebuy_fee"]);

  pgm.dropColumns("memberships", ["discounted"]);

  pgm.dropColumns("participants", ["rebuys"]);

  pgm.dropTable("transactions");
};
