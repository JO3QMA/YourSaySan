# frozen_string_literal: true

# Command Module
module YouSaySan
  module Commands
    # Return Ping-Pong.
    # This is a simple way to check whether a bot is dead or alive.
    module Ping
      extend Discordrb::Commands::CommandContainer
      command(:ping) do |event|
        event.respond('Pong!')
      end
    end
  end
end
