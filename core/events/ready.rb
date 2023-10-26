# frozen_string_literal: true

# Command Module
module YouSaySan
  module Events
    # Return Ping-Pong.
    # This is a simple way to check whether a bot is dead or alive.
    module Ready
      extend Discordrb::Commands::EventContainer
      ready do |event|
        event.bot.game = BOT::CONFIG.bot.status
      end
    end
  end
end
