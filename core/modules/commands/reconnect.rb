# frozen_string_literal: true

module YourSaySan
  module Commands
    # Reconnectコマンドモジュール
    module Reconnect
      extend Discordrb::EventContainer

      def self.register_slash_command(bot)
        bot.register_application_command(:reconnect, '読み上げBotを再接続します')
      end

      application_command :reconnect do |event|
        if event.user.voice_channel
          YourSaySan::BOT.voice_destroy(event.server)
          YourSaySan::BOT.voice_connect(event.user.voice_channel)
          event.respond(content: '再接続しました。', ephemeral: true)
        else
          event.respond(content: 'ボイスチャットに参加してください。', ephemeral: true)
        end
      end
    end
  end
end


