export type RuntimeConfig = {
  apiBaseURL: string;
  authAppBaseURL: string;
  appReturnURL?: string;
};

type RuntimeConfigEnvironment = Readonly<Record<string, string | undefined>>;

export function createRuntimeConfig(
  environment: RuntimeConfigEnvironment,
): RuntimeConfig {
  return {
    apiBaseURL: environment.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080",
    authAppBaseURL:
      environment.NEXT_PUBLIC_AUTH_APP_BASE_URL ?? "http://localhost:3000",
    appReturnURL: environment.NEXT_PUBLIC_APP_RETURN_URL,
  };
}
