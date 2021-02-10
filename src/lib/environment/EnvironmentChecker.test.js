const EnvironmentChecker = require("./EnvironmentChecker");

describe("EnvironmentChecker", () => {
  test("calling verify() should not throw an error if all vars are set", () => {
    // mock env vars being set
    process.env.TEST_VAR = "HELLO";
    process.env.TEST_VAR2 = "WORLD";

    const ec = new EnvironmentChecker({
      required: ["TEST_VAR", "TEST_VAR2"]
    });

    expect(() => ec.verify()).not.toThrow();
  });

  test("calling verify() should throw an error when a variable is missing", () => {
    const ec = new EnvironmentChecker({
      required: ["TEST_VAR", "TEST_VAR2", "TEST_VAR3"]
    });

    expect(() => ec.verify()).toThrow();
  });
});
