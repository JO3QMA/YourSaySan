# frozen_string_literal: true

module YourSaySan
  module Commands
    # Pingコマンドモジュール
    module Ping
      extend Discordrb::Commands::CommandContainer

      command :ping do |event|
        event.respond 'Pong!'
      end
    end
  end
end
