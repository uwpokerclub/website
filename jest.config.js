module.exports = {
  globals: {
    "ts-jest": {
      tsConfig: "tsconfig.json"
    }
  },
  moduleFileExtensions: ["ts", "js", "d.ts"],
  transform: {
    "^.+\\.(ts|tsx)$": "ts-jest"
  },
  testMatch: ["**/src/**/*.test.(ts|js)", "**/src/**/test/**/*.test.(ts|js)"],
  testEnvironment: "node"
};
