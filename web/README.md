# app

Next.js のフロントエンドです。アプリAPIのOIDCログイン開始URLへ遷移し、callback後のユーザー情報とDB保存結果を表示します。

## 主な画面

| パス | 内容 |
| ---- | ---- |
| `/` | トップページです。OIDCログインを開始できます。 |
| `/dashboard` | OIDC IDトークンから得たユーザー情報とDB保存結果を確認できます。 |

## バックエンド連携

API の接続先は `NEXT_PUBLIC_API_BASE_URL` で指定します。

```txt
NEXT_PUBLIC_API_BASE_URL=http://localhost:8081
```

Docker Compose で起動する場合は、ルートディレクトリの `SANDBOX_API_URL` が `NEXT_PUBLIC_API_BASE_URL` に渡されます。

Docker イメージは環境ごとに作り分けません。次の同一コマンドでビルドし、接続先はコンテナ起動時の環境変数または Kubernetes の `ConfigMap` で指定します。

```txt
docker build . -t sandbox-app-web
```

sandbox_auth側では、OIDC clientのcallback URLを完全一致で登録してください。

```txt
OIDC_CLIENTS=[{"client_id":"external-app","client_secret":"replace-with-a-random-secret","redirect_uris":["http://localhost:8081/auth/callback"]}]
```

## ローカル起動

```txt
pnpm install
pnpm dev
```

起動後は Next.js が表示した `Local` の URL にアクセスしてください。通常は `http://localhost:3000` です。`3000` が既に使われている場合は、Next.js が `3001` などの空きポートで起動します。

## コマンド

| コマンド | 内容 |
| -------- | ---- |
| `pnpm dev` | Next.js を開発起動します。 |
| `pnpm build` | Next.js をビルドします。 |
| `pnpm start` | ビルド済みの Next.js を起動します。 |
| `pnpm test` | 単体テストを実行します。 |
| `pnpm lint` | JavaScript と CSS の lint を実行します。 |
| `pnpm lint:js` | oxlint を実行します。 |
| `pnpm lint:css` | stylelint を実行します。 |
| `pnpm format` | oxfmt で整形します。 |
| `pnpm typecheck` | TypeScript の型チェックを実行します。 |
| `pnpm storybook` | Storybook を起動します。 |
| `pnpm build-storybook` | Storybook をビルドします。 |
