type EnvironmentConfig = {
  required: string[];
}

export default class EnvironmentChecker {
  private environmentConfig: EnvironmentConfig

  public constructor(environmentConfig: EnvironmentConfig) {
    this.environmentConfig = environmentConfig;
  }

  public verify(): void {
    for (const envVar of this.environmentConfig.required) {
      if (process.env[envVar] === undefined) {
        throw new Error(`ERR_MISSING_ENV_VAR: ${envVar}`);
      }
    }
  }
}
