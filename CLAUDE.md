# Gaze - Terminal Image Viewer

## プロジェクト概要

GoによるKitty Graphicsプロトコル対応のターミナル画像ビューア。
ズーム・パン操作をキーボード/マウスで行える。
Kitty Graphics非対応環境（tmux等）ではSixelフォールバックを自動選択する。

## アーキテクチャ

Clean Architectureに従う。各層の依存方向: domain ← usecase ← adapter/infrastructure

### 層の責務

- **domain/**: エンティティ (ImageEntity, Viewport, Config)。外部依存なし
- **usecase/**: ポートインターフェース定義 + ユースケース実装
- **adapter/**: TUI (Bubbletea), Config (TOML), Renderer (Kitty, Sixel)
- **infrastructure/**: ファイルシステム画像読み込み

### 境界ルール

- domain は他のどの層にも依存してはならない
- usecase は domain のみに依存する
- adapter/infrastructure は usecase のポートインターフェースを実装する
- cmd/gaze/main.go でDI配線を行う

## 開発コマンド

```bash
make build    # バイナリビルド → bin/gaze
make test     # テスト実行 (race detector付き)
make lint     # golangci-lint実行
make ci       # lint + test + build
make clean    # ビルド成果物削除
```

## コーディングルール

### エラーハンドリング
- エラーは必ずハンドリングする (`_` で無視しない)
- エラーメッセージは `fmt.Errorf("doing X: %w", err)` でラップする

### テスト
- table-driven tests を使用する
- テストファイルは対象ファイルと同一パッケージに配置 (`*_test.go`)
- ドメイン層のテストカバレッジは95%以上を目標とする

### 命名規則
- Go標準の命名規則に従う (camelCase, 頭文字語は大文字)
- インターフェース名は動詞+erまたは役割名 (例: `RendererPort`, `ImageLoaderPort`)

### 依存関係
- 新しい外部依存を追加する際は必要性を検討する
- domain層に外部パッケージのimportを追加してはならない
