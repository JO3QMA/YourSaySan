# frozen_string_literal: true

# Command Module
module YourSaySan
  module Commands
    # Command to interrupt someone speaking
    module Stop
      extend Discordrb::EventContainer
      
      def self.register_slash_command(bot)
        bot.register_application_command(:stop, '読み上げを中断します')
      end
      
      application_command(:stop) do |event|
        if event.voice
          event.voice.stop_playing if event.voice.playing?
          event.respond(content: '読み上げを停止しました。', ephemeral: true)
        else
          event.respond(content: 'VCに参加していません。', ephemeral: true)
        end
      end
    end
  end
end
