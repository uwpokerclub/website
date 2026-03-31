module.exports = {
  e2e: {
    setupNodeEvents(on, config) {
      // implement node event listeners here
    },
    baseUrl: "http://localhost:5000",
    testIsolation: true,
    retries: {
      runMode: 2,
      openMode: 0,
    },
  },
  projectId: "rxir4a",
};
