# frozen_string_literal: true

# Command Module
module YourSaySan
  module Commands
    # Command to interrupt someone speaking
    module Stop
      extend Discordrb::Commands::CommandContainer
      command(:stop) do |event|
        if event.voice
          event.voice.stop_playing if event.voice.playing?
          nil
        else
          event.respond('VCに参加していません。')
        end
      end
    end
  end
end
