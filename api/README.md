# api

Gin で動く SSO 検証用 API サーバーです。sandbox_auth の `/me` を読み、返ってきた認証済みユーザーを PostgreSQL に保存します。ログアウト時は sandbox_auth の `/auth/logout` を Cookie 付きで中継します。

## 環境変数

`api/.env.template` を `api/.env.local` としてコピーし、DB と sandbox_auth API の URL を設定してください。

sandbox_auth の `ALLOWED_REDIRECT_URLS` には次の URL を含めてください。

```txt
http://localhost:3000/dashboard
```

## ローカル起動

先に migration を適用してください。

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
| `GET` | `/me` | sandbox_auth の `/me` を中継し、認証済みなら DB に保存します。 |
| `POST` | `/auth/logout` | sandbox_auth の `/auth/logout` を中継します。 |

## コマンド

| コマンド | 内容 |
| -------- | ---- |
| `mise run dev` | Air で API を開発起動します。 |
| `mise run build` | API のバイナリをビルドします。 |
| `mise run run` | API を `go run` で起動します。 |
| `mise run install` | Go module を整理します。 |
| `mise run migrate:up` | DB migration を適用します。 |
| `mise run migrate:down` | DB migration を戻します。 |
| `mise run format` | Go のコードを整形します。 |
| `mise run lint` | golangci-lint を実行します。 |
| `mise run lint:config-chk` | golangci-lint の設定を検証します。 |
| `mise run test` | Go のテストを実行します。 |
| `mise run ci` | build、format、lint、test を実行します。 |

## トラブルシューティング

### `DATABASE_URL is required` で終了する

`api/.env.local` に PostgreSQL の接続文字列を設定してください。

### `/me` が `failed to connect auth service` を返す

`AUTH_SERVER_URL` に Go API から見える sandbox_auth API の URL を設定してください。Docker Compose では `http://host.docker.internal:8080` を使います。

### sandbox_auth から dashboard に戻らない

sandbox_auth 側の `ALLOWED_REDIRECT_URLS` に、このアプリを開いている origin の `/dashboard` が含まれているか確認してください。通常は `http://localhost:3000/dashboard` です。
