# Discord 音声読み上げBot

## 概要

このプロジェクトは、Discord上で音声読み上げを行うBotです。<br>
Voicevox Engineを使用して、テキストを音声に変換し、Discordのボイスチャットでユーザーの代わりに喋ります。

## 環境

*   Ruby 3.4.2
*   Docker
*   Docker Compose

## 必要なもの

*   Discord BotのToken
*   Discord BotのClient ID
*   Voicevox Engine (Dockerコンテナで起動)

## 使い方

1.  `config.yml`を作成し、必要な情報を設定します。`config.yml.sample`を参考にしてください。
2.  Docker Composeを使用して、BotとVoicevox Engineを起動します。

    ```bash
    docker-compose up --build
    ```

3.  DiscordでBotを招待し、`!ping`コマンドを試してみてください。`Pong!`と返信されれば、Botは正常に動作しています。

## 設定ファイル (config.yml)

```yaml
bot:
  token: 'YOUR_BOT_TOKEN'       # Discord BotのToken
  client_id: 'YOUR_CLIENT_ID'   # Discord BotのClient ID
  prefix: '!'                   # コマンドのプレフィックス
  status: '[TESTING] 読み上げBot' # Botのステータス

voicevox:
  max: 50                      # 最大文字数
  host: 'http://voicevox:50021' # Voicevox Engineのホスト
```

## コマンド

*   `!ping`: Botの死活確認を行います。
*   `!summon`: 読み上げBotをVCに参加させます。
*   `!bye`: 読み上げBotをVCから退出させます。
*   `!reconnect`: Discordが調子悪いときなどに、再接続します。

