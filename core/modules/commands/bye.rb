# frozen_string_literal: true

module YourSaySan
  module Commands
    # Byeコマンドモジュール
    module Bye
      extend Discordrb::EventContainer

      def self.register_slash_command(bot)
        bot.register_application_command(:bye, '読み上げBotを切断します')
      end

      application_command :bye do |event|
        if event.user.voice_channel
          YourSaySan::BOT.voice_destroy(event.server)
          event.respond(content: 'ボイスチャットから切断しました。', ephemeral: true)
        else
          event.respond(content: 'ボイスチャットに参加していません。', ephemeral: true)
        end
      end
    end
  end
end
