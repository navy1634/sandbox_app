# sandbox_app

sandbox_authをOIDC Providerとして利用する簡易アプリケーションです。
このアプリケーションAPIがOIDC callbackを受け、IDトークンを検証してアプリ用セッションを発行します。

## 構成

| パス | 内容 |
| ---- | ---- |
| `web` | Next.js のフロントエンドです。 |
| `api` | Go と Gin の API です。 |
| `api/migrations` | goose の SQL migration です。 |
| `compose.yml` | 開発用の Docker Compose 定義です。 |

## 役割

このアプリケーションはOIDCクライアントとして認証します。Google OAuth、パスキー、sandbox_authのログイン画面はsandbox_authの責務です。

このアプリケーションが担当するのは、OIDC callbackで受け取ったIDトークンを検証し、`sub`をキーにユーザー情報を `app_accounts` へ保存することです。

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

## OIDC設定

sandbox_authの `OIDC_CLIENTS` に、このアプリケーションのcallback URLを完全一致で登録してください。

```txt
OIDC_ISSUER_URL=http://localhost:8080
OIDC_CLIENTS=[{"client_id":"external-app","client_secret":"replace-with-a-random-secret","redirect_uris":["http://localhost:8081/auth/callback"]}]
OIDC_REDIRECT_URL=http://localhost:8081/auth/callback
```

アプリAPIの `OIDC_ISSUER_URL` はブラウザから到達できるIssuer、`OIDC_INTERNAL_URL` はAPIコンテナからDiscovery・Token・JWKSへ到達するURLです。`APP_AUTH_SECRET` はアプリ専用の32バイト以上の秘密値にしてください。

## 注意

ローカルComposeではsandbox_auth APIを8080、sandbox_app APIを8081で公開します。ポートを変更する場合は、OIDCの `OIDC_REDIRECT_URL`、sandbox_auth側の `redirect_uris`、Webの `NEXT_PUBLIC_API_BASE_URL` を同じcallback構成に合わせてください。

```txt
192.168.0.100 registry.local
192.168.0.242 auth.sandbox.navy1634.com
192.168.0.242 app.sandbox.navy1634.com
```
