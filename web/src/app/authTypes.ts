import type { RuntimeConfig } from "./runtimeConfig";

export type Account = {
  id: number;
  identity?: {
    provider: string;
    providerAccountId: string;
    email: string;
    emailVerified: boolean;
    name: string;
    picture: string;
  };
  displayName: string;
  bio: string;
  registeredAt: string | null;
};

export type MeResponse = {
  authenticated: boolean;
  account?: Account;
  appAccount?: {
    id: number;
    email?: string;
    name?: string;
    picture?: string;
    createdAt: string;
    updatedAt: string;
  };
  needsRegistration?: boolean;
  redirectTo?: string;
  user?: {
    accountId: number;
    provider?: string;
    providerAccountId?: string;
    email?: string;
    name?: string;
    picture?: string;
  };
};

export function appReturnURL(runtimeConfig: RuntimeConfig) {
  if (runtimeConfig.appReturnURL) {
    return runtimeConfig.appReturnURL;
  }

  return `${window.location.origin}/dashboard`;
}

export function authAppLoginURL(runtimeConfig: RuntimeConfig) {
  return `${runtimeConfig.authAppBaseURL}/login?redirect_to=${encodeURIComponent(appReturnURL(runtimeConfig))}`;
}
