# frozen_string_literal: true

# Command Module
module YouSaySan
  module Events
    # Return Ping-Pong.
    # This is a simple way to check whether a bot is dead or alive.
    module VoiceState
      extend Discordrb::Commands::EventContainer
      voice_state_update do |event|
        if !event.channel && event.bot.voice(event.server) && !event.voice(event.server).channel.users.map(&:current_bot?).include?(false)
          tc = (event.server.text_channels.map(&:id) & @text_channels)[0]
          @text_channels.delete(tc)
          event.bot.voice_destroy(event.server)
          event.send_message(tc, 'See you!')
        end
      end
    end
  end
end
