# frozen_string_literal: true

module YourSaySan
  module Commands
    # Reconnectコマンドモジュール
    module Reconnect
      extend Discordrb::EventContainer

      # コマンド情報を定義
      COMMAND_INFO = {
        desc: '読み上げBotを再接続します。読み上げされない場合に有効かもしれません。',
        usage: '引数は必要ありません。'
      }.freeze

      def self.register_slash_command(bot)
        bot.register_application_command(:reconnect, COMMAND_INFO[:desc])
      end

      application_command :reconnect do |event|
        if YourSaySan::BOT.voice(event.server)
          YourSaySan::BOT.voice(event.server).destroy_connection
          YourSaySan::BOT.voice_connect(event.user.voice_channel)
          event.respond(content: '再接続しました。', ephemeral: true)
        else
          event.respond(content: 'ボイスチャットに接続していません。', ephemeral: true)
        end
      end
    end
  end
end
