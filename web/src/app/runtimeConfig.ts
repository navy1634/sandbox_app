export type RuntimeConfig = {
  apiBaseURL: string;
};

type RuntimeConfigEnvironment = Readonly<Record<string, string | undefined>>;

export function createRuntimeConfig(
  environment: RuntimeConfigEnvironment,
): RuntimeConfig {
  return {
    apiBaseURL: environment.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080",
  };
}
