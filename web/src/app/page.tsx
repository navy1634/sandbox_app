"use client";

import Link from "next/link";
import { appLoginURL } from "@/app/authTypes";
import { useRuntimeConfig } from "@/app/runtimeConfigContext";
import styles from "./page.module.css";

export default function Home() {
  const runtimeConfig = useRuntimeConfig();

  function login() {
    window.location.href = appLoginURL(runtimeConfig);
  }

  return (
    <div className={styles.page}>
      <main className={styles.main}>
        <div className={styles.intro}>
          <p className={styles.label}>Auth Sandbox</p>
          <h1>SSO 動作確認アプリ</h1>
          <p>
            sandbox_auth のOIDC callbackから戻った後に、返ってくるユーザー情報と
            アプリ側 DB への保存結果を確認できます。
          </p>
        </div>
        <div className={styles.ctas}>
          <button className={styles.primary} type="button" onClick={login}>
            OIDCでログイン
          </button>
          <Link className={styles.secondary} href="/dashboard">
            セッションを確認する
          </Link>
        </div>
      </main>
    </div>
  );
}
