# frozen_string_literal: true

module YourSaySan
  module Commands
    # Pingコマンドモジュール
    module Ping
      extend Discordrb::EventContainer

      def self.register_slash_command(bot)
        bot.register_application_command(:ping, 'Pongを返します')
      end

      application_command :ping do |event|
        event.respond(content: 'Pong!', ephemeral: true)
      end
    end
  end
end
