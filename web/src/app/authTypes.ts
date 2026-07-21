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

export const apiBaseURL =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

export const authAppBaseURL =
  process.env.NEXT_PUBLIC_AUTH_APP_BASE_URL ?? "http://localhost:3000";

export function appReturnURL() {
  if (process.env.NEXT_PUBLIC_APP_RETURN_URL) {
    return process.env.NEXT_PUBLIC_APP_RETURN_URL;
  }

  return `${window.location.origin}/dashboard`;
}

export function authAppLoginURL() {
  return `${authAppBaseURL}/login?redirect_to=${encodeURIComponent(appReturnURL())}`;
}
