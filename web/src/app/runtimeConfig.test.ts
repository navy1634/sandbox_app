import { describe, expect, it } from "vitest";

import { authAppLoginURL } from "./authTypes";
import { createRuntimeConfig } from "./runtimeConfig";

describe("createRuntimeConfig", () => {
  it("環境変数から実行時設定を作る", () => {
    expect(
      createRuntimeConfig({
        NEXT_PUBLIC_API_BASE_URL: "https://api.example.com",
        NEXT_PUBLIC_AUTH_APP_BASE_URL: "https://auth.example.com",
        NEXT_PUBLIC_APP_RETURN_URL: "https://app.example.com/dashboard",
      }),
    ).toEqual({
      apiBaseURL: "https://api.example.com",
      authAppBaseURL: "https://auth.example.com",
      appReturnURL: "https://app.example.com/dashboard",
    });
  });

  it("同一 origin の API を表す空文字を保持する", () => {
    expect(
      createRuntimeConfig({ NEXT_PUBLIC_API_BASE_URL: "" }).apiBaseURL,
    ).toBe("");
  });

  it("環境変数がない場合はローカル開発用の既定値を使う", () => {
    expect(createRuntimeConfig({})).toEqual({
      apiBaseURL: "http://localhost:8080",
      authAppBaseURL: "http://localhost:3000",
      appReturnURL: undefined,
    });
  });
});

describe("authAppLoginURL", () => {
  it("実行時設定からリダイレクト先を含むログインURLを作る", () => {
    expect(
      authAppLoginURL({
        apiBaseURL: "",
        authAppBaseURL: "https://auth.example.com",
        appReturnURL: "https://app.example.com/dashboard",
      }),
    ).toBe(
      "https://auth.example.com/login?redirect_to=https%3A%2F%2Fapp.example.com%2Fdashboard",
    );
  });
});
