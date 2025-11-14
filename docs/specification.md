# Discord チャット読み上げBot 仕様書

## 目次

1. [概要](#概要)
2. [プロジェクト構造](#プロジェクト構造)
3. [起動フロー](#起動フロー)
4. [主要コンポーネント](#主要コンポーネント)
5. [機能詳細](#機能詳細)
6. [技術スタック](#技術スタック)
7. [設定ファイル](#設定ファイル)

---

## 概要

このプロジェクトは、Discord上でテキストチャットを音声で読み上げるBotです。VoiceVox Engineを使用してテキストを音声に変換し、Discordのボイスチャンネルで再生します。

### 主な特徴

- Discordのテキストチャンネルのメッセージを自動で音声読み上げ
- ユーザーごとの話者設定（VoiceVoxの話者を個別に設定可能）
- スラッシュコマンドによる操作
- Redisによる話者設定の永続化
- 自動再接続機能
- 自動切断機能（VCにBotのみが残った場合）

---

## プロジェクト構造

```
/app
├── run.rb                          # エントリーポイント
├── core/
│   ├── yoursay.rb                 # Botのメインクラス
│   ├── voicevox.rb                # VoiceVox API連携クラス
│   ├── speaker_manager.rb         # 話者設定管理クラス
│   └── modules/
│       ├── commands/              # スラッシュコマンドモジュール
│       │   ├── ping.rb            # 死活確認コマンド
│       │   ├── help.rb            # ヘルプコマンド
│       │   ├── invite.rb          # 招待リンク表示コマンド
│       │   ├── summon.rb          # VC参加コマンド
│       │   ├── bye.rb             # VC退出コマンド
│       │   ├── reconnect.rb       # 再接続コマンド
│       │   ├── stop.rb            # 読み上げ中断コマンド
│       │   ├── speaker.rb         # 話者設定コマンド
│       │   ├── speaker_list.rb    # 話者一覧表示コマンド
│       │   └── eval.rb            # コード実行コマンド（開発者用）
│       └── events/                # イベントハンドラモジュール
│           ├── ready.rb           # Bot起動時イベント
│           ├── read.rb            # メッセージ読み上げイベント
│           ├── auto_reconnect.rb  # 自動再接続イベント
│           └── auto_dissconnect.rb # 自動切断イベント
├── config.sample.yml              # 設定ファイルのサンプル
├── compose.yml                    # Docker Compose設定
├── Gemfile                        # Ruby依存関係
└── docs/                          # ドキュメント
```

---

## 起動フロー

### 1. エントリーポイント（`run.rb`）

```ruby
require 'bundler/setup'
require 'dotenv/load'
require_relative 'core/yoursay'
YourSaySan.run
```

1. Bundlerで依存関係を読み込み
2. `.env`ファイルから環境変数を読み込み
3. `YourSaySan`モジュールの`run`メソッドを実行

### 2. Bot初期化（`core/yoursay.rb`）

1. **設定ファイル読み込み**
   - `config.yml`をERBで処理（環境変数の埋め込み）
   - YAMLとして解析し、`Config`オブジェクトに設定

2. **Discord Botインスタンス作成**
   - `Discordrb::Bot`を初期化
   - Botトークン、クライアントIDを設定
   - `ignore_bots: true`でBotのメッセージを無視

3. **共有リソースの初期化**
   - `@text_channels`: 読み上げ対象のテキストチャンネルIDの配列
   - `@voicevox`: VoiceVox API連携クラスのインスタンス
   - `@speaker_manager`: 話者設定管理クラスのインスタンス

4. **モジュール読み込み（`load_modules`）**
   - `core/modules/commands/*.rb`を読み込み
   - `core/modules/events/*.rb`を読み込み
   - 各モジュールをBotに登録
   - スラッシュコマンドを登録

5. **Bot起動**
   - `BOT.run`でDiscordに接続

---

## 主要コンポーネント

### 1. YourSaySanモジュール（`core/yoursay.rb`）

Botのメインクラス。以下の責務を持ちます：

- Botインスタンスの管理
- 設定ファイルの読み込み
- モジュールの動的読み込みと登録
- 共有リソース（VoiceVox、SpeakerManager）の管理

#### 主要メソッド

- `load_modules`: コマンドとイベントモジュールを読み込み
- `register_slash_commands_from_modules`: スラッシュコマンドを登録
- `run`: Botを起動

### 2. VoiceVoxクラス（`core/voicevox.rb`）

VoiceVox Engine APIとの通信を担当します。

#### 主要メソッド

- `speak(text, speaker)`: テキストを音声データ（WAV形式）に変換
  - 内部で`voice_query`と`generate_voice`を呼び出し
- `get_speakers`: 利用可能な話者の一覧を取得
  - `/speakers`エンドポイントから取得
  - 話者IDと話者名のハッシュを返す

#### APIエンドポイント

- `POST /audio_query`: テキストから音声クエリを生成
- `POST /synthesis`: クエリから音声データを生成
- `GET /speakers`: 利用可能な話者一覧を取得

### 3. SpeakerManagerクラス（`core/speaker_manager.rb`）

ユーザーごとの話者設定をRedisで管理します。

#### 主要メソッド

- `get_speaker(user_id)`: ユーザーの話者設定を取得（デフォルト: 2）
- `set_speaker(user_id, speaker_id)`: ユーザーの話者設定を保存
- `get_available_speakers`: VoiceVoxから利用可能な話者一覧を取得
- `valid_speaker?(speaker_id)`: 話者IDが有効かチェック

#### Redisキー形式

- `speaker:{user_id}`: ユーザーの話者ID（整数）

---

## 機能詳細

### イベントハンドラ

#### 1. Readyイベント（`core/modules/events/ready.rb`）

BotがDiscordに接続完了した際に実行されます。

- Botのステータス（ゲーム表示）を設定
- コンソールに「Bot is Ready!」を出力

#### 2. Readイベント（`core/modules/events/read.rb`）

テキストチャンネルでメッセージが投稿された際に実行されます。

**処理フロー：**

1. メッセージが読み上げ対象チャンネル（`@text_channels`）かチェック
2. スラッシュコマンド（`/`で始まる）は読み上げない
3. 本文が空のメッセージ（画像のみなど）は読み上げない
4. メッセージを変換：
   - メンション（`<@user_id>`）を表示名（`@username`）に変換
   - URLを「URL省略」に置換
   - 最大文字数（デフォルト50文字）を超える場合は「以下略」を追加
5. ユーザーの話者設定を取得
6. VoiceVoxで音声データを生成
7. 一時ファイルにWAVデータを書き込み
8. ボイスチャンネルで音声を再生

#### 3. AutoReconnectイベント（`core/modules/events/auto_reconnect.rb`）

WebSocket接続が切断された際の処理を行います。

- 切断を検知してログ出力
- Discordrbの自動再接続機能に任せる
- 再接続完了時にログ出力

#### 4. AutoDissconnectイベント（`core/modules/events/auto_dissconnect.rb`）

VCからユーザーが退出した際に、Botのみが残った場合は自動で切断します。

- `voice_state_update`イベントを監視
- VCにBotのみが残っている場合、`voice_destroy`で切断

### スラッシュコマンド

すべてのコマンドは`Discordrb::EventContainer`を拡張したモジュールとして実装されています。

#### 1. `/ping`

**ファイル**: `core/modules/commands/ping.rb`

- **説明**: Botの死活確認
- **処理**: 「Pong!」を返信

#### 2. `/help`

**ファイル**: `core/modules/commands/help.rb`

- **説明**: 利用可能なコマンドの一覧または詳細を表示
- **オプション**: `command`（コマンド名、省略可）
- **処理**:
  - オプションなし: 全コマンドの一覧を表示
  - オプションあり: 指定コマンドの詳細を表示
  - 各コマンドモジュールの`COMMAND_INFO`定数から情報を自動取得

#### 3. `/invite`

**ファイル**: `core/modules/commands/invite.rb`

- **説明**: Botを他のサーバーに招待するためのURLを表示
- **処理**:
  - OAuth2認証URLを生成
  - 必要な権限（メッセージ送信、VC接続、音声再生、チャンネル閲覧）を含める
  - Embed形式で表示

#### 4. `/summon`

**ファイル**: `core/modules/commands/summon.rb`

- **説明**: BotをVCに参加させる
- **処理**:
  - コマンド実行者のVCに接続
  - 実行したテキストチャンネルを読み上げ対象に追加（`@text_channels`に追加）
  - 成功メッセージを返信

#### 5. `/bye`

**ファイル**: `core/modules/commands/bye.rb`

- **説明**: BotをVCから退出させる
- **処理**:
  - サーバーからVC接続を切断
  - 成功メッセージを返信

#### 6. `/reconnect`

**ファイル**: `core/modules/commands/reconnect.rb`

- **説明**: VC接続を再接続する
- **処理**:
  - 現在のVC接続を切断
  - コマンド実行者のVCに再接続
  - 成功メッセージを返信

#### 7. `/stop`

**ファイル**: `core/modules/commands/stop.rb`

- **説明**: 現在の読み上げを中断する
- **処理**:
  - `voice.stop_playing`で再生を停止
  - 成功メッセージを返信

#### 8. `/speaker`

**ファイル**: `core/modules/commands/speaker.rb`

- **説明**: ユーザーの話者を設定する
- **オプション**: `speaker_id`（話者ID、必須）
- **処理**:
  1. VoiceVoxから利用可能な話者一覧を取得
  2. 指定された話者IDが有効かチェック
  3. Redisに話者設定を保存
  4. 設定した話者名を返信

#### 9. `/speaker_list`

**ファイル**: `core/modules/commands/speaker_list.rb`

- **説明**: 利用可能な話者の一覧を表示
- **オプション**: `page`（ページ番号、省略可、デフォルト: 1）
- **処理**:
  1. VoiceVoxから利用可能な話者一覧を取得
  2. 現在のユーザーの話者設定を取得
  3. ページネーション（1ページ20件）で表示
  4. 現在の設定を強調表示（▶マーク）
  5. Embed形式で表示

#### 10. `/eval`

**ファイル**: `core/modules/commands/eval.rb`

- **説明**: コードを実行する（開発者用）
- **オプション**: `code`（実行するコード、必須）
- **権限**: Botオーナーのみ実行可能
- **処理**:
  1. オーナーIDをチェック
  2. コードを`eval`で実行
  3. 実行結果またはエラーメッセージを返信

---

## 技術スタック

### 言語・フレームワーク

- **Ruby**: 3.4.2以上
- **Discordrb**: Discord Bot APIクライアント
- **Config**: 設定ファイル管理

### 主要ライブラリ

- `discordrb`: Discord Bot API
- `opus-ruby`: Opus音声コーデック
- `wavefile`: WAVファイル処理
- `redis`: Redisクライアント
- `dotenv`: 環境変数管理
- `config`: 設定ファイル管理

### 外部サービス

- **VoiceVox Engine**: テキスト音声合成エンジン（Dockerコンテナ）
- **Redis**: 話者設定の永続化（Dockerコンテナ）

### インフラ

- **Docker**: コンテナ化
- **Docker Compose**: マルチコンテナ管理

---

## 設定ファイル

### config.yml

ERBテンプレートとして実装されており、環境変数を埋め込めます。

```yaml
bot:
  token: <%= ENV['DISCORD_BOT_TOKEN'] %>       # Discord Botトークン
  client_id: <%= ENV['DISCORD_CLIENT_ID'] %>   # Discord BotクライアントID
  prefix: '!'                                  # コマンドプレフィックス（未使用）
  status: '[TESTING] 読み上げBot'              # Botのステータス
  owner: <%= ENV['DISCORD_OWNER_ID'] || 123456789012345678 %> # BotオーナーID

voicevox:
  max: 50                                      # 最大読み上げ文字数
  host: <%= ENV['VOICEVOX_HOST'] || 'http://voicevox:50021' %> # VoiceVox Engineのホスト

redis:
  host: <%= ENV['REDIS_HOST'] || 'redis' %>    # Redisホスト
  port: <%= ENV['REDIS_PORT'] || 6379 %>       # Redisポート
  db: <%= ENV['REDIS_DB'] || 0 %>              # Redisデータベース番号
```

### 環境変数

- `DISCORD_BOT_TOKEN`: Discord Botトークン（必須）
- `DISCORD_CLIENT_ID`: Discord BotクライアントID（必須）
- `DISCORD_OWNER_ID`: BotオーナーID（任意）
- `VOICEVOX_HOST`: VoiceVox Engineのホスト（デフォルト: `http://voicevox:50021`）
- `REDIS_HOST`: Redisホスト（デフォルト: `redis`）
- `REDIS_PORT`: Redisポート（デフォルト: `6379`）
- `REDIS_DB`: Redisデータベース番号（デフォルト: `0`）

---

## データフロー

### メッセージ読み上げの流れ

1. **メッセージ投稿**
   - ユーザーがテキストチャンネルにメッセージを投稿

2. **イベント発火**
   - `Read`イベントハンドラが発火
   - 読み上げ対象チャンネルかチェック

3. **メッセージ変換**
   - メンションを表示名に変換
   - URLを「URL省略」に置換
   - 文字数制限を適用

4. **話者設定取得**
   - `SpeakerManager`からユーザーの話者IDを取得
   - Redisから取得、存在しない場合はデフォルト（2）を使用

5. **音声生成**
   - `VoiceVox.speak`を呼び出し
   - VoiceVox APIにリクエスト送信
   - WAV形式の音声データを取得

6. **音声再生**
   - 一時ファイルにWAVデータを書き込み
   - Discordのボイスチャンネルで再生

### 話者設定の流れ

1. **設定コマンド実行**
   - ユーザーが`/speaker`コマンドを実行

2. **話者ID検証**
   - VoiceVox APIから利用可能な話者一覧を取得
   - 指定された話者IDが有効かチェック

3. **設定保存**
   - Redisに`speaker:{user_id}`キーで保存

4. **読み上げ時の使用**
   - メッセージ読み上げ時に`SpeakerManager.get_speaker`で取得
   - VoiceVox APIに渡して音声生成

---

## エラーハンドリング

### VoiceVox API

- タイムアウト: 3秒（接続）、10秒（読み込み）
- エラー時は`nil`を返し、ログに記録

### Redis

- 接続エラー時は`SpeakerManager`が`nil`になる可能性
- 各コマンドで`nil`チェックを実施

### Discord接続

- WebSocket切断時は自動再接続
- エラー発生時はスタックトレースを出力

---

## モジュール設計

### モジュールの登録方法

1. **コマンドモジュール**
   - `YourSaySan::Commands`モジュール内に定義
   - `Discordrb::EventContainer`を拡張
   - `register_slash_command`メソッドでスラッシュコマンドを登録
   - `COMMAND_INFO`定数でコマンド情報を定義

2. **イベントモジュール**
   - `YourSaySan::Events`モジュール内に定義
   - `Discordrb::EventContainer`を拡張
   - `setup`メソッドで初期化（必要に応じて）

### モジュールの動的読み込み

- `Dir['./core/modules/commands/*.rb'].sort`でファイルを順次読み込み
- `require`で読み込み後、定数を取得してBotに登録

---

## まとめ

このBotは、モジュール設計により機能を拡張しやすく、VoiceVoxとRedisを活用してユーザーごとの話者設定を実現しています。スラッシュコマンドによる直感的な操作と、自動再接続・自動切断などの運用面での配慮も実装されています。

