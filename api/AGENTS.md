# AGENTS.md

## 基本方針

- 回答は日本語の敬語で書いてください。
- 変更前に対象ファイルを読み、既存の書き方に合わせてください。
- 不要な整形、コメント削除、命名変更は避けてください。
- Go の実行、ビルド、テスト、migration は `api/` で `mise` の task を優先して使ってください。
- Python は使わないでください。必要な場合でも `uv` 経由で実行してください。

## API の概要

- `api/` は Go 1.26.2、Gin、Ent、PostgreSQL で動く API サーバーです。
- 起動入口は `src/main.go` です。設定を読み、DB に接続し、`router.RegisterRoutes` でルートを登録します。
- ルート定義は `src/router/auth.go` にあります。CORS と request transaction middleware を全体に適用します。
- 認証は Google OAuth、アプリ用セッション Cookie、パスキーを組み合わせています。
- 生成された Ent コードは `src/ent/` にあります。通常のレビューや修正では、まず schema、handler、repository、infrastructure を確認してください。

## 主要ディレクトリ

- `src/config` は環境変数を読みます。
- `src/router` は HTTP ルートと middleware を登録します。
- `src/handler` は Gin の handler を置きます。
- `src/domain` は API のドメイン型と入力検証を置きます。
- `src/repository` は永続化の interface を置きます。
- `src/infrastructure/auth` は Google OAuth と WebAuthn の実装を置きます。
- `src/infrastructure/session` は Cookie に入れるセッション値の署名と検証を担当します。
- `src/infrastructure/persistence` は Ent を使った repository 実装を置きます。
- `src/infrastructure/database` は DB 接続と request transaction middleware を置きます。
- `migrations` は goose の SQL migration を置きます。

## DB と migration

- `accounts` はプロフィールを持ちます。
- Ent schema と migration の両方を確認してください。片方だけを見ると制約や型の理解を誤る可能性があります。

## 作業時の確認ポイント

- DB 更新を触るときは、request transaction middleware の挙動を確認してください。HTTP 400 以上または `c.Errors` がある場合は commit されません。
- セキュリティレビューでは、Cookie の属性、セッション期限、CSRF 境界、WebAuthn ceremony session の一回消費、エラー内容の漏えいを重点的に確認してください。

## スクリプト

- 開発サーバーは `mise run dev` で起動します。
- ビルドは `mise run build` で確認します。
- テストは `mise run test` で実行します。
- lint は `mise run lint` で実行します。
- migration は `mise run migrate:up` と `mise run migrate:down` を使います。
