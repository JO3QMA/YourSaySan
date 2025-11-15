# Opusエンコーダー移行ガイド

## 概要

このプロジェクトでは、音声エンコーディングに2つの実装をサポートしています：

1. **DCAエンコーダー**（デフォルト）: `github.com/jonas747/dca`を使用、FFmpegに依存
2. **Opusエンコーダー**（新実装）: `github.com/hraban/opus`と`github.com/go-audio/wav`を使用、Pure Go実装

## 切り替え方法

環境変数`USE_PION_OPUS=true`を設定することで、新しいOpusエンコーダーを使用できます。

```bash
# Opusエンコーダーを使用
export USE_PION_OPUS=true
./yoursay-bot

# またはDocker Composeの場合
USE_PION_OPUS=true docker-compose up
```

## 違い

### DCAエンコーダー（デフォルト）

- **依存**: FFmpeg（システムにインストール必要）
- **処理方法**: 一時ファイルを使用
- **メリット**: 実績のある実装、安定性が高い
- **デメリット**: FFmpeg依存、ディスクI/O発生

### Opusエンコーダー（新実装）

- **依存**: libopus（CGOビルド時）、Pure Go WAVパーサー
- **処理方法**: メモリ上で直接処理
- **メリット**: 一時ファイル不要、ディスクI/O削減、メンテナンスが活発
- **デメリット**: CGOビルドが必要、opus-devが必要

## ビルド要件

Opusエンコーダーを使用する場合、ビルド時に以下が必要です：

- `CGO_ENABLED=1`（デフォルトで有効）
- `opus-dev`パッケージ（Alpine Linuxの場合）

Dockerfileには既に`opus-dev`が含まれています。

## エラー処理

Opusエンコーダーは以下のエラーを返す可能性があります：

- `ErrInvalidWAVFormat`: WAVファイルフォーマットが無効
- `ErrUnsupportedSampleRate`: サンプルレートが48kHz以外
- `ErrUnsupportedChannels`: チャンネル数が1または2以外
- `ErrOpusEncodeFailed`: Opusエンコードに失敗

## パフォーマンス

Opusエンコーダーは一時ファイルを使用しないため、ディスクI/Oが削減されます。ただし、CGOオーバーヘッドがあるため、パフォーマンスは環境によって異なります。

## ロールバック

問題が発生した場合、環境変数を削除または`USE_PION_OPUS=false`に設定することで、DCAエンコーダーに戻すことができます。

```bash
# DCAエンコーダーに戻す
unset USE_PION_OPUS
# または
export USE_PION_OPUS=false
```

## 今後の計画

将来的には、Opusエンコーダーをデフォルトにし、DCAエンコーダーを削除する予定です。

