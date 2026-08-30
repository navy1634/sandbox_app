# api

Gin で動くOIDCクライアントAPIサーバーです。sandbox_authのOIDC callbackで認可コードをTokenへ交換し、IDトークンを検証してアプリ用Cookieセッションを発行します。検証済みの `sub` をPostgreSQLのアプリユーザーへ紐づけます。

## 環境変数

`api/.env.template` を `api/.env.local` としてコピーし、DBとOIDCの設定を指定してください。

`OIDC_ISSUER_URL` はブラウザから到達できる公開Issuer、`OIDC_INTERNAL_URL` はAPIコンテナからDiscovery・Token・JWKSへ到達するURLを指定します。Docker ComposeではIssuerに `http://localhost:8080`、内部URLに `http://host.docker.internal:8080` を使います。

```txt
OIDC_ISSUER_URL=http://localhost:8080
OIDC_INTERNAL_URL=http://host.docker.internal:8080
OIDC_CLIENT_ID=external-app
OIDC_CLIENT_SECRET=replace-with-a-random-secret
OIDC_REDIRECT_URL=http://localhost:8081/auth/callback
APP_AUTH_SECRET=replace-with-at-least-32-bytes
```

sandbox_auth側の `OIDC_CLIENTS` には、client ID、client secret、callback URLを完全一致で登録してください。

```txt
OIDC_CLIENTS=[{"client_id":"external-app","client_secret":"replace-with-a-random-secret","redirect_uris":["http://localhost:8081/auth/callback"]}]
```

## ローカル起動

先にmigrationを適用してください。

```txt
mise run migrate:up
```

```txt
mise run dev
```

通常の `go run` で起動する場合は、次のコマンドを使います。

```txt
mise run run
```

## エンドポイント

| メソッド | パス | 内容 |
| -------- | ---- | ---- |
| `GET` | `/health` | ヘルスチェックを返します。 |
| `GET` | `/me` | アプリ用Cookieセッションを検証し、認証済みユーザーをDBへ保存します。 |
| `GET` | `/auth/login` | OIDC認可要求を開始します。 |
| `GET` | `/auth/callback` | OIDC callbackを検証し、アプリ用セッションを発行します。 |
| `POST` | `/auth/logout` | アプリ用Cookieセッションを削除します。 |

## コマンド

| コマンド | 内容 |
| -------- | ---- |
| `mise run dev` | AirでAPIを開発起動します。 |
| `mise run build` | APIのバイナリをビルドします。 |
| `mise run run` | APIを `go run` で起動します。 |
| `mise run install` | Go moduleを整理します。 |
| `mise run migrate:up` | Go migrationを適用します。 |
| `mise run migrate:down` | Go migrationを戻します。 |
| `mise run format` | Goのコードを整形します。 |
| `mise run lint` | golangci-lintを実行します。 |
| `mise run lint:config-chk` | golangci-lintの設定を検証します。 |
| `mise run test` | Goのテストを実行します。 |
| `mise run ci` | build、format、lint、testを実行します。 |

## トラブルシューティング

### `DATABASE_URL is required` で終了する

`api/.env.local` にPostgreSQLの接続文字列を設定してください。

### `/auth/login` が `failed to prepare OIDC authorization` を返す

`OIDC_INTERNAL_URL` からsandbox_authの `/.well-known/openid-configuration` へ到達できるか確認してください。DiscoveryのIssuerは `OIDC_ISSUER_URL` と一致している必要があります。

### callbackで `OIDC token validation failed` になる

sandbox_auth側の `OIDC_CLIENTS` に、`OIDC_CLIENT_ID`、`OIDC_CLIENT_SECRET`、`OIDC_REDIRECT_URL` と完全一致するクライアント設定があるか確認してください。callback URLは認可要求とToken交換の両方で `redirect_uri` として使われます。
