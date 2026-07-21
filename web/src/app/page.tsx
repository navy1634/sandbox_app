"use client";

import Link from "next/link";
import { authAppLoginURL } from "@/app/authTypes";
import styles from "./page.module.css";

export default function Home() {
  function login() {
    window.location.href = authAppLoginURL();
  }

  return (
    <div className={styles.page}>
      <main className={styles.main}>
        <div className={styles.intro}>
          <p className={styles.label}>Auth Sandbox</p>
          <h1>SSO 動作確認アプリ</h1>
          <p>
            sandbox_auth の認証画面から戻った後に、返ってくるユーザー情報と
            アプリ側 DB への保存結果を確認できます。
          </p>
        </div>
        <div className={styles.ctas}>
          <button className={styles.primary} type="button" onClick={login}>
            認証アプリへ進む
          </button>
          <Link className={styles.secondary} href="/dashboard">
            セッションを確認する
          </Link>
        </div>
      </main>
    </div>
  );
}
