# sandbox_app

sandbox_auth の SSO認証を前提とした簡易アプリケーション。
このアプリケーションは sandbox_auth のログイン画面へ遷移して、認証後に戻ってきた状態を確認する。

## 構成

| パス | 内容 |
| ---- | ---- |
| `web` | Next.js のフロントエンドです。 |
| `api` | Go と Gin の API です。 |
| `api/migrations` | goose の SQL migration です。 |
| `compose.yml` | 開発用の Docker Compose 定義です。 |

## 役割

このアプリケーションは認証ロジックを持ちません。ログイン画面、OAuth、パスキー、セッション発行、ログイン履歴は sandbox_auth の責務です。

このアプリケーションが担当するのは、sandbox_auth から返る認証済みユーザーを確認し、このアプリケーションで必要なユーザー情報を `app_accounts` に保存することです。

## 起動

Docker Compose で起動する場合は、次のコマンドを使います。

```txt
docker compose up --build
```

ローカルで個別に起動する場合は、各ディレクトリの README を参照してください。

| README | 内容 |
| ------ | ---- |
| `web/README.md` | Web の環境変数、起動方法、画面の説明です。 |
| `api/README.md` | API の環境変数、migration、エンドポイント、コマンドの説明です。 |

## sandbox_auth 側の設定

sandbox_auth の `ALLOWED_REDIRECT_URLS` には、このアプリケーションの `/dashboard` を許可してください。

```txt
ALLOWED_REDIRECT_URLS=http://localhost:3000/mypage,http://localhost:3000/dashboard
DEFAULT_REDIRECT_URL=http://localhost:3000/dashboard
```

別の URL でこのアプリケーションを開く場合は、その URL の `/dashboard` を許可してください。

## 注意

このアプリケーションの API と sandbox_auth API を同じホストで同じポートに起動することはできません。ポートを変更する場合は、`AUTH_SERVER_URL` と `NEXT_PUBLIC_API_BASE_URL` も合わせて変更してください。
