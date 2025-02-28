# frozen_string_literal: true

# Command Module
module YouSaySan
  module Events
    # Return Ping-Pong.
    # This is a simple way to check whether a bot is dead or alive.
    module Ready
      extend Discordrb::EventContainer
      ready do |event|
        event.bot.game = YouSaySan::CONFIG.bot.status
        puts('Bot is Ready!')
      end
    end
  end
end
