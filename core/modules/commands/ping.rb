# frozen_string_literal: true

module YourSaySan
  module Commands
    # Pingコマンドモジュール
    module Ping
      extend Discordrb::EventContainer

      # コマンド情報を定義
      COMMAND_INFO = {
        desc: '`Pong`を返します。',
        usage: '引数は必要ありません。'
      }.freeze

      def self.register_slash_command(bot)
        bot.register_application_command(:ping, COMMAND_INFO[:desc])
      end

      application_command :ping do |event|
        event.respond(content: 'Pong!', ephemeral: true)
      end
    end
  end
end
