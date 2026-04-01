module.exports = {
  e2e: {
    setupNodeEvents(on, config) {
      on("task", {
        log(message) {
          console.log(message);
          return null;
        },
        table(message) {
          console.table(message);
          return null;
        },
      });
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
