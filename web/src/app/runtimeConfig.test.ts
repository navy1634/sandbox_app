import { describe, expect, it } from "vitest";

import { appLoginURL } from "./authTypes";
import { createRuntimeConfig } from "./runtimeConfig";

describe("createRuntimeConfig", () => {
  it("環境変数から実行時設定を作る", () => {
    expect(
      createRuntimeConfig({
        NEXT_PUBLIC_API_BASE_URL: "https://api.example.com",
      }),
    ).toEqual({
      apiBaseURL: "https://api.example.com",
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
    });
  });
});

describe("appLoginURL", () => {
  it("アプリAPIのOIDCログインcallback開始URLを作る", () => {
    expect(appLoginURL({ apiBaseURL: "https://app.example.com" })).toBe(
      "https://app.example.com/auth/login",
    );
  });
});
