# frozen_string_literal: true

# Command Module
module YourSaySan
  module Events
    # Return Ping-Pong.
    # This is a simple way to check whether a bot is dead or alive.
    module Ready
      extend Discordrb::EventContainer

      ready do |event|
        event.bot.game = YourSaySan::CONFIG.bot.status
        puts('Bot is Ready!')
      end
    end
  end
end
