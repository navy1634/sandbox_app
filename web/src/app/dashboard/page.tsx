"use client";

import { useEffect, useState } from "react";
import { authAppLoginURL, type MeResponse } from "@/app/authTypes";
import { useRuntimeConfig } from "@/app/runtimeConfigContext";
import styles from "../page.module.css";

export default function DashboardPage() {
  const [me, setMe] = useState<MeResponse | null>(null);
  const runtimeConfig = useRuntimeConfig();

  useEffect(() => {
    void fetch(`${runtimeConfig.apiBaseURL}/me`, { credentials: "include" })
      .then((response) => response.json() as Promise<MeResponse>)
      .then(setMe)
      .catch(() => setMe({ authenticated: false }));
  }, [runtimeConfig.apiBaseURL]);

  async function logout() {
    await fetch(`${runtimeConfig.apiBaseURL}/auth/logout`, {
      method: "POST",
      credentials: "include",
    });
    setMe({ authenticated: false });
  }

  function login() {
    window.location.href = authAppLoginURL(runtimeConfig);
  }

  if (me === null) {
    return (
      <div className={styles.page}>
        <section className={styles.main}>
          <p className={styles.label}>Loading</p>
          <p>ログイン状態を確認しています。</p>
        </section>
      </div>
    );
  }

  if (!me.authenticated || !me.user) {
    return (
      <div className={styles.page}>
        <section className={styles.main}>
          <div className={styles.intro}>
            <p className={styles.label}>Guest</p>
            <h1>未ログインです</h1>
            <p>SSO ログインを完了すると、ユーザー情報を確認できます。</p>
          </div>
          <div className={styles.ctas}>
            <button className={styles.primary} type="button" onClick={login}>
              認証アプリへ進む
            </button>
          </div>
        </section>
      </div>
    );
  }

  return (
    <div className={styles.page}>
      <section className={styles.main}>
        <div className={styles.intro}>
          <p className={styles.label}>Signed in</p>
          <h1>{me.user.name || me.user.email || `Account ${me.user.accountId}`}</h1>
          <p>sandbox_auth から返るユーザー情報と、アプリ側 DB の保存結果です。</p>
        </div>
        <dl className={styles.details}>
          <div>
            <dt>Account ID</dt>
            <dd>{me.user.accountId}</dd>
          </div>
          <div>
            <dt>Email</dt>
            <dd>{me.user.email || "未設定"}</dd>
          </div>
          <div>
            <dt>Provider</dt>
            <dd>{me.user.provider || "未設定"}（sandbox_auth からの返却値）</dd>
          </div>
          <div>
            <dt>Registration</dt>
            <dd>{me.needsRegistration ? "未完了" : "完了"}（sandbox_auth からの返却値）</dd>
          </div>
          <div>
            <dt>Saved account ID</dt>
            <dd>{me.appAccount?.id ?? "未保存"}</dd>
          </div>
          <div>
            <dt>DB updated</dt>
            <dd>
              {me.appAccount?.updatedAt
                ? new Date(me.appAccount.updatedAt).toLocaleString("ja-JP")
                : "未保存"}
            </dd>
          </div>
        </dl>
        <div className={styles.ctas}>
          <button className={styles.secondary} type="button" onClick={logout}>
            ログアウト
          </button>
        </div>
      </section>
    </div>
  );
}
