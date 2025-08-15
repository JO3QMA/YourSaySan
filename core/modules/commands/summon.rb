# frozen_string_literal: true

module YourSaySan
  module Commands
    # Summonコマンドモジュール
    module Summon
      extend Discordrb::EventContainer

      def self.register_slash_command(bot)
        bot.register_application_command(:summon, '読み上げBotを呼び出します')
      end

      application_command :summon do |event|
        if event.user.voice_channel
          YourSaySan::BOT.voice_connect(event.user.voice_channel)
          YourSaySan.text_channels << event.channel.id unless YourSaySan.text_channels.include?(event.channel.id)
          event.respond(content: 'ボイスチャットに参加しました。', ephemeral: true)
        else
          event.respond(content: 'ボイスチャットに参加してください。', ephemeral: true)
        end
      end
    end
  end
end
