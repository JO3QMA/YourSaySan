# frozen_string_literal: true

# Command Module
module YouSaySan
  module Events
    module Read
      extend Discordrb::EventContainer

      def self.setup(bot, text_channels, voicevox, config)
        bot.message(start_with: not!(bot.prefix), in: text_channels) do |event|
          if event.voice
            next if event.message.content == '' # 画像など本文がない投稿を弾く

            message = event.message.content
            message = message.gsub(/<@([0-9]{18})>/) do
              "@#{bot.member(event.server, Regexp.last_match(1)).display_name}"
            end
            message = message.gsub(URI::DEFAULT_PARSER.make_regexp(%w[http https]), 'URL省略')
            message = "#{message[0, config.voicevox.max - 1]} 以下略" if message.size >= config.voicevox.max
            say(event.voice, message, voicevox)
          else
            text_channels.delete(event.channel.id)
            nil
          end
        end
      end

      def say(voice_chat, message, voicevox)
        Tempfile.create do |tempfile|
          sound = voicevox.speak(message)
          tempfile.write(sound)
          voice_chat.play_file(tempfile.path)
        end
      end
      
      setup(YouSaySan::BOT, YouSaySan.text_channels, YouSaySan.voicevox, YouSaySan::CONFIG)
    end
  end
end

