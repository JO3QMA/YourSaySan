# frozen_string_literal: true

# Command Module
module YouSaySan
  module Events
    # Return Ping-Pong.
    # This is a simple way to check whether a bot is dead or alive.
    module Read
      extend Discordrb::EventContainer
      message(start_with: not!(BOT::CONFIG.bot.prefix), in: @text_channels) do |event|
        if event.voice
          next if event.message.content == '' # 画像など本文がない投稿を弾く

          message = event.message.content
          message = message.gsub(/<@([0-9]{18})>/) { "@#{bot.member(event.server, Regexp.last_match(1)).display_name}" }
          message = message.gsub(URI::DEFAULT_PARSER.make_regexp(%w[http https]), 'URL省略')
          message = "#{message[0, config.voicevox.max - 1]} 以下略" if message.size >= config.voicevox.max
          say(event.voice, message)
        else
          @text_channels.delete(event.channel.id)
          nil
        end
      end

      def say(voice_chat, message)
        Tempfile.create do |tempfile|
          sound = @voicevox.speak(message)
          tempfile.write(sound)
          voice_chat.play_file(tempfile.path)
        end
      end
    end
  end
end
