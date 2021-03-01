class EnvironmentChecker {
  constructor(environmentConfig) {
    this.environmentConfig = environmentConfig;
  }

  verify() {
    for (const envVar of this.environmentConfig.required) {
      if (process.env[envVar] === undefined) {
        throw new Error(`ERR_MISSING_ENV_VAR: ${envVar}`);
      }
    }
  }
}

module.exports = EnvironmentChecker;
